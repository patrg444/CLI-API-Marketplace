package aws

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// AccountInfo contains AWS account information
type AccountInfo struct {
	AccountID string `json:"Account"`
	Arn       string `json:"Arn"`
	UserID    string `json:"UserId"`
}

// GetCallerIdentity gets the current AWS caller identity
func GetCallerIdentity() (*AccountInfo, error) {
	cmd := exec.Command("aws", "sts", "get-caller-identity", "--output", "json")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get AWS caller identity: %w", err)
	}

	var info AccountInfo
	if err := json.Unmarshal(output, &info); err != nil {
		return nil, fmt.Errorf("failed to parse AWS response: %w", err)
	}

	return &info, nil
}

// CheckAWSCredentials verifies AWS credentials are configured
func CheckAWSCredentials() error {
	cmd := exec.Command("aws", "sts", "get-caller-identity")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("AWS credentials not configured. Please run 'aws configure' or set AWS environment variables")
	}
	return nil
}

// CheckAWSCLI verifies AWS CLI is installed
func CheckAWSCLI() error {
	cmd := exec.Command("aws", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("AWS CLI not found. Please install AWS CLI: https://aws.amazon.com/cli/")
	}
	return nil
}

// AssumeRole assumes an AWS IAM role
func AssumeRole(roleArn, sessionName string) error {
	cmd := exec.Command("aws", "sts", "assume-role",
		"--role-arn", roleArn,
		"--role-session-name", sessionName,
		"--output", "json")
	
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to assume role: %w", err)
	}

	var result struct {
		Credentials struct {
			AccessKeyId     string
			SecretAccessKey string
			SessionToken    string
		}
	}

	if err := json.Unmarshal(output, &result); err != nil {
		return fmt.Errorf("failed to parse assume role response: %w", err)
	}

	// Set environment variables
	os.Setenv("AWS_ACCESS_KEY_ID", result.Credentials.AccessKeyId)
	os.Setenv("AWS_SECRET_ACCESS_KEY", result.Credentials.SecretAccessKey)
	os.Setenv("AWS_SESSION_TOKEN", result.Credentials.SessionToken)

	return nil
}

// GetRegion gets the configured AWS region
func GetRegion() (string, error) {
	// Check environment variable first
	if region := os.Getenv("AWS_REGION"); region != "" {
		return region, nil
	}
	if region := os.Getenv("AWS_DEFAULT_REGION"); region != "" {
		return region, nil
	}

	// Try AWS CLI config
	cmd := exec.Command("aws", "configure", "get", "region")
	output, err := cmd.Output()
	if err != nil {
		return "us-east-1", nil // Default region
	}

	return strings.TrimSpace(string(output)), nil
}

// CreateS3Bucket creates an S3 bucket for Terraform state
func CreateS3Bucket(bucketName, region string) error {
	// Check if bucket exists
	checkCmd := exec.Command("aws", "s3api", "head-bucket", "--bucket", bucketName)
	if checkCmd.Run() == nil {
		// Bucket already exists
		return nil
	}

	// Create bucket
	args := []string{"s3api", "create-bucket", "--bucket", bucketName}
	if region != "us-east-1" {
		args = append(args, "--region", region, "--create-bucket-configuration", fmt.Sprintf("LocationConstraint=%s", region))
	}

	cmd := exec.Command("aws", args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create S3 bucket: %w", err)
	}

	// Enable versioning
	versionCmd := exec.Command("aws", "s3api", "put-bucket-versioning",
		"--bucket", bucketName,
		"--versioning-configuration", "Status=Enabled")
	if err := versionCmd.Run(); err != nil {
		return fmt.Errorf("failed to enable bucket versioning: %w", err)
	}

	// Enable encryption
	encryptCmd := exec.Command("aws", "s3api", "put-bucket-encryption",
		"--bucket", bucketName,
		"--server-side-encryption-configuration",
		`{"Rules": [{"ApplyServerSideEncryptionByDefault": {"SSEAlgorithm": "AES256"}}]}`)
	if err := encryptCmd.Run(); err != nil {
		return fmt.Errorf("failed to enable bucket encryption: %w", err)
	}

	return nil
}

// CreateDynamoDBTable creates a DynamoDB table for Terraform state locking
func CreateDynamoDBTable(tableName, region string) error {
	// Check if table exists
	checkCmd := exec.Command("aws", "dynamodb", "describe-table", "--table-name", tableName, "--region", region)
	if checkCmd.Run() == nil {
		// Table already exists
		return nil
	}

	// Create table
	cmd := exec.Command("aws", "dynamodb", "create-table",
		"--table-name", tableName,
		"--attribute-definitions", "AttributeName=LockID,AttributeType=S",
		"--key-schema", "AttributeName=LockID,KeyType=HASH",
		"--billing-mode", "PAY_PER_REQUEST",
		"--region", region)
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create DynamoDB table: %w", err)
	}

	// Wait for table to be active
	waitCmd := exec.Command("aws", "dynamodb", "wait", "table-exists",
		"--table-name", tableName,
		"--region", region)
	
	if err := waitCmd.Run(); err != nil {
		return fmt.Errorf("failed waiting for table to be ready: %w", err)
	}

	return nil
}

// GenerateExternalID generates a unique external ID for cross-account role
func GenerateExternalID() string {
	// In production, this should be a secure random string
	// For now, using a combination of account ID and timestamp
	info, err := GetCallerIdentity()
	if err != nil {
		return fmt.Sprintf("apidirect-%d", os.Getpid())
	}
	return fmt.Sprintf("apidirect-%s-%d", info.AccountID, os.Getpid())
}

// VerifyCrossAccountRole verifies that the cross-account role can be assumed
func VerifyCrossAccountRole(roleArn, externalId string) error {
	cmd := exec.Command("aws", "sts", "assume-role",
		"--role-arn", roleArn,
		"--role-session-name", "apidirect-verify",
		"--external-id", externalId,
		"--duration-seconds", "900",
		"--output", "json")
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("unable to assume cross-account role: %w", err)
	}
	
	return nil
}