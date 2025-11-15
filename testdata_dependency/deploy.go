package main

import (
	"context"
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
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
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
	Profile    string
	Stage      string
	RetainMode bool
}

type DeployStackService struct {
	Options       Options
	CfnPjPrefix   string
	AccountID     string
	ProfileOption string
	Ctx           context.Context
	CfnClient     *cloudformation.Client
	S3Client      *s3.Client
	StsClient     *sts.Client
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

	// Deploy using CDK
	if err := service.cdkDeploy(); err != nil {
		color.Red("Failed to deploy: %v", err)
		os.Exit(1)
	}

	// Upload objects to S3 for all stacks (A, B, C, D, E, F)
	color.Green("=== object_upload ===")
	stackNames := []string{
		fmt.Sprintf("%s-Stack-A", service.CfnPjPrefix),
		fmt.Sprintf("%s-Stack-B", service.CfnPjPrefix),
		fmt.Sprintf("%s-Stack-C", service.CfnPjPrefix),
		fmt.Sprintf("%s-Stack-D", service.CfnPjPrefix),
		fmt.Sprintf("%s-Stack-E", service.CfnPjPrefix),
		fmt.Sprintf("%s-Stack-F", service.CfnPjPrefix),
	}
	for _, stackName := range stackNames {
		if err := service.objectUpload(stackName); err != nil {
			color.Red("Failed to upload objects: %v", err)
			os.Exit(1)
		}
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

	profileOption := ""
	if options.Profile != "" {
		profileOption = fmt.Sprintf("--profile %s --region %s", options.Profile, region)
	}

	return &DeployStackService{
		Options:       options,
		CfnPjPrefix:   cfnPjPrefix,
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
	s.StsClient = sts.NewFromConfig(cfg)

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
