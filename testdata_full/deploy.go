package main

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/backup"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	iamtypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/aws-sdk-go-v2/service/s3tables"
	s3tablesTypes "github.com/aws/aws-sdk-go-v2/service/s3tables/types"
	"github.com/aws/aws-sdk-go-v2/service/s3vectors"
	s3vectorsTypes "github.com/aws/aws-sdk-go-v2/service/s3vectors/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/fatih/color"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

const (
	region = "us-east-1"
)

type Options struct {
	Profile    string
	Stage      string
	RetainMode bool
}

type DeployStackService struct {
	Options         Options
	CfnPjPrefix     string
	CfnStackName    string
	AccountID       string
	ProfileOption   string
	Ctx             context.Context
	CfnClient       *cloudformation.Client
	S3Client        *s3.Client
	S3TablesClient  *s3tables.Client
	S3VectorsClient *s3vectors.Client
	IamClient       *iam.Client
	EcrClient       *ecr.Client
	StsClient       *sts.Client
	BackupClient    *backup.Client
}

// This script allows you to deploy the stack for delstack testing.
// Due to quota limitations, only up to [5 test stacks] can be created by this script at the same time.
func main() {
	ctx := context.Background()
	options := parseArgs()

	if options.Stage == "" {
		// Generate a random number using current time as seed
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		randomNum := r.Intn(10000)
		options.Stage = fmt.Sprintf("delstack-test-%04d", randomNum)
	}

	service := NewDeployStackService(ctx, options)

	// Initialize AWS clients
	if err := service.initAWSClients(); err != nil {
		color.Red("Failed to initialize AWS clients: %v", err)
		os.Exit(1)
	}

	// Build Docker image
	if err := service.buildImage(); err != nil {
		color.Red("Failed to build Docker image: %v", err)
		os.Exit(1)
	}

	// Deploy using CDK
	if err := service.cdkDeploy(); err != nil {
		color.Red("Failed to deploy: %v", err)
		os.Exit(1)
	}

	// Attach user to group
	color.Green("=== attach_user_to_group ===")
	if err := service.attachUserToGroup(service.CfnStackName); err != nil {
		color.Red("Failed to attach user to group: %v", err)
		os.Exit(1)
	}

	// Upload objects to S3
	color.Green("=== object_upload ===")
	if err := service.objectUpload(service.CfnStackName); err != nil {
		color.Red("Failed to upload objects: %v", err)
		os.Exit(1)
	}

	// Upload tables to table bucket
	color.Green("=== tables_upload_to_table_bucket ===")
	if err := service.tablesUploadToTableBucket(service.CfnStackName); err != nil {
		color.Red("Failed to upload tables to table bucket: %v", err)
		os.Exit(1)
	}

	// Upload indexes to vector bucket
	color.Green("=== indexes_upload_to_vector_bucket ===")
	if err := service.indexesUploadToVectorBucket(service.CfnStackName); err != nil {
		color.Red("Failed to upload indexes to vector bucket: %v", err)
		os.Exit(1)
	}

	// Build and upload Docker images
	color.Green("=== build_upload ===")
	if err := service.buildUpload(service.CfnStackName); err != nil {
		color.Red("Failed to build and upload: %v", err)
		os.Exit(1)
	}

	// Start backup
	color.Green("=== start_backup ===")
	if err := service.startBackup(service.CfnStackName); err != nil {
		color.Red("Failed to start backup: %v", err)
		os.Exit(1)
	}

	// Attach policy to role
	// The following function is no longer needed as the IAM role no longer fails on normal deletion, but it is left in place just in case.
	color.Green("=== attach_policy_to_role ===")
	if err := service.attachPolicyToRole(service.CfnStackName); err != nil {
		color.Red("Failed to attach policy to role: %v", err)
		os.Exit(1)
	}
}

func parseArgs() Options {
	options := Options{}

	for i := 1; i < len(os.Args); i++ {
		if os.Args[i] == "-p" && i+1 < len(os.Args) {
			options.Profile = os.Args[i+1]
			i++
		} else if os.Args[i] == "-s" && i+1 < len(os.Args) {
			options.Stage = os.Args[i+1]
			i++
		} else if os.Args[i] == "-r" {
			options.RetainMode = true
		}
	}

	return options
}

func NewDeployStackService(ctx context.Context, options Options) *DeployStackService {
	cfnPjPrefix := options.Stage
	cfnStackName := fmt.Sprintf("%s-Test-Full", cfnPjPrefix)

	profileOption := ""
	if options.Profile != "" {
		profileOption = fmt.Sprintf("--profile %s --region %s", options.Profile, region)
	}

	return &DeployStackService{
		Options:       options,
		CfnPjPrefix:   cfnPjPrefix,
		CfnStackName:  cfnStackName,
		ProfileOption: profileOption,
		Ctx:           ctx,
	}
}

func (s *DeployStackService) initAWSClients() error {
	var cfg aws.Config
	var err error

	if s.Options.Profile != "" {
		cfg, err = config.LoadDefaultConfig(s.Ctx,
			config.WithRegion(region),
			config.WithSharedConfigProfile(s.Options.Profile),
		)
	} else {
		cfg, err = config.LoadDefaultConfig(s.Ctx,
			config.WithRegion(region),
		)
	}

	if err != nil {
		return fmt.Errorf("failed to load AWS config: %v", err)
	}

	s.CfnClient = cloudformation.NewFromConfig(cfg)
	s.S3Client = s3.NewFromConfig(cfg)
	s.S3TablesClient = s3tables.NewFromConfig(cfg)
	s.S3VectorsClient = s3vectors.NewFromConfig(cfg)
	s.IamClient = iam.NewFromConfig(cfg)
	s.EcrClient = ecr.NewFromConfig(cfg)
	s.StsClient = sts.NewFromConfig(cfg)
	s.BackupClient = backup.NewFromConfig(cfg)

	// Get account ID
	stsOutput, err := s.StsClient.GetCallerIdentity(s.Ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return fmt.Errorf("failed to get AWS account ID: %v", err)
	}
	s.AccountID = *stsOutput.Account

	return nil
}

func (s *DeployStackService) runCommand(command string) error {
	cmd := exec.Command("bash", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (s *DeployStackService) buildImage() error {
	// Login to ECR using AWS SDK
	if err := s.loginToECR(); err != nil {
		return fmt.Errorf("failed to login to ECR: %v", err)
	}

	// Build Docker image
	buildCmd := "docker build -t delstack-test ."
	if err := s.runCommand(buildCmd); err != nil {
		return fmt.Errorf("failed to build Docker image: %v", err)
	}

	return nil
}

func (s *DeployStackService) cdkDeploy() error {
	// Set region
	os.Setenv("CDK_DEFAULT_REGION", region)

	// Get the account ID
	os.Setenv("CDK_DEFAULT_ACCOUNT", s.AccountID)

	// Deploy with CDK (from the cdk directory)
	profileOption := ""
	if s.Options.Profile != "" {
		profileOption = fmt.Sprintf("--profile %s", s.Options.Profile)
	}
	deployCmd := fmt.Sprintf("cd cdk && npx cdk deploy --all -c PJ_PREFIX=%s -c RETAIN_MODE=%s --require-approval never %s", s.CfnPjPrefix, strconv.FormatBool(s.Options.RetainMode), profileOption)
	if err := s.runCommand(deployCmd); err != nil {
		return fmt.Errorf("failed to deploy with CDK: %v", err)
	}

	return nil
}

// loginToECR - Get ECR authorization token and login to Docker using AWS SDK
func (s *DeployStackService) loginToECR() error {
	// Call ECR GetAuthorizationToken API
	authOutput, err := s.EcrClient.GetAuthorizationToken(s.Ctx, &ecr.GetAuthorizationTokenInput{})
	if err != nil {
		return fmt.Errorf("failed to get ECR authorization token: %v", err)
	}

	if len(authOutput.AuthorizationData) == 0 {
		return fmt.Errorf("ECR authorization data is empty")
	}

	// Decode Base64 encoded token
	authToken := *authOutput.AuthorizationData[0].AuthorizationToken
	decodedToken, err := base64.StdEncoding.DecodeString(authToken)
	if err != nil {
		return fmt.Errorf("failed to decode authorization token: %v", err)
	}

	// Split credentials in username:password format
	creds := strings.SplitN(string(decodedToken), ":", 2)
	if len(creds) != 2 {
		return fmt.Errorf("invalid credential format")
	}

	// Get registry URL from ECR endpoint
	ecrEndpoint := *authOutput.AuthorizationData[0].ProxyEndpoint

	// Login using Docker API
	cmd := exec.Command("docker", "login", "--username", creds[0], "--password-stdin", ecrEndpoint)
	cmd.Stdin = strings.NewReader(creds[1])
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to login to Docker: %v", err)
	}

	color.Green("Successfully logged in to ECR: %s", ecrEndpoint)
	return nil
}

// The following function is no longer needed as the IAM role no longer fails on normal deletion, but it is left in place just in case.
func (s *DeployStackService) attachPolicyToRole(stackName string) error {
	// Get resources in the stack
	resources, nestedStackNames, err := s.getStackResources(stackName)
	if err != nil {
		return err
	}

	// Process nested stacks in parallel
	var wg sync.WaitGroup
	errorChan := make(chan error, len(nestedStackNames))

	for _, nestedStackName := range nestedStackNames {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()
			if nestErr := s.attachPolicyToRole(name); nestErr != nil {
				errorChan <- nestErr
			}
		}(nestedStackName)
	}

	wg.Wait()
	close(errorChan)

	for err := range errorChan {
		if err != nil {
			return err
		}
	}

	policyName := "DelstackTestPolicy"

	// Create policy if it doesn't exist
	policyDoc, err := os.ReadFile("./policy_document.json")
	if err != nil {
		return fmt.Errorf("failed to read policy document: %v", err)
	}

	_, err = s.IamClient.CreatePolicy(s.Ctx, &iam.CreatePolicyInput{
		PolicyName:     aws.String(policyName),
		PolicyDocument: aws.String(string(policyDoc)),
		Description:    aws.String("test policy"),
	})

	var e *iamtypes.EntityAlreadyExistsException
	if err != nil && !errors.As(err, &e) {
		return fmt.Errorf("failed to create policy: %v", err)
	}

	// Attach policy to IAM roles
	policyArn := fmt.Sprintf("arn:aws:iam::%s:policy/%s", s.AccountID, policyName)
	for _, resource := range resources {
		if resource["ResourceType"] != "AWS::IAM::Role" {
			continue
		}

		roleName := resource["PhysicalResourceId"]

		_, err = s.IamClient.AttachRolePolicy(s.Ctx, &iam.AttachRolePolicyInput{
			RoleName:  aws.String(roleName),
			PolicyArn: aws.String(policyArn),
		})

		if err != nil {
			return fmt.Errorf("failed to attach policy to role: %v", err)
		}
	}

	return nil
}

func (s *DeployStackService) attachUserToGroup(stackName string) error {
	// Get resources in the stack
	resources, nestedStackNames, err := s.getStackResources(stackName)
	if err != nil {
		return err
	}

	// Process nested stacks in parallel
	var wg sync.WaitGroup
	errorChan := make(chan error, len(nestedStackNames))

	for _, nestedStackName := range nestedStackNames {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()
			if nestErr := s.attachUserToGroup(name); nestErr != nil {
				errorChan <- nestErr
			}
		}(nestedStackName)
	}

	wg.Wait()
	close(errorChan)

	for err := range errorChan {
		if err != nil {
			return err
		}
	}

	// Create user if it doesn't exist
	userName := "DelstackTestUser"

	_, err = s.IamClient.CreateUser(s.Ctx, &iam.CreateUserInput{
		UserName: aws.String(userName),
	})

	var e *iamtypes.EntityAlreadyExistsException
	if err != nil && !errors.As(err, &e) {
		return fmt.Errorf("failed to create user: %v", err)
	}

	// Add user to IAM groups
	for _, resource := range resources {
		if resource["ResourceType"] != "AWS::IAM::Group" {
			continue
		}

		groupName := resource["PhysicalResourceId"]

		_, err = s.IamClient.AddUserToGroup(s.Ctx, &iam.AddUserToGroupInput{
			GroupName: aws.String(groupName),
			UserName:  aws.String(userName),
		})

		if err != nil {
			return fmt.Errorf("failed to add user to group: %v", err)
		}
	}

	return nil
}

func (s *DeployStackService) buildUpload(stackName string) error {
	// Get resources in the stack
	resources, nestedStackNames, err := s.getStackResources(stackName)
	if err != nil {
		return err
	}

	// Process nested stacks in parallel
	var wg sync.WaitGroup
	errorChan := make(chan error, len(nestedStackNames))

	for _, nestedStackName := range nestedStackNames {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()
			if err := s.buildUpload(name); err != nil {
				errorChan <- err
			}
		}(nestedStackName)
	}

	wg.Wait()
	close(errorChan)

	for err := range errorChan {
		if err != nil {
			return err
		}
	}

	// Push image to ECR repositories
	ecrRepositoryEndpoint := fmt.Sprintf("%s.dkr.ecr.%s.amazonaws.com", s.AccountID, region)

	for _, resource := range resources {
		if resource["ResourceType"] != "AWS::ECR::Repository" {
			continue
		}

		ecrName := resource["PhysicalResourceId"]
		ecrRepositoryUri := fmt.Sprintf("%s/%s", ecrRepositoryEndpoint, ecrName)
		ecrTag := "test"
		uriTag := fmt.Sprintf("%s:%s", ecrRepositoryUri, ecrTag)

		// Tag image
		tagCmd := exec.Command("docker", "tag", "delstack-test:latest", uriTag)
		tagCmd.Stdout = os.Stdout
		tagCmd.Stderr = os.Stderr
		if err := tagCmd.Run(); err != nil {
			return fmt.Errorf("failed to tag Docker image: %v", err)
		}

		// Push image
		pushCmd := exec.Command("docker", "push", uriTag)
		pushCmd.Stdout = os.Stdout
		pushCmd.Stderr = os.Stderr
		if err := pushCmd.Run(); err != nil {
			return fmt.Errorf("failed to push Docker image: %v", err)
		}

		color.Green("Successfully pushed image to %s", uriTag)
	}

	return nil
}

func (s *DeployStackService) objectUpload(stackName string) error {
	// Get resources in the stack
	resources, nestedStackNames, err := s.getStackResources(stackName)
	if err != nil {
		return err
	}

	// Process nested stacks in parallel
	var wg sync.WaitGroup
	errorChan := make(chan error, len(nestedStackNames))

	for _, nestedStackName := range nestedStackNames {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()
			if err := s.objectUpload(name); err != nil {
				errorChan <- err
			}
		}(nestedStackName)
	}

	wg.Wait()
	close(errorChan)

	for err := range errorChan {
		if err != nil {
			return err
		}
	}

	// Use a wait group for parallel uploads and deletions
	var uploadWg sync.WaitGroup
	uploadErrorChan := make(chan error, 10)

	// Upload objects to S3 buckets using AWS SDK
	for _, resource := range resources {
		resourceType := resource["ResourceType"]
		bucketName := resource["PhysicalResourceId"]

		if resourceType == "AWS::S3::Bucket" {
			// Run both uploads in parallel
			uploadWg.Add(2)

			// First upload
			go func(bucket string) {
				defer uploadWg.Done()
				if err := s.uploadDirectoryToS3(bucket, false); err != nil {
					uploadErrorChan <- fmt.Errorf("failed to upload files to S3 bucket: %v", err)
				}
			}(bucketName)

			// Second upload (for versioning)
			go func(bucket string) {
				defer uploadWg.Done()
				if err := s.uploadDirectoryToS3(bucket, false); err != nil {
					uploadErrorChan <- fmt.Errorf("failed to upload files to S3 bucket (version): %v", err)
				}
			}(bucketName)

			// Wait for both uploads to complete
			uploadWg.Wait()

			// Check for errors
			select {
			case err := <-uploadErrorChan:
				return err
			default:
				// No errors, continue with deletion
			}

			// Delete all objects to create delete markers
			if err := s.deleteS3BucketContents(bucketName); err != nil {
				return fmt.Errorf("failed to delete objects from S3 bucket: %v", err)
			}
		}

		if resourceType == "AWS::S3Express::DirectoryBucket" {
			// Upload files to directory bucket.
			// Ignore errors even in the event of an error because the following error will occur.
			// upload failed: testfiles/5594.txt to s3://dev-goto-002-descend--use1-az4--x-s3/5594.txt An error occurred (400) when calling the PutObject operation: Bad Request
			_ = s.uploadDirectoryToS3(bucketName, true)
		}
	}

	return nil
}

// uploadDirectoryToS3 - Upload virtual test files to S3 bucket
func (s *DeployStackService) uploadDirectoryToS3(bucketName string, ignoreErrors bool) error {
	var uploadCount int
	var errorCount int
	var mu sync.Mutex

	// Create a semaphore with 20 max weights for parallel processing
	sem := semaphore.NewWeighted(20)
	var wg sync.WaitGroup

	// Channel to collect errors
	errorChan := make(chan error, 100)

	// Number of virtual files to create (same as before)
	totalFiles := 1500

	// Upload files in parallel
	for i := 1; i <= totalFiles; i++ {
		wg.Add(1)

		// Acquire a semaphore slot
		if err := sem.Acquire(s.Ctx, 1); err != nil {
			errorChan <- fmt.Errorf("failed to acquire semaphore: %v", err)
			wg.Done()
			continue
		}

		go func(fileNum int) {
			defer wg.Done()
			defer sem.Release(1) // Release the semaphore slot

			// Virtual file path (no need to actually create the file)
			s3Key := fmt.Sprintf("%d.txt", fileNum)

			// Create a reader for the empty content
			contentReader := strings.NewReader("")

			// Upload to S3 directly
			_, err := s.S3Client.PutObject(s.Ctx, &s3.PutObjectInput{
				Bucket: aws.String(bucketName),
				Key:    aws.String(s3Key),
				Body:   contentReader,
			})

			if err != nil {
				if ignoreErrors {
					mu.Lock()
					errorCount++
					mu.Unlock()
					return
				}
				errorChan <- fmt.Errorf("failed to upload %s to %s: %v", s3Key, bucketName, err)
				return
			}

			mu.Lock()
			uploadCount++
			mu.Unlock()
		}(i)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(errorChan)

	// Return errors if any
	for err := range errorChan {
		if !ignoreErrors {
			return err
		}
		// For ignoreErrors mode, just increment the error count
		errorCount++
	}

	if errorCount > 0 {
		color.Yellow("Uploaded %d files to bucket %s (%d errors ignored)", uploadCount, bucketName, errorCount)
	} else {
		color.Green("Uploaded %d files to bucket %s", uploadCount, bucketName)
	}

	return nil
}

// deleteS3BucketContents - Delete all objects in a bucket using AWS SDK
func (s *DeployStackService) deleteS3BucketContents(bucketName string) error {
	// List objects
	listObjOutput, err := s.S3Client.ListObjectsV2(s.Ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
	})

	if err != nil {
		return fmt.Errorf("failed to list objects in bucket %s: %v", bucketName, err)
	}

	// Do nothing if no objects exist
	if len(listObjOutput.Contents) == 0 {
		return nil
	}

	// Create a semaphore with 20 max weights for parallel processing
	sem := semaphore.NewWeighted(20)
	var wg sync.WaitGroup
	errorChan := make(chan error, 100)

	// Process objects in batches of 1000 (AWS DeleteObjects can handle up to 1000)
	batchSize := 1000
	totalObjects := len(listObjOutput.Contents)
	totalBatches := (totalObjects + batchSize - 1) / batchSize

	// Process each batch in parallel
	for i := range totalBatches {
		wg.Add(1)

		// Determine batch range
		start := i * batchSize
		end := (i + 1) * batchSize
		if end > totalObjects {
			end = totalObjects
		}

		// Extract the batch of objects
		objectBatch := listObjOutput.Contents[start:end]

		// Acquire a semaphore slot
		if err := sem.Acquire(s.Ctx, 1); err != nil {
			errorChan <- fmt.Errorf("failed to acquire semaphore: %v", err)
			wg.Done()
			continue
		}

		go func(objects []types.Object, batchNum int) {
			defer wg.Done()
			defer sem.Release(1) // Release the semaphore slot

			// Create ObjectIdentifier slice for this batch
			var objectsToDelete []types.ObjectIdentifier
			for _, obj := range objects {
				objectsToDelete = append(objectsToDelete, types.ObjectIdentifier{
					Key: obj.Key,
				})
			}

			// Delete this batch of objects
			_, err := s.S3Client.DeleteObjects(s.Ctx, &s3.DeleteObjectsInput{
				Bucket: aws.String(bucketName),
				Delete: &types.Delete{
					Objects: objectsToDelete,
					Quiet:   aws.Bool(true), // Set to true to reduce response size
				},
			})

			if err != nil {
				errorChan <- fmt.Errorf("failed to delete objects batch %d from bucket %s: %v", batchNum, bucketName, err)
			}
		}(objectBatch, i)
	}

	// Wait for all deletions to complete
	wg.Wait()
	close(errorChan)

	// Check for errors
	for err := range errorChan {
		return err
	}

	return nil
}

func (s *DeployStackService) tablesUploadToTableBucket(stackName string) error {
	// Get resources in the stack
	resources, nestedStackNames, err := s.getStackResources(stackName)
	if err != nil {
		return err
	}

	// Process nested stacks in parallel
	var wg sync.WaitGroup
	errorChan := make(chan error, len(nestedStackNames))

	for _, nestedStackName := range nestedStackNames {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()
			if err := s.tablesUploadToTableBucket(name); err != nil {
				errorChan <- err
			}
		}(nestedStackName)
	}

	wg.Wait()
	close(errorChan)

	for err := range errorChan {
		if err != nil {
			return err
		}
	}

	sdkNamespaceAmount := 2
	tableAmount := 2

	// Create namespaces and tables in the table bucket
	for _, resource := range resources {
		if resource["ResourceType"] != "AWS::S3Tables::TableBucket" {
			continue
		}

		tableBucketArn := resource["PhysicalResourceId"]

		// List CFN namespaces
		listNamespacesOutput, err := s.S3TablesClient.ListNamespaces(s.Ctx, &s3tables.ListNamespacesInput{
			TableBucketARN: aws.String(tableBucketArn),
		})
		if err != nil {
			return fmt.Errorf("failed to list namespaces: %v", err)
		}

		cfnNamespaces := []string{}
		for _, ns := range listNamespacesOutput.Namespaces {
			cfnNamespaces = append(cfnNamespaces, strings.Join(ns.Namespace, "/"))
		}

		var eg errgroup.Group
		sem := semaphore.NewWeighted(16)

		// Create SDK namespaces and tables
		sdkNamespaces := make([]string, 0, sdkNamespaceAmount)
		for i := range sdkNamespaceAmount {
			if err := sem.Acquire(s.Ctx, 1); err != nil {
				return fmt.Errorf("failed to acquire semaphore: %v", err)
			}

			namespaceNum := i
			eg.Go(func() error {
				defer sem.Release(1)
				namespaceName := fmt.Sprintf("sdk_namespace_%d", namespaceNum)

				// Create namespace using SDK
				_, err := s.S3TablesClient.CreateNamespace(s.Ctx, &s3tables.CreateNamespaceInput{
					TableBucketARN: aws.String(tableBucketArn),
					Namespace:      []string{namespaceName},
				})
				if err != nil {
					return fmt.Errorf("failed to create namespace %s: %v", namespaceName, err)
				}

				return nil
			})

			sdkNamespaces = append(sdkNamespaces, fmt.Sprintf("sdk_namespace_%d", namespaceNum))
		}

		if err := eg.Wait(); err != nil {
			return fmt.Errorf("failed to create SDK namespaces: %v", err)
		}

		// Create tables in both SDK and CFN namespaces
		allNamespaces := make([]string, 0, len(sdkNamespaces)+len(cfnNamespaces))
		allNamespaces = append(allNamespaces, sdkNamespaces...)
		allNamespaces = append(allNamespaces, cfnNamespaces...)

		for _, namespaceName := range allNamespaces {
			if err := sem.Acquire(s.Ctx, 1); err != nil {
				return fmt.Errorf("failed to acquire semaphore: %v", err)
			}

			currentNamespaceName := namespaceName
			eg.Go(func() error {
				defer sem.Release(1)

				for j := range tableAmount {
					tableName := fmt.Sprintf("sdk_table_%d", j)

					// Create metadata structure for Iceberg table
					schemaField := s3tablesTypes.SchemaField{
						Name:     aws.String("column"),
						Type:     aws.String("int"),
						Required: false,
					}

					icebergSchema := &s3tablesTypes.IcebergSchema{
						Fields: []s3tablesTypes.SchemaField{schemaField},
					}

					icebergMetadata := &s3tablesTypes.IcebergMetadata{
						Schema: icebergSchema,
					}

					tableMetadata := &s3tablesTypes.TableMetadataMemberIceberg{
						Value: *icebergMetadata,
					}

					_, err := s.S3TablesClient.CreateTable(s.Ctx, &s3tables.CreateTableInput{
						TableBucketARN: aws.String(tableBucketArn),
						Namespace:      aws.String(currentNamespaceName),
						Name:           aws.String(tableName),
						Metadata:       tableMetadata,
						Format:         s3tablesTypes.OpenTableFormatIceberg,
					})

					if err != nil {
						return fmt.Errorf("failed to create table %s in namespace %s: %v", tableName, currentNamespaceName, err)
					}
				}

				return nil
			})
		}

		if err := eg.Wait(); err != nil {
			return fmt.Errorf("failed to create tables: %v", err)
		}
	}

	return nil
}

func (s *DeployStackService) indexesUploadToVectorBucket(stackName string) error {
	// Get resources in the stack
	resources, nestedStackNames, err := s.getStackResources(stackName)
	if err != nil {
		return err
	}

	// Process nested stacks in parallel
	var wg sync.WaitGroup
	errorChan := make(chan error, len(nestedStackNames))

	for _, nestedStackName := range nestedStackNames {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()
			if err := s.indexesUploadToVectorBucket(name); err != nil {
				errorChan <- err
			}
		}(nestedStackName)
	}

	wg.Wait()
	close(errorChan)

	for err := range errorChan {
		if err != nil {
			return err
		}
	}

	sdkIndexAmount := 10
	vectorsPerIndex := 100

	// Create indexes in the vector bucket and upload vectors
	for _, resource := range resources {
		if resource["ResourceType"] != "AWS::S3Vectors::VectorBucket" {
			continue
		}

		physicalResourceId := resource["PhysicalResourceId"]

		// PhysicalResourceId is ARN format: arn:aws:s3vectors:region:account-id:vector-bucket/bucket-name
		// Extract the bucket name from the ARN
		vectorBucketName := physicalResourceId
		parts := strings.Split(physicalResourceId, "/")
		if len(parts) > 0 {
			vectorBucketName = parts[len(parts)-1]
		}

		// List CFN indexes
		listIndexesOutput, err := s.S3VectorsClient.ListIndexes(s.Ctx, &s3vectors.ListIndexesInput{
			VectorBucketName: aws.String(vectorBucketName),
		})
		if err != nil {
			return fmt.Errorf("failed to list indexes: %v", err)
		}

		cfnIndexes := []string{}
		for _, index := range listIndexesOutput.Indexes {
			cfnIndexes = append(cfnIndexes, aws.ToString(index.IndexName))
		}

		var eg errgroup.Group
		sem := semaphore.NewWeighted(16)

		// Create SDK indexes and upload vectors
		sdkIndexes := make([]string, 0, sdkIndexAmount)
		for i := range sdkIndexAmount {
			if err := sem.Acquire(s.Ctx, 1); err != nil {
				return fmt.Errorf("failed to acquire semaphore: %v", err)
			}

			indexNum := i
			eg.Go(func() error {
				defer sem.Release(1)
				indexName := fmt.Sprintf("sdk-index-%d", indexNum)

				// Create vector index using SDK
				_, err := s.S3VectorsClient.CreateIndex(s.Ctx, &s3vectors.CreateIndexInput{
					VectorBucketName: aws.String(vectorBucketName),
					IndexName:        aws.String(indexName),
					DataType:         s3vectorsTypes.DataTypeFloat32,
					Dimension:        aws.Int32(128),
					DistanceMetric:   s3vectorsTypes.DistanceMetricCosine,
				})

				if err != nil {
					return fmt.Errorf("failed to create index %s: %v", indexName, err)
				}

				return nil
			})

			sdkIndexes = append(sdkIndexes, fmt.Sprintf("sdk-index-%d", indexNum))
		}

		if err := eg.Wait(); err != nil {
			return fmt.Errorf("failed to create SDK indexes: %v", err)
		}

		// Upload vectors to both SDK and CFN indexes
		allIndexes := make([]string, 0, len(sdkIndexes)+len(cfnIndexes))
		allIndexes = append(allIndexes, sdkIndexes...)
		allIndexes = append(allIndexes, cfnIndexes...)

		for _, indexName := range allIndexes {
			if err := sem.Acquire(s.Ctx, 1); err != nil {
				return fmt.Errorf("failed to acquire semaphore: %v", err)
			}

			currentIndexName := indexName
			eg.Go(func() error {
				defer sem.Release(1)

				// Process vectors in batches of 500 (PutVectors API limit)
				batchSize := 500
				for batchStart := 1; batchStart <= vectorsPerIndex; batchStart += batchSize {
					batchEnd := batchStart + batchSize - 1
					if batchEnd > vectorsPerIndex {
						batchEnd = vectorsPerIndex
					}

					// Create batch of vectors
					vectors := make([]s3vectorsTypes.PutInputVector, 0, batchEnd-batchStart+1)
					for vector := batchStart; vector <= batchEnd; vector++ {
						vectorId := fmt.Sprintf("vector-%d", vector)

						// Generate sample vector data (128 dimensions)
						vectorData := make([]float32, 128)
						for j := range vectorData {
							vectorData[j] = rand.Float32()
						}

						vectors = append(vectors, s3vectorsTypes.PutInputVector{
							Key:  aws.String(vectorId),
							Data: &s3vectorsTypes.VectorDataMemberFloat32{Value: vectorData},
						})
					}

					// Upload batch
					_, err := s.S3VectorsClient.PutVectors(s.Ctx, &s3vectors.PutVectorsInput{
						VectorBucketName: aws.String(vectorBucketName),
						IndexName:        aws.String(currentIndexName),
						Vectors:          vectors,
					})
					if err != nil {
						return fmt.Errorf("failed to put vectors to index %s: %v", currentIndexName, err)
					}
				}

				return nil
			})
		}

		if err := eg.Wait(); err != nil {
			return fmt.Errorf("failed to upload vectors to indexes: %v", err)
		}
	}

	return nil
}

func (s *DeployStackService) startBackup(stackName string) error {
	// Get resources in the stack
	resources, nestedStackNames, err := s.getStackResources(stackName)
	if err != nil {
		return err
	}

	// Process nested stacks in parallel
	var wg sync.WaitGroup
	errorChan := make(chan error, len(nestedStackNames))

	for _, nestedStackName := range nestedStackNames {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()
			if err := s.startBackup(name); err != nil {
				errorChan <- err
			}
		}(nestedStackName)
	}

	wg.Wait()
	close(errorChan)

	for err := range errorChan {
		if err != nil {
			return err
		}
	}

	// Start backup jobs
	resourceArn := fmt.Sprintf("arn:aws:dynamodb:%s:%s:table/%s-Root-Table", region, s.AccountID, s.CfnPjPrefix)
	iamRoleArn := fmt.Sprintf("arn:aws:iam::%s:role/service-role/%s-Root-AWSBackupServiceRole", s.AccountID, s.CfnPjPrefix)

	for _, resource := range resources {
		if resource["ResourceType"] != "AWS::Backup::BackupVault" {
			continue
		}

		vaultName := resource["PhysicalResourceId"]

		// Start backup job
		startBackupJobOutput, err := s.BackupClient.StartBackupJob(s.Ctx, &backup.StartBackupJobInput{
			BackupVaultName: aws.String(vaultName),
			ResourceArn:     aws.String(resourceArn),
			IamRoleArn:      aws.String(iamRoleArn),
		})

		if err != nil {
			return fmt.Errorf("failed to start backup job: %v", err)
		}

		backupJobId := *startBackupJobOutput.BackupJobId

		// Wait for backup to complete
		for {
			describeJobOutput, err := s.BackupClient.DescribeBackupJob(s.Ctx, &backup.DescribeBackupJobInput{
				BackupJobId: aws.String(backupJobId),
			})

			if err != nil {
				return fmt.Errorf("failed to describe backup job: %v", err)
			}

			state := describeJobOutput.State

			if state == "COMPLETED" {
				break
			} else if state == "FAILED" || state == "ABORTED" {
				return fmt.Errorf("backup failed: %s", state)
			}

			time.Sleep(10 * time.Second)
		}
	}

	return nil
}

func (s *DeployStackService) getStackResources(stackName string) ([]map[string]string, []string, error) {
	// List stack resources
	output, err := s.CfnClient.ListStackResources(s.Ctx, &cloudformation.ListStackResourcesInput{
		StackName: aws.String(stackName),
	})

	if err != nil {
		return nil, nil, fmt.Errorf("failed to list stack resources: %v", err)
	}

	// Extract resources and nested stacks
	resources := make([]map[string]string, 0)
	nestedStackNames := make([]string, 0)

	for _, resource := range output.StackResourceSummaries {
		resourceMap := map[string]string{
			"LogicalResourceId":  *resource.LogicalResourceId,
			"PhysicalResourceId": *resource.PhysicalResourceId,
			"ResourceType":       *resource.ResourceType,
		}

		resources = append(resources, resourceMap)

		// Check if this is a nested CloudFormation stack
		if *resource.ResourceType == "AWS::CloudFormation::Stack" {
			// Extract stack name from ARN
			stackArn := *resource.PhysicalResourceId
			parts := strings.Split(stackArn, "/")
			if len(parts) >= 2 {
				nestedStackName := parts[1]
				nestedStackNames = append(nestedStackNames, nestedStackName)
			}
		}
	}

	return resources, nestedStackNames, nil
}
