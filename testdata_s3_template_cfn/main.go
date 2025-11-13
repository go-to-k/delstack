package main

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	cfnTypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3Types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

const (
	templateFile = "large-template.yaml"
)

func main() {
	ctx := context.Background()

	// Parse command line arguments
	profile := ""
	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		if args[i] == "-p" || args[i] == "--profile" {
			if i+1 < len(args) {
				profile = args[i+1]
				i++
			}
		}
	}

	// Generate random 4-digit suffix
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	randomNum := r.Intn(10000)
	randomSuffix := fmt.Sprintf("%04d", randomNum)

	stackName := fmt.Sprintf("delstack-test-large-template-%s", randomSuffix)

	// Check if template exists
	if _, err := os.Stat(templateFile); err == nil {
		// Template already exists, skip generation
	} else {
		// Generate large template
		fmt.Println("=== Generating large CloudFormation template ===")
		if err := generateLargeTemplate(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}

	// Load AWS config
	optFns := []func(*config.LoadOptions) error{}
	if profile != "" {
		optFns = append(optFns, config.WithSharedConfigProfile(profile))
	}

	cfg, err := config.LoadDefaultConfig(ctx, optFns...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to load AWS config: %v\n", err)
		os.Exit(1)
	}

	// Get AWS account ID and region
	accountID, region, err := getAWSInfo(ctx, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to get AWS info: %v\n", err)
		os.Exit(1)
	}

	if region == "" {
		region = "us-east-1"
	}

	// Create S3 bucket with random suffix
	bucketName := fmt.Sprintf("delstack-test-cfn-templates-%s-%s-%s", accountID, region, randomSuffix)
	fmt.Printf("Creating S3 bucket: %s\n", bucketName)
	if err := createS3Bucket(ctx, cfg, bucketName, region); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Upload template to S3
	fmt.Println("Uploading template to S3...")
	if err := uploadTemplateToS3(ctx, cfg, bucketName); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Create CloudFormation stack
	fmt.Printf("Creating CloudFormation stack: %s\n", stackName)
	templateURL := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", bucketName, region, templateFile)
	if err := createStack(ctx, cfg, stackName, templateURL); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Wait for stack creation to complete
	fmt.Println("Waiting for stack creation to complete...")
	if err := waitStackCreate(ctx, cfg, stackName); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Stack created successfully: %s\n", stackName)

	// Delete S3 bucket (no longer needed after stack creation)
	fmt.Printf("Deleting S3 bucket: %s\n", bucketName)
	if err := deleteS3Bucket(ctx, cfg, bucketName); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to delete S3 bucket: %v\n", err)
	} else {
		fmt.Println("✓ S3 bucket deleted successfully")
	}
}

func generateLargeTemplate() error {
	var builder strings.Builder

	// Write header
	builder.WriteString(`AWSTemplateFormatVersion: '2010-09-09'
Description: 'Large CloudFormation template for testing DeletionPolicy with S3 upload (>51200 bytes)'

Resources:
`)

	// Create 150 IAM roles to make the template large (>51200 bytes)
	// Each role definition is ~350 bytes, so 150 roles = ~52KB
	for i := 1; i <= 150; i++ {
		builder.WriteString(fmt.Sprintf(`  TestRole%d:
    Type: AWS::IAM::Role
    DeletionPolicy: Retain
    Properties:
      AssumeRolePolicyDocument:
        Version: '2012-10-17'
        Statement:
          - Effect: Allow
            Principal:
              Service: lambda.amazonaws.com
            Action: 'sts:AssumeRole'
      Tags:
        - Key: Index
          Value: '%d'
        - Key: Purpose
          Value: 'Testing large CloudFormation template for delstack'

`, i, i))
	}

	// Write footer
	builder.WriteString(`
Outputs:
  FirstRoleArn:
    Description: 'ARN of the first IAM role'
    Value: !GetAtt TestRole1.Arn
`)

	templateContent := builder.String()

	// Write to file
	if err := os.WriteFile(templateFile, []byte(templateContent), 0644); err != nil {
		return fmt.Errorf("failed to write template file: %w", err)
	}

	// Check file size
	fileInfo, err := os.Stat(templateFile)
	if err != nil {
		return fmt.Errorf("failed to stat template file: %w", err)
	}

	fileSize := fileInfo.Size()
	fmt.Printf("Generated template size: %d bytes\n", fileSize)

	if fileSize > 51200 {
		fmt.Println("✓ Template is larger than 51200 bytes (will require S3 upload for UpdateStack)")
	} else {
		return fmt.Errorf("template is not large enough: %d bytes", fileSize)
	}

	return nil
}

func getAWSInfo(ctx context.Context, cfg aws.Config) (string, string, error) {
	// Get account ID using STS
	stsClient := sts.NewFromConfig(cfg)
	identity, err := stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return "", "", fmt.Errorf("failed to get caller identity: %w", err)
	}

	accountID := *identity.Account
	region := cfg.Region

	return accountID, region, nil
}

func createS3Bucket(ctx context.Context, cfg aws.Config, bucketName, region string) error {
	s3Client := s3.NewFromConfig(cfg)

	input := &s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	}

	// For regions other than us-east-1, we need to specify the LocationConstraint
	if region != "us-east-1" && region != "" {
		input.CreateBucketConfiguration = &s3Types.CreateBucketConfiguration{
			LocationConstraint: s3Types.BucketLocationConstraint(region),
		}
	}

	_, err := s3Client.CreateBucket(ctx, input)
	if err != nil {
		// Bucket might already exist
		var bucketAlreadyExists *s3Types.BucketAlreadyOwnedByYou
		var bucketAlreadyExistsElsewhere *s3Types.BucketAlreadyExists
		if errors.As(err, &bucketAlreadyExists) || errors.As(err, &bucketAlreadyExistsElsewhere) {
			fmt.Println("Bucket already exists")
			return nil
		}
		return fmt.Errorf("failed to create bucket: %w", err)
	}

	return nil
}

func uploadTemplateToS3(ctx context.Context, cfg aws.Config, bucketName string) error {
	s3Client := s3.NewFromConfig(cfg)

	// Read template file
	templateContent, err := os.ReadFile(templateFile)
	if err != nil {
		return fmt.Errorf("failed to read template file: %w", err)
	}

	// Upload to S3
	_, err = s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(templateFile),
		Body:   strings.NewReader(string(templateContent)),
	})
	if err != nil {
		return fmt.Errorf("failed to upload template to S3: %w", err)
	}

	return nil
}

func createStack(ctx context.Context, cfg aws.Config, stackName, templateURL string) error {
	cfnClient := cloudformation.NewFromConfig(cfg)

	_, err := cfnClient.CreateStack(ctx, &cloudformation.CreateStackInput{
		StackName:    aws.String(stackName),
		TemplateURL:  aws.String(templateURL),
		Capabilities: []cfnTypes.Capability{cfnTypes.CapabilityCapabilityNamedIam},
	})
	if err != nil {
		return fmt.Errorf("failed to create stack: %w", err)
	}

	return nil
}

func waitStackCreate(ctx context.Context, cfg aws.Config, stackName string) error {
	cfnClient := cloudformation.NewFromConfig(cfg)

	waiter := cloudformation.NewStackCreateCompleteWaiter(cfnClient)
	err := waiter.Wait(ctx, &cloudformation.DescribeStacksInput{
		StackName: aws.String(stackName),
	}, 30*time.Minute)
	if err != nil {
		return fmt.Errorf("failed to wait for stack creation: %w", err)
	}

	return nil
}

func deleteS3Bucket(ctx context.Context, cfg aws.Config, bucketName string) error {
	s3Client := s3.NewFromConfig(cfg)

	// List and delete all objects in the bucket
	listObjectsOutput, err := s3Client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		return fmt.Errorf("failed to list objects: %w", err)
	}

	// Delete all objects
	for _, obj := range listObjectsOutput.Contents {
		_, err := s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
			Bucket: aws.String(bucketName),
			Key:    obj.Key,
		})
		if err != nil {
			return fmt.Errorf("failed to delete object %s: %w", *obj.Key, err)
		}
	}

	// Delete the bucket
	_, err = s3Client.DeleteBucket(ctx, &s3.DeleteBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		return fmt.Errorf("failed to delete bucket: %w", err)
	}

	return nil
}
