package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/backup"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/fatih/color"
	"golang.org/x/sync/semaphore"
)

const (
	region = "us-east-1"
)

type Options struct {
	Profile string
	Stage   string
}

type DeployStackService struct {
	Options           Options
	CfnTemplate       string
	CfnOutputTemplate string
	CfnPjPrefix       string
	CfnStackName      string
	SamBucket         string
	AccountID         string
	ProfileOption     string
	Ctx               context.Context
	CfnClient         *cloudformation.Client
	S3Client          *s3.Client
	IamClient         *iam.Client
	EcrClient         *ecr.Client
	StsClient         *sts.Client
	BackupClient      *backup.Client
}

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

	// Ensure the S3 bucket exists
	if err := service.ensureS3Bucket(); err != nil {
		color.Red("Failed to create S3 bucket: %v", err)
		os.Exit(1)
	}

	// Package and deploy using SAM
	if err := service.packageAndDeploy(); err != nil {
		color.Red("Failed to package and deploy: %v", err)
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
		}
	}

	return options
}

func NewDeployStackService(ctx context.Context, options Options) *DeployStackService {
	cfnPjPrefix := fmt.Sprintf("dev-%s", options.Stage)
	cfnStackName := fmt.Sprintf("%s-TestStack", cfnPjPrefix)
	samBucket := strings.ToLower(cfnStackName)

	profileOption := ""
	if options.Profile != "" {
		profileOption = fmt.Sprintf("--profile %s --region %s", options.Profile, region)
	}

	return &DeployStackService{
		Options:           options,
		CfnTemplate:       "./yamldir/test_root.yaml",
		CfnOutputTemplate: "./yamldir/test_root_output.yaml",
		CfnPjPrefix:       cfnPjPrefix,
		CfnStackName:      cfnStackName,
		SamBucket:         samBucket,
		ProfileOption:     profileOption,
		Ctx:               ctx,
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

func (s *DeployStackService) ensureS3Bucket() error {
	// Check if bucket exists
	listBucketsOutput, err := s.S3Client.ListBuckets(s.Ctx, &s3.ListBucketsInput{})
	if err != nil {
		return fmt.Errorf("failed to list S3 buckets: %v", err)
	}

	bucketExists := false
	for _, bucket := range listBucketsOutput.Buckets {
		if *bucket.Name == s.SamBucket {
			bucketExists = true
			break
		}
	}

	// Create bucket if it doesn't exist
	if !bucketExists {
		_, err := s.S3Client.CreateBucket(s.Ctx, &s3.CreateBucketInput{
			Bucket: aws.String(s.SamBucket),
		})
		if err != nil {
			return fmt.Errorf("failed to create S3 bucket: %v", err)
		}
	}

	return nil
}

func (s *DeployStackService) packageAndDeploy() error {
	// Login to ECR using AWS SDK
	if err := s.loginToECR(); err != nil {
		return fmt.Errorf("failed to login to ECR: %v", err)
	}

	// Build Docker image
	buildCmd := "docker build -t delstack-test ."
	if err := s.runCommand(buildCmd); err != nil {
		return fmt.Errorf("failed to build Docker image: %v", err)
	}

	// SAM package
	packageCmd := fmt.Sprintf("sam package --template-file %s --output-template-file %s --s3-bucket %s %s",
		s.CfnTemplate,
		s.CfnOutputTemplate,
		s.SamBucket,
		s.ProfileOption)

	if err := s.runCommand(packageCmd); err != nil {
		return fmt.Errorf("failed to package with SAM: %v", err)
	}

	// SAM deploy
	deployCmd := fmt.Sprintf("sam deploy --template-file %s --stack-name %s --capabilities CAPABILITY_IAM CAPABILITY_AUTO_EXPAND CAPABILITY_NAMED_IAM --parameter-overrides PJPrefix=%s %s",
		s.CfnOutputTemplate,
		s.CfnStackName,
		s.CfnPjPrefix,
		s.ProfileOption)

	if err := s.runCommand(deployCmd); err != nil {
		return fmt.Errorf("failed to deploy with SAM: %v", err)
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

	// Check if policy exists
	policyArn := fmt.Sprintf("arn:aws:iam::%s:policy/DelstackTestPolicy", s.AccountID)

	_, err = s.IamClient.GetPolicy(s.Ctx, &iam.GetPolicyInput{
		PolicyArn: aws.String(policyArn),
	})

	if err != nil {
		// Create policy if it doesn't exist
		policyDoc, readErr := os.ReadFile("./policy_document.json")
		if readErr != nil {
			return fmt.Errorf("failed to read policy document: %v", readErr)
		}

		_, readErr = s.IamClient.CreatePolicy(s.Ctx, &iam.CreatePolicyInput{
			PolicyName:     aws.String("DelstackTestPolicy"),
			PolicyDocument: aws.String(string(policyDoc)),
			Description:    aws.String("test policy"),
		})

		if readErr != nil {
			return fmt.Errorf("failed to create policy: %v", readErr)
		}
	}

	// Attach policy to IAM roles
	for _, resource := range resources {
		if resource["ResourceType"] == "AWS::IAM::Role" {
			roleName := resource["PhysicalResourceId"]

			_, err = s.IamClient.AttachRolePolicy(s.Ctx, &iam.AttachRolePolicyInput{
				RoleName:  aws.String(roleName),
				PolicyArn: aws.String(policyArn),
			})

			if err != nil {
				return fmt.Errorf("failed to attach policy to role: %v", err)
			}
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

	_, err = s.IamClient.GetUser(s.Ctx, &iam.GetUserInput{
		UserName: aws.String(userName),
	})

	if err != nil {
		_, err = s.IamClient.CreateUser(s.Ctx, &iam.CreateUserInput{
			UserName: aws.String(userName),
		})

		if err != nil {
			return fmt.Errorf("failed to create user: %v", err)
		}
	}

	// Add user to IAM groups
	for _, resource := range resources {
		if resource["ResourceType"] == "AWS::IAM::Group" {
			groupName := resource["PhysicalResourceId"]

			_, err = s.IamClient.AddUserToGroup(s.Ctx, &iam.AddUserToGroupInput{
				GroupName: aws.String(groupName),
				UserName:  aws.String(userName),
			})

			if err != nil {
				return fmt.Errorf("failed to add user to group: %v", err)
			}
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
		if resource["ResourceType"] == "AWS::ECR::Repository" {
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
		if resource["ResourceType"] == "AWS::S3::Bucket" {
			bucketName := resource["PhysicalResourceId"]

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

			// Delete all objects for delete markers
			if err := s.deleteS3BucketContents(bucketName); err != nil {
				return fmt.Errorf("failed to delete objects from S3 bucket: %v", err)
			}

		} else if resource["ResourceType"] == "AWS::S3Express::DirectoryBucket" {
			bucketName := resource["PhysicalResourceId"]

			// Upload files to directory bucket (ignore errors)
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

	// Process objects in batches of 100 (AWS DeleteObjects can handle up to 1000,
	// but we're using smaller batches for more parallelism)
	batchSize := 100
	totalObjects := len(listObjOutput.Contents)
	totalBatches := (totalObjects + batchSize - 1) / batchSize

	// Process each batch in parallel
	for i := 0; i < totalBatches; i++ {
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
	resourceArn := fmt.Sprintf("arn:aws:dynamodb:%s:%s:table/%s-Table", region, s.AccountID, s.CfnPjPrefix)
	iamRoleArn := fmt.Sprintf("arn:aws:iam::%s:role/service-role/%s-AWSBackupServiceRole", s.AccountID, s.CfnPjPrefix)

	for _, resource := range resources {
		if resource["ResourceType"] == "AWS::Backup::BackupVault" {
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
