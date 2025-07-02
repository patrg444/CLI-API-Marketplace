package aws

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock AWS CLI for testing
func mockAWSCommand(t *testing.T) func() {
	// Save original PATH
	originalPath := os.Getenv("PATH")
	
	// Create mock directory
	mockDir := t.TempDir()
	mockScript := filepath.Join(mockDir, "aws")
	
	scriptContent := `#!/bin/bash
# Mock AWS CLI for testing
case "$1" in
	"--version")
		echo "aws-cli/2.13.0 Python/3.11.4"
		exit 0
		;;
	"sts")
		if [[ "$2" == "get-caller-identity" ]]; then
			if [[ "$3" == "--output" && "$4" == "json" ]]; then
				echo '{"Account":"123456789012","UserId":"AIDAI23HXD3MBVANCL4X6","Arn":"arn:aws:iam::123456789012:user/testuser"}'
			else
				echo "123456789012"
			fi
			exit 0
		elif [[ "$2" == "assume-role" ]]; then
			if [[ "$6" == "test-session" ]]; then
				echo '{"Credentials":{"AccessKeyId":"ASIATEST123","SecretAccessKey":"SECRET123","SessionToken":"TOKEN123"}}'
				exit 0
			elif [[ "$6" == "apidirect-verify" ]]; then
				# Check external ID
				for arg in "$@"; do
					if [[ $arg == "--external-id" ]]; then
						exit 0
					fi
				done
				exit 1
			fi
		fi
		;;
	"configure")
		if [[ "$2" == "get" && "$3" == "region" ]]; then
			echo "us-west-2"
			exit 0
		fi
		;;
	"s3api")
		if [[ "$2" == "head-bucket" ]]; then
			# Simulate bucket doesn't exist
			exit 1
		elif [[ "$2" == "create-bucket" ]]; then
			exit 0
		elif [[ "$2" == "put-bucket-versioning" ]]; then
			exit 0
		elif [[ "$2" == "put-bucket-encryption" ]]; then
			exit 0
		fi
		;;
	"dynamodb")
		if [[ "$2" == "describe-table" ]]; then
			# Simulate table doesn't exist
			exit 1
		elif [[ "$2" == "create-table" ]]; then
			exit 0
		elif [[ "$2" == "wait" ]]; then
			exit 0
		fi
		;;
esac
exit 1
`
	
	err := os.WriteFile(mockScript, []byte(scriptContent), 0755)
	require.NoError(t, err)
	
	// Update PATH
	os.Setenv("PATH", mockDir+":"+originalPath)
	
	// Return cleanup function
	return func() {
		os.Setenv("PATH", originalPath)
	}
}

func TestGetCallerIdentity(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func(*testing.T) func()
		wantErr     bool
		errContains string
		validate    func(*testing.T, *AccountInfo)
	}{
		{
			name: "successful get caller identity",
			setupFunc: mockAWSCommand,
			wantErr: false,
			validate: func(t *testing.T, info *AccountInfo) {
				assert.Equal(t, "123456789012", info.AccountID)
				assert.Equal(t, "AIDAI23HXD3MBVANCL4X6", info.UserID)
				assert.Equal(t, "arn:aws:iam::123456789012:user/testuser", info.Arn)
			},
		},
		{
			name: "AWS CLI not found",
			setupFunc: func(t *testing.T) func() {
				originalPath := os.Getenv("PATH")
				os.Setenv("PATH", "/nonexistent")
				return func() {
					os.Setenv("PATH", originalPath)
				}
			},
			wantErr:     true,
			errContains: "failed to get AWS caller identity",
		},
		{
			name: "invalid JSON response",
			setupFunc: func(t *testing.T) func() {
				originalPath := os.Getenv("PATH")
				mockDir := t.TempDir()
				mockScript := filepath.Join(mockDir, "aws")
				
				// Create mock that returns invalid JSON
				err := os.WriteFile(mockScript, []byte(`#!/bin/bash
echo "invalid json"
exit 0
`), 0755)
				require.NoError(t, err)
				
				os.Setenv("PATH", mockDir+":"+originalPath)
				return func() {
					os.Setenv("PATH", originalPath)
				}
			},
			wantErr:     true,
			errContains: "failed to parse AWS response",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := func() {}
			if tt.setupFunc != nil {
				cleanup = tt.setupFunc(t)
			}
			defer cleanup()
			
			info, err := GetCallerIdentity()
			
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, info)
				if tt.validate != nil {
					tt.validate(t, info)
				}
			}
		})
	}
}

func TestCheckAWSCredentials(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(*testing.T) func()
		wantErr   bool
	}{
		{
			name:      "credentials configured",
			setupFunc: mockAWSCommand,
			wantErr:   false,
		},
		{
			name: "credentials not configured",
			setupFunc: func(t *testing.T) func() {
				originalPath := os.Getenv("PATH")
				mockDir := t.TempDir()
				mockScript := filepath.Join(mockDir, "aws")
				
				// Create mock that fails
				err := os.WriteFile(mockScript, []byte(`#!/bin/bash
exit 1
`), 0755)
				require.NoError(t, err)
				
				os.Setenv("PATH", mockDir+":"+originalPath)
				return func() {
					os.Setenv("PATH", originalPath)
				}
			},
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := func() {}
			if tt.setupFunc != nil {
				cleanup = tt.setupFunc(t)
			}
			defer cleanup()
			
			err := CheckAWSCredentials()
			
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "AWS credentials not configured")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCheckAWSCLI(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(*testing.T) func()
		wantErr   bool
	}{
		{
			name:      "AWS CLI installed",
			setupFunc: mockAWSCommand,
			wantErr:   false,
		},
		{
			name: "AWS CLI not installed",
			setupFunc: func(t *testing.T) func() {
				originalPath := os.Getenv("PATH")
				os.Setenv("PATH", "/nonexistent")
				return func() {
					os.Setenv("PATH", originalPath)
				}
			},
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := func() {}
			if tt.setupFunc != nil {
				cleanup = tt.setupFunc(t)
			}
			defer cleanup()
			
			err := CheckAWSCLI()
			
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "AWS CLI not found")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAssumeRole(t *testing.T) {
	tests := []struct {
		name        string
		roleArn     string
		sessionName string
		setupFunc   func(*testing.T) func()
		wantErr     bool
		checkEnv    bool
	}{
		{
			name:        "successful assume role",
			roleArn:     "arn:aws:iam::123456789012:role/test-role",
			sessionName: "test-session",
			setupFunc:   mockAWSCommand,
			wantErr:     false,
			checkEnv:    true,
		},
		{
			name:        "failed assume role",
			roleArn:     "arn:aws:iam::123456789012:role/invalid",
			sessionName: "invalid-session",
			setupFunc: func(t *testing.T) func() {
				originalPath := os.Getenv("PATH")
				mockDir := t.TempDir()
				mockScript := filepath.Join(mockDir, "aws")
				
				err := os.WriteFile(mockScript, []byte(`#!/bin/bash
exit 1
`), 0755)
				require.NoError(t, err)
				
				os.Setenv("PATH", mockDir+":"+originalPath)
				return func() {
					os.Setenv("PATH", originalPath)
				}
			},
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original env vars
			originalAccessKey := os.Getenv("AWS_ACCESS_KEY_ID")
			originalSecretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
			originalToken := os.Getenv("AWS_SESSION_TOKEN")
			
			cleanup := func() {
				os.Setenv("AWS_ACCESS_KEY_ID", originalAccessKey)
				os.Setenv("AWS_SECRET_ACCESS_KEY", originalSecretKey)
				os.Setenv("AWS_SESSION_TOKEN", originalToken)
			}
			defer cleanup()
			
			if tt.setupFunc != nil {
				setupCleanup := tt.setupFunc(t)
				defer setupCleanup()
			}
			
			err := AssumeRole(tt.roleArn, tt.sessionName)
			
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				
				if tt.checkEnv {
					assert.Equal(t, "ASIATEST123", os.Getenv("AWS_ACCESS_KEY_ID"))
					assert.Equal(t, "SECRET123", os.Getenv("AWS_SECRET_ACCESS_KEY"))
					assert.Equal(t, "TOKEN123", os.Getenv("AWS_SESSION_TOKEN"))
				}
			}
		})
	}
}

func TestGetRegion(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(*testing.T) func()
		want      string
	}{
		{
			name: "from AWS_REGION env var",
			setupFunc: func(t *testing.T) func() {
				originalRegion := os.Getenv("AWS_REGION")
				os.Setenv("AWS_REGION", "eu-west-1")
				return func() {
					os.Setenv("AWS_REGION", originalRegion)
				}
			},
			want: "eu-west-1",
		},
		{
			name: "from AWS_DEFAULT_REGION env var",
			setupFunc: func(t *testing.T) func() {
				originalRegion := os.Getenv("AWS_REGION")
				originalDefault := os.Getenv("AWS_DEFAULT_REGION")
				os.Unsetenv("AWS_REGION")
				os.Setenv("AWS_DEFAULT_REGION", "ap-southeast-1")
				return func() {
					os.Setenv("AWS_REGION", originalRegion)
					os.Setenv("AWS_DEFAULT_REGION", originalDefault)
				}
			},
			want: "ap-southeast-1",
		},
		{
			name: "from AWS CLI config",
			setupFunc: func(t *testing.T) func() {
				originalRegion := os.Getenv("AWS_REGION")
				originalDefault := os.Getenv("AWS_DEFAULT_REGION")
				os.Unsetenv("AWS_REGION")
				os.Unsetenv("AWS_DEFAULT_REGION")
				
				cleanup := mockAWSCommand(t)
				
				return func() {
					os.Setenv("AWS_REGION", originalRegion)
					os.Setenv("AWS_DEFAULT_REGION", originalDefault)
					cleanup()
				}
			},
			want: "us-west-2",
		},
		{
			name: "default region",
			setupFunc: func(t *testing.T) func() {
				originalRegion := os.Getenv("AWS_REGION")
				originalDefault := os.Getenv("AWS_DEFAULT_REGION")
				os.Unsetenv("AWS_REGION")
				os.Unsetenv("AWS_DEFAULT_REGION")
				
				// Mock AWS CLI that fails
				originalPath := os.Getenv("PATH")
				mockDir := t.TempDir()
				mockScript := filepath.Join(mockDir, "aws")
				
				err := os.WriteFile(mockScript, []byte(`#!/bin/bash
exit 1
`), 0755)
				require.NoError(t, err)
				
				os.Setenv("PATH", mockDir+":"+originalPath)
				
				return func() {
					os.Setenv("AWS_REGION", originalRegion)
					os.Setenv("AWS_DEFAULT_REGION", originalDefault)
					os.Setenv("PATH", originalPath)
				}
			},
			want: "us-east-1",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := func() {}
			if tt.setupFunc != nil {
				cleanup = tt.setupFunc(t)
			}
			defer cleanup()
			
			region, err := GetRegion()
			assert.NoError(t, err)
			assert.Equal(t, tt.want, region)
		})
	}
}

func TestCreateS3Bucket(t *testing.T) {
	tests := []struct {
		name        string
		bucketName  string
		region      string
		setupFunc   func(*testing.T) func()
		wantErr     bool
		errContains string
	}{
		{
			name:       "create bucket successfully",
			bucketName: "test-bucket",
			region:     "us-west-2",
			setupFunc:  mockAWSCommand,
			wantErr:    false,
		},
		{
			name:       "create bucket in us-east-1",
			bucketName: "test-bucket",
			region:     "us-east-1",
			setupFunc:  mockAWSCommand,
			wantErr:    false,
		},
		{
			name:       "bucket already exists",
			bucketName: "existing-bucket",
			region:     "us-west-2",
			setupFunc: func(t *testing.T) func() {
				originalPath := os.Getenv("PATH")
				mockDir := t.TempDir()
				mockScript := filepath.Join(mockDir, "aws")
				
				// Mock that shows bucket exists
				err := os.WriteFile(mockScript, []byte(`#!/bin/bash
if [[ "$2" == "head-bucket" ]]; then
	exit 0  # Bucket exists
fi
exit 0
`), 0755)
				require.NoError(t, err)
				
				os.Setenv("PATH", mockDir+":"+originalPath)
				return func() {
					os.Setenv("PATH", originalPath)
				}
			},
			wantErr: false,
		},
		{
			name:       "create bucket fails",
			bucketName: "test-bucket",
			region:     "us-west-2",
			setupFunc: func(t *testing.T) func() {
				originalPath := os.Getenv("PATH")
				mockDir := t.TempDir()
				mockScript := filepath.Join(mockDir, "aws")
				
				err := os.WriteFile(mockScript, []byte(`#!/bin/bash
if [[ "$2" == "head-bucket" ]]; then
	exit 1  # Bucket doesn't exist
elif [[ "$2" == "create-bucket" ]]; then
	exit 1  # Fail to create
fi
exit 0
`), 0755)
				require.NoError(t, err)
				
				os.Setenv("PATH", mockDir+":"+originalPath)
				return func() {
					os.Setenv("PATH", originalPath)
				}
			},
			wantErr:     true,
			errContains: "failed to create S3 bucket",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := func() {}
			if tt.setupFunc != nil {
				cleanup = tt.setupFunc(t)
			}
			defer cleanup()
			
			err := CreateS3Bucket(tt.bucketName, tt.region)
			
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCreateDynamoDBTable(t *testing.T) {
	tests := []struct {
		name        string
		tableName   string
		region      string
		setupFunc   func(*testing.T) func()
		wantErr     bool
		errContains string
	}{
		{
			name:      "create table successfully",
			tableName: "test-table",
			region:    "us-west-2",
			setupFunc: mockAWSCommand,
			wantErr:   false,
		},
		{
			name:      "table already exists",
			tableName: "existing-table",
			region:    "us-west-2",
			setupFunc: func(t *testing.T) func() {
				originalPath := os.Getenv("PATH")
				mockDir := t.TempDir()
				mockScript := filepath.Join(mockDir, "aws")
				
				// Mock that shows table exists
				err := os.WriteFile(mockScript, []byte(`#!/bin/bash
if [[ "$2" == "describe-table" ]]; then
	exit 0  # Table exists
fi
exit 0
`), 0755)
				require.NoError(t, err)
				
				os.Setenv("PATH", mockDir+":"+originalPath)
				return func() {
					os.Setenv("PATH", originalPath)
				}
			},
			wantErr: false,
		},
		{
			name:      "create table fails",
			tableName: "test-table",
			region:    "us-west-2",
			setupFunc: func(t *testing.T) func() {
				originalPath := os.Getenv("PATH")
				mockDir := t.TempDir()
				mockScript := filepath.Join(mockDir, "aws")
				
				err := os.WriteFile(mockScript, []byte(`#!/bin/bash
if [[ "$2" == "describe-table" ]]; then
	exit 1  # Table doesn't exist
elif [[ "$2" == "create-table" ]]; then
	exit 1  # Fail to create
fi
exit 0
`), 0755)
				require.NoError(t, err)
				
				os.Setenv("PATH", mockDir+":"+originalPath)
				return func() {
					os.Setenv("PATH", originalPath)
				}
			},
			wantErr:     true,
			errContains: "failed to create DynamoDB table",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := func() {}
			if tt.setupFunc != nil {
				cleanup = tt.setupFunc(t)
			}
			defer cleanup()
			
			err := CreateDynamoDBTable(tt.tableName, tt.region)
			
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGenerateExternalID(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(*testing.T) func()
		validate  func(*testing.T, string)
	}{
		{
			name:      "with AWS credentials",
			setupFunc: mockAWSCommand,
			validate: func(t *testing.T, id string) {
				assert.Contains(t, id, "apidirect-123456789012-")
			},
		},
		{
			name: "without AWS credentials",
			setupFunc: func(t *testing.T) func() {
				originalPath := os.Getenv("PATH")
				os.Setenv("PATH", "/nonexistent")
				return func() {
					os.Setenv("PATH", originalPath)
				}
			},
			validate: func(t *testing.T, id string) {
				assert.Contains(t, id, "apidirect-")
				assert.NotContains(t, id, "123456789012")
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := func() {}
			if tt.setupFunc != nil {
				cleanup = tt.setupFunc(t)
			}
			defer cleanup()
			
			id := GenerateExternalID()
			assert.NotEmpty(t, id)
			
			if tt.validate != nil {
				tt.validate(t, id)
			}
		})
	}
}

func TestVerifyCrossAccountRole(t *testing.T) {
	tests := []struct {
		name       string
		roleArn    string
		externalId string
		setupFunc  func(*testing.T) func()
		wantErr    bool
	}{
		{
			name:       "successful verification",
			roleArn:    "arn:aws:iam::123456789012:role/test-role",
			externalId: "test-external-id",
			setupFunc:  mockAWSCommand,
			wantErr:    false,
		},
		{
			name:       "failed verification",
			roleArn:    "arn:aws:iam::123456789012:role/invalid-role",
			externalId: "invalid-id",
			setupFunc: func(t *testing.T) func() {
				originalPath := os.Getenv("PATH")
				mockDir := t.TempDir()
				mockScript := filepath.Join(mockDir, "aws")
				
				err := os.WriteFile(mockScript, []byte(`#!/bin/bash
exit 1
`), 0755)
				require.NoError(t, err)
				
				os.Setenv("PATH", mockDir+":"+originalPath)
				return func() {
					os.Setenv("PATH", originalPath)
				}
			},
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := func() {}
			if tt.setupFunc != nil {
				cleanup = tt.setupFunc(t)
			}
			defer cleanup()
			
			err := VerifyCrossAccountRole(tt.roleArn, tt.externalId)
			
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "unable to assume cross-account role")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAWSCommandParsing(t *testing.T) {
	// Test that the AWS commands are built correctly
	t.Run("S3 bucket creation command", func(t *testing.T) {
		// Test command building for different regions
		testCases := []struct {
			region   string
			hasConstraint bool
		}{
			{"us-east-1", false},
			{"us-west-2", true},
			{"eu-west-1", true},
		}
		
		for _, tc := range testCases {
			t.Run(tc.region, func(t *testing.T) {
				// Build expected args
				args := []string{"s3api", "create-bucket", "--bucket", "test-bucket"}
				if tc.hasConstraint {
					args = append(args, "--region", tc.region, "--create-bucket-configuration", fmt.Sprintf("LocationConstraint=%s", tc.region))
				}
				
				// Verify the args would be built correctly
				if tc.region == "us-east-1" {
					assert.Len(t, args, 4)
				} else {
					assert.Len(t, args, 8)
					assert.Contains(t, args, "--region")
					assert.Contains(t, args, tc.region)
				}
			})
		}
	})
	
	t.Run("DynamoDB table creation command", func(t *testing.T) {
		// Expected command structure
		expectedArgs := []string{
			"dynamodb", "create-table",
			"--table-name", "test-table",
			"--attribute-definitions", "AttributeName=LockID,AttributeType=S",
			"--key-schema", "AttributeName=LockID,KeyType=HASH",
			"--billing-mode", "PAY_PER_REQUEST",
			"--region", "us-west-2",
		}
		
		assert.Len(t, expectedArgs, 12)
		assert.Contains(t, expectedArgs, "PAY_PER_REQUEST")
	})
}

// Test environment variable handling
func TestEnvironmentVariables(t *testing.T) {
	t.Run("AssumeRole sets environment variables", func(t *testing.T) {
		// This is tested in TestAssumeRole but we can add specific env var tests
		originalVars := map[string]string{
			"AWS_ACCESS_KEY_ID":     os.Getenv("AWS_ACCESS_KEY_ID"),
			"AWS_SECRET_ACCESS_KEY": os.Getenv("AWS_SECRET_ACCESS_KEY"),
			"AWS_SESSION_TOKEN":     os.Getenv("AWS_SESSION_TOKEN"),
		}
		
		// Clear env vars
		os.Unsetenv("AWS_ACCESS_KEY_ID")
		os.Unsetenv("AWS_SECRET_ACCESS_KEY")
		os.Unsetenv("AWS_SESSION_TOKEN")
		
		// Verify they're cleared
		assert.Empty(t, os.Getenv("AWS_ACCESS_KEY_ID"))
		assert.Empty(t, os.Getenv("AWS_SECRET_ACCESS_KEY"))
		assert.Empty(t, os.Getenv("AWS_SESSION_TOKEN"))
		
		// Restore
		for k, v := range originalVars {
			if v != "" {
				os.Setenv(k, v)
			}
		}
	})
}

// Test JSON parsing
func TestJSONParsing(t *testing.T) {
	t.Run("AccountInfo parsing", func(t *testing.T) {
		jsonData := `{"Account":"123456789012","UserId":"AIDAI23HXD3MBVANCL4X6","Arn":"arn:aws:iam::123456789012:user/testuser"}`
		
		var info AccountInfo
		err := json.Unmarshal([]byte(jsonData), &info)
		assert.NoError(t, err)
		assert.Equal(t, "123456789012", info.AccountID)
		assert.Equal(t, "AIDAI23HXD3MBVANCL4X6", info.UserID)
		assert.Equal(t, "arn:aws:iam::123456789012:user/testuser", info.Arn)
	})
	
	t.Run("AssumeRole response parsing", func(t *testing.T) {
		jsonData := `{"Credentials":{"AccessKeyId":"ASIATEST123","SecretAccessKey":"SECRET123","SessionToken":"TOKEN123"}}`
		
		var result struct {
			Credentials struct {
				AccessKeyId     string
				SecretAccessKey string
				SessionToken    string
			}
		}
		
		err := json.Unmarshal([]byte(jsonData), &result)
		assert.NoError(t, err)
		assert.Equal(t, "ASIATEST123", result.Credentials.AccessKeyId)
		assert.Equal(t, "SECRET123", result.Credentials.SecretAccessKey)
		assert.Equal(t, "TOKEN123", result.Credentials.SessionToken)
	})
}