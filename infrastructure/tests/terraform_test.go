package test

import (
	"testing"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/gruntwork-io/terratest/modules/aws"
	"github.com/gruntwork-io/terratest/modules/docker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTerraformAWSInfrastructure(t *testing.T) {
	t.Parallel()

	// Test AWS ECS infrastructure
	t.Run("validate ECS cluster configuration", func(t *testing.T) {
		terraformOptions := &terraform.Options{
			TerraformDir: "../terraform/aws/ecs",
			Vars: map[string]interface{}{
				"environment": "test",
				"cluster_name": "api-marketplace-test",
				"desired_count": 2,
				"cpu": "256",
				"memory": "512",
			},
		}

		// Clean up resources
		defer terraform.Destroy(t, terraformOptions)

		// Deploy infrastructure
		terraform.InitAndApply(t, terraformOptions)

		// Validate outputs
		clusterName := terraform.Output(t, terraformOptions, "cluster_name")
		assert.Equal(t, "api-marketplace-test", clusterName)

		serviceArn := terraform.Output(t, terraformOptions, "service_arn")
		assert.Contains(t, serviceArn, "arn:aws:ecs")

		// Validate ECS service is running
		region := terraform.Output(t, terraformOptions, "aws_region")
		serviceName := terraform.Output(t, terraformOptions, "service_name")
		
		runningCount := aws.GetEcsServiceRunningTaskCount(t, region, clusterName, serviceName)
		assert.Equal(t, 2, runningCount)
	})

	// Test RDS database configuration
	t.Run("validate RDS instance configuration", func(t *testing.T) {
		terraformOptions := &terraform.Options{
			TerraformDir: "../terraform/aws/rds",
			Vars: map[string]interface{}{
				"environment": "test",
				"db_name": "apimarketplace",
				"db_instance_class": "db.t3.micro",
				"allocated_storage": 20,
				"multi_az": false,
			},
		}

		defer terraform.Destroy(t, terraformOptions)
		terraform.InitAndApply(t, terraformOptions)

		// Validate database endpoint
		dbEndpoint := terraform.Output(t, terraformOptions, "db_endpoint")
		assert.NotEmpty(t, dbEndpoint)
		assert.Contains(t, dbEndpoint, ".rds.amazonaws.com")

		// Validate security settings
		publiclyAccessible := terraform.Output(t, terraformOptions, "publicly_accessible")
		assert.Equal(t, "false", publiclyAccessible)

		// Validate backup settings
		backupRetention := terraform.Output(t, terraformOptions, "backup_retention_period")
		assert.Equal(t, "7", backupRetention)
	})

	// Test API Gateway configuration
	t.Run("validate API Gateway setup", func(t *testing.T) {
		terraformOptions := &terraform.Options{
			TerraformDir: "../terraform/aws/api-gateway",
			Vars: map[string]interface{}{
				"environment": "test",
				"api_name": "api-marketplace-gateway",
				"stage_name": "test",
			},
		}

		defer terraform.Destroy(t, terraformOptions)
		terraform.InitAndApply(t, terraformOptions)

		// Validate API endpoint
		apiURL := terraform.Output(t, terraformOptions, "api_url")
		assert.Contains(t, apiURL, "execute-api")
		assert.Contains(t, apiURL, "amazonaws.com")

		// Validate throttling settings
		throttleRate := terraform.Output(t, terraformOptions, "throttle_rate_limit")
		assert.Equal(t, "1000", throttleRate)

		throttleBurst := terraform.Output(t, terraformOptions, "throttle_burst_limit")
		assert.Equal(t, "2000", throttleBurst)
	})
}

func TestTerraformSecurityConfiguration(t *testing.T) {
	t.Parallel()

	// Test security group configurations
	t.Run("validate security group rules", func(t *testing.T) {
		terraformOptions := &terraform.Options{
			TerraformDir: "../terraform/modules/security",
			Vars: map[string]interface{}{
				"environment": "test",
				"vpc_id": "vpc-test123",
			},
		}

		defer terraform.Destroy(t, terraformOptions)
		terraform.InitAndApply(t, terraformOptions)

		// Validate web security group
		webSgId := terraform.Output(t, terraformOptions, "web_security_group_id")
		assert.NotEmpty(t, webSgId)

		// Validate database security group
		dbSgId := terraform.Output(t, terraformOptions, "db_security_group_id")
		assert.NotEmpty(t, dbSgId)

		// Ensure database SG only allows traffic from web SG
		dbIngressRules := terraform.OutputList(t, terraformOptions, "db_ingress_rules")
		assert.Len(t, dbIngressRules, 1)
		assert.Contains(t, dbIngressRules[0], webSgId)
	})

	// Test IAM roles and policies
	t.Run("validate IAM configurations", func(t *testing.T) {
		terraformOptions := &terraform.Options{
			TerraformDir: "../terraform/modules/iam",
			Vars: map[string]interface{}{
				"environment": "test",
				"service_name": "api-marketplace",
			},
		}

		defer terraform.Destroy(t, terraformOptions)
		terraform.InitAndApply(t, terraformOptions)

		// Validate ECS task role
		taskRoleArn := terraform.Output(t, terraformOptions, "ecs_task_role_arn")
		assert.Contains(t, taskRoleArn, "arn:aws:iam")
		assert.Contains(t, taskRoleArn, "role/api-marketplace-test-task")

		// Validate Lambda execution role
		lambdaRoleArn := terraform.Output(t, terraformOptions, "lambda_role_arn")
		assert.Contains(t, lambdaRoleArn, "role/api-marketplace-test-lambda")

		// Validate least privilege policies
		taskPolicyDocument := terraform.Output(t, terraformOptions, "task_policy_document")
		assert.Contains(t, taskPolicyDocument, "s3:GetObject")
		assert.NotContains(t, taskPolicyDocument, "s3:*") // No wildcard permissions
	})
}

func TestTerraformNetworking(t *testing.T) {
	t.Parallel()

	// Test VPC configuration
	t.Run("validate VPC and subnet configuration", func(t *testing.T) {
		terraformOptions := &terraform.Options{
			TerraformDir: "../terraform/modules/networking",
			Vars: map[string]interface{}{
				"environment": "test",
				"vpc_cidr": "10.0.0.0/16",
				"availability_zones": []string{"us-east-1a", "us-east-1b"},
			},
		}

		defer terraform.Destroy(t, terraformOptions)
		terraform.InitAndApply(t, terraformOptions)

		// Validate VPC
		vpcId := terraform.Output(t, terraformOptions, "vpc_id")
		assert.NotEmpty(t, vpcId)

		// Validate public subnets
		publicSubnets := terraform.OutputList(t, terraformOptions, "public_subnet_ids")
		assert.Len(t, publicSubnets, 2)

		// Validate private subnets
		privateSubnets := terraform.OutputList(t, terraformOptions, "private_subnet_ids")
		assert.Len(t, privateSubnets, 2)

		// Validate NAT Gateway exists
		natGatewayId := terraform.Output(t, terraformOptions, "nat_gateway_id")
		assert.NotEmpty(t, natGatewayId)
	})

	// Test load balancer configuration
	t.Run("validate ALB configuration", func(t *testing.T) {
		terraformOptions := &terraform.Options{
			TerraformDir: "../terraform/modules/alb",
			Vars: map[string]interface{}{
				"environment": "test",
				"alb_name": "api-marketplace-test",
				"certificate_arn": "arn:aws:acm:us-east-1:123456789012:certificate/test",
			},
		}

		defer terraform.Destroy(t, terraformOptions)
		terraform.InitAndApply(t, terraformOptions)

		// Validate ALB DNS
		albDns := terraform.Output(t, terraformOptions, "alb_dns_name")
		assert.Contains(t, albDns, ".elb.amazonaws.com")

		// Validate HTTPS listener
		httpsListenerArn := terraform.Output(t, terraformOptions, "https_listener_arn")
		assert.Contains(t, httpsListenerArn, "arn:aws:elasticloadbalancing")

		// Validate security headers
		securityHeaders := terraform.OutputMap(t, terraformOptions, "security_headers")
		assert.Equal(t, "max-age=31536000", securityHeaders["Strict-Transport-Security"])
		assert.Equal(t, "nosniff", securityHeaders["X-Content-Type-Options"])
	})
}

func TestTerraformMonitoring(t *testing.T) {
	t.Parallel()

	// Test CloudWatch configuration
	t.Run("validate CloudWatch alarms", func(t *testing.T) {
		terraformOptions := &terraform.Options{
			TerraformDir: "../terraform/modules/monitoring",
			Vars: map[string]interface{}{
				"environment": "test",
				"alarm_email": "alerts@example.com",
			},
		}

		defer terraform.Destroy(t, terraformOptions)
		terraform.InitAndApply(t, terraformOptions)

		// Validate CPU alarm
		cpuAlarmName := terraform.Output(t, terraformOptions, "cpu_alarm_name")
		assert.Contains(t, cpuAlarmName, "high-cpu")

		// Validate memory alarm
		memoryAlarmName := terraform.Output(t, terraformOptions, "memory_alarm_name")
		assert.Contains(t, memoryAlarmName, "high-memory")

		// Validate error rate alarm
		errorAlarmName := terraform.Output(t, terraformOptions, "error_rate_alarm_name")
		assert.Contains(t, errorAlarmName, "high-error-rate")

		// Validate SNS topic
		snsTopicArn := terraform.Output(t, terraformOptions, "sns_topic_arn")
		assert.Contains(t, snsTopicArn, "arn:aws:sns")
	})

	// Test log configuration
	t.Run("validate CloudWatch logs", func(t *testing.T) {
		terraformOptions := &terraform.Options{
			TerraformDir: "../terraform/modules/logging",
			Vars: map[string]interface{}{
				"environment": "test",
				"retention_days": 30,
			},
		}

		defer terraform.Destroy(t, terraformOptions)
		terraform.InitAndApply(t, terraformOptions)

		// Validate log groups
		appLogGroup := terraform.Output(t, terraformOptions, "app_log_group")
		assert.Equal(t, "/ecs/api-marketplace-test", appLogGroup)

		// Validate log retention
		retentionDays := terraform.Output(t, terraformOptions, "retention_in_days")
		assert.Equal(t, "30", retentionDays)

		// Validate log stream prefix
		logStreamPrefix := terraform.Output(t, terraformOptions, "log_stream_prefix")
		assert.Equal(t, "ecs", logStreamPrefix)
	})
}

func TestTerraformBackup(t *testing.T) {
	t.Parallel()

	// Test backup configuration
	t.Run("validate AWS Backup configuration", func(t *testing.T) {
		terraformOptions := &terraform.Options{
			TerraformDir: "../terraform/modules/backup",
			Vars: map[string]interface{}{
				"environment": "test",
				"backup_retention_days": 30,
				"backup_schedule": "cron(0 3 * * ? *)", // 3 AM daily
			},
		}

		defer terraform.Destroy(t, terraformOptions)
		terraform.InitAndApply(t, terraformOptions)

		// Validate backup vault
		backupVaultName := terraform.Output(t, terraformOptions, "backup_vault_name")
		assert.Equal(t, "api-marketplace-test-vault", backupVaultName)

		// Validate backup plan
		backupPlanId := terraform.Output(t, terraformOptions, "backup_plan_id")
		assert.NotEmpty(t, backupPlanId)

		// Validate backup selection
		backupSelectionId := terraform.Output(t, terraformOptions, "backup_selection_id")
		assert.NotEmpty(t, backupSelectionId)
	})
}

func TestTerraformCostOptimization(t *testing.T) {
	t.Parallel()

	// Test auto-scaling configuration
	t.Run("validate auto-scaling policies", func(t *testing.T) {
		terraformOptions := &terraform.Options{
			TerraformDir: "../terraform/modules/autoscaling",
			Vars: map[string]interface{}{
				"environment": "test",
				"min_capacity": 1,
				"max_capacity": 10,
				"target_cpu": 70,
				"target_memory": 80,
			},
		}

		defer terraform.Destroy(t, terraformOptions)
		terraform.InitAndApply(t, terraformOptions)

		// Validate scaling target
		scalingTargetId := terraform.Output(t, terraformOptions, "scaling_target_id")
		assert.NotEmpty(t, scalingTargetId)

		// Validate CPU scaling policy
		cpuPolicyName := terraform.Output(t, terraformOptions, "cpu_scaling_policy_name")
		assert.Contains(t, cpuPolicyName, "cpu-scaling")

		// Validate memory scaling policy
		memoryPolicyName := terraform.Output(t, terraformOptions, "memory_scaling_policy_name")
		assert.Contains(t, memoryPolicyName, "memory-scaling")
	})

	// Test cost allocation tags
	t.Run("validate cost allocation tags", func(t *testing.T) {
		terraformOptions := &terraform.Options{
			TerraformDir: "../terraform/modules/tags",
			Vars: map[string]interface{}{
				"environment": "test",
				"project": "api-marketplace",
				"cost_center": "engineering",
			},
		}

		terraform.Init(t, terraformOptions)

		// Validate required tags in all resources
		planOutput := terraform.Plan(t, terraformOptions)
		assert.Contains(t, planOutput, "Environment")
		assert.Contains(t, planOutput, "Project")
		assert.Contains(t, planOutput, "CostCenter")
		assert.Contains(t, planOutput, "ManagedBy")
	})
}

// Helper function to validate Terraform syntax
func TestTerraformSyntax(t *testing.T) {
	// Test all Terraform modules for syntax errors
	modules := []string{
		"../terraform/aws/ecs",
		"../terraform/aws/rds",
		"../terraform/aws/api-gateway",
		"../terraform/modules/security",
		"../terraform/modules/networking",
		"../terraform/modules/monitoring",
	}

	for _, module := range modules {
		t.Run("validate syntax for "+module, func(t *testing.T) {
			terraformOptions := &terraform.Options{
				TerraformDir: module,
				Upgrade:      true,
			}

			// Validate Terraform configuration
			terraform.Init(t, terraformOptions)
			terraform.Validate(t, terraformOptions)
		})
	}
}