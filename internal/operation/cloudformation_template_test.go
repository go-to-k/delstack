package operation

import (
	"encoding/json"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"gopkg.in/yaml.v3"
)

type testCase struct {
	name          string
	template      string
	expectChanged bool
	checkFn       func(t *testing.T, result string)
}

func runTest(t *testing.T, tt testCase) {
	t.Helper()
	got, changed, err := removeDeletionPolicyFromTemplate(aws.String(tt.template))
	if err != nil {
		t.Fatalf("removeDeletionPolicyFromTemplate() error = %v", err)
	}
	if changed != tt.expectChanged {
		t.Errorf("changed = %v, want %v", changed, tt.expectChanged)
	}
	tt.checkFn(t, got)
}

func Test_removeDeletionPolicyFromTemplate_RemovesRetain(t *testing.T) {
	tests := []testCase{
		{
			name: "YAML",
			template: `AWSTemplateFormatVersion: '2010-09-09'
Resources:
  MyBucket:
    Type: AWS::S3::Bucket
    DeletionPolicy: Retain
    Properties:
      BucketName: test`,
			expectChanged: true,
			checkFn: func(t *testing.T, result string) {
				var data map[string]interface{}
				if err := yaml.Unmarshal([]byte(result), &data); err != nil {
					t.Fatalf("Failed to parse YAML: %v", err)
				}
				resources := data["Resources"].(map[string]interface{})
				bucket := resources["MyBucket"].(map[string]interface{})
				verifyDeletionPolicyRemoved(t, bucket, "AWS::S3::Bucket", map[string]interface{}{"BucketName": "test"})
			},
		},
		{
			name: "JSON",
			template: `{
  "Resources": {
    "MyBucket": {
      "Type": "AWS::S3::Bucket",
      "DeletionPolicy": "Retain",
      "Properties": {
        "BucketName": "test"
      }
    }
  }
}`,
			expectChanged: true,
			checkFn: func(t *testing.T, result string) {
				var data map[string]interface{}
				if err := json.Unmarshal([]byte(result), &data); err != nil {
					t.Fatalf("Failed to parse JSON: %v", err)
				}
				resources := data["Resources"].(map[string]interface{})
				bucket := resources["MyBucket"].(map[string]interface{})
				verifyDeletionPolicyRemoved(t, bucket, "AWS::S3::Bucket", map[string]interface{}{"BucketName": "test"})
			},
		},
		{
			name:          "JSON minified",
			template:      `{"Resources":{"MyBucket":{"Type":"AWS::S3::Bucket","DeletionPolicy":"Retain","Properties":{"BucketName":"test"}}}}`,
			expectChanged: true,
			checkFn: func(t *testing.T, result string) {
				var data map[string]interface{}
				if err := json.Unmarshal([]byte(result), &data); err != nil {
					t.Fatalf("Failed to parse JSON: %v", err)
				}
				resources := data["Resources"].(map[string]interface{})
				bucket := resources["MyBucket"].(map[string]interface{})
				verifyDeletionPolicyRemoved(t, bucket, "AWS::S3::Bucket", map[string]interface{}{"BucketName": "test"})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runTest(t, tt)
		})
	}
}

func Test_removeDeletionPolicyFromTemplate_RemovesRetainExceptOnCreate(t *testing.T) {
	tests := []testCase{
		{
			name: "YAML",
			template: `Resources:
  MyTable:
    Type: AWS::DynamoDB::Table
    DeletionPolicy: RetainExceptOnCreate
    Properties:
      TableName: test`,
			expectChanged: true,
			checkFn: func(t *testing.T, result string) {
				var data map[string]interface{}
				if err := yaml.Unmarshal([]byte(result), &data); err != nil {
					t.Fatalf("Failed to parse YAML: %v", err)
				}
				resources := data["Resources"].(map[string]interface{})
				table := resources["MyTable"].(map[string]interface{})
				verifyDeletionPolicyRemoved(t, table, "AWS::DynamoDB::Table", map[string]interface{}{"TableName": "test"})
			},
		},
		{
			name: "JSON",
			template: `{
  "Resources": {
    "MyTable": {
      "Type": "AWS::DynamoDB::Table",
      "DeletionPolicy": "RetainExceptOnCreate",
      "Properties": {
        "TableName": "test"
      }
    }
  }
}`,
			expectChanged: true,
			checkFn: func(t *testing.T, result string) {
				var data map[string]interface{}
				if err := json.Unmarshal([]byte(result), &data); err != nil {
					t.Fatalf("Failed to parse JSON: %v", err)
				}
				resources := data["Resources"].(map[string]interface{})
				table := resources["MyTable"].(map[string]interface{})
				verifyDeletionPolicyRemoved(t, table, "AWS::DynamoDB::Table", map[string]interface{}{"TableName": "test"})
			},
		},
		{
			name:          "JSON minified",
			template:      `{"Resources":{"MyTable":{"Type":"AWS::DynamoDB::Table","DeletionPolicy":"RetainExceptOnCreate","Properties":{"TableName":"test"}}}}`,
			expectChanged: true,
			checkFn: func(t *testing.T, result string) {
				var data map[string]interface{}
				if err := json.Unmarshal([]byte(result), &data); err != nil {
					t.Fatalf("Failed to parse JSON: %v", err)
				}
				resources := data["Resources"].(map[string]interface{})
				table := resources["MyTable"].(map[string]interface{})
				verifyDeletionPolicyRemoved(t, table, "AWS::DynamoDB::Table", map[string]interface{}{"TableName": "test"})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runTest(t, tt)
		})
	}
}

func Test_removeDeletionPolicyFromTemplate_KeepsSnapshot(t *testing.T) {
	tests := []testCase{
		{
			name: "YAML",
			template: `Resources:
  MyDB:
    Type: AWS::RDS::DBInstance
    DeletionPolicy: Snapshot
    Properties:
      Engine: mysql`,
			expectChanged: false,
			checkFn: func(t *testing.T, result string) {
				var data map[string]interface{}
				if err := yaml.Unmarshal([]byte(result), &data); err != nil {
					t.Fatalf("Failed to parse YAML: %v", err)
				}
				resources := data["Resources"].(map[string]interface{})
				db := resources["MyDB"].(map[string]interface{})
				verifyDeletionPolicyKept(t, db, "AWS::RDS::DBInstance", "Snapshot", map[string]interface{}{"Engine": "mysql"})
			},
		},
		{
			name: "JSON",
			template: `{
  "Resources": {
    "MyDB": {
      "Type": "AWS::RDS::DBInstance",
      "DeletionPolicy": "Snapshot",
      "Properties": {
        "Engine": "mysql"
      }
    }
  }
}`,
			expectChanged: false,
			checkFn: func(t *testing.T, result string) {
				var data map[string]interface{}
				if err := json.Unmarshal([]byte(result), &data); err != nil {
					t.Fatalf("Failed to parse JSON: %v", err)
				}
				resources := data["Resources"].(map[string]interface{})
				db := resources["MyDB"].(map[string]interface{})
				verifyDeletionPolicyKept(t, db, "AWS::RDS::DBInstance", "Snapshot", map[string]interface{}{"Engine": "mysql"})
			},
		},
		{
			name:          "JSON minified",
			template:      `{"Resources":{"MyDB":{"Type":"AWS::RDS::DBInstance","DeletionPolicy":"Snapshot","Properties":{"Engine":"mysql"}}}}`,
			expectChanged: false,
			checkFn: func(t *testing.T, result string) {
				var data map[string]interface{}
				if err := json.Unmarshal([]byte(result), &data); err != nil {
					t.Fatalf("Failed to parse JSON: %v", err)
				}
				resources := data["Resources"].(map[string]interface{})
				db := resources["MyDB"].(map[string]interface{})
				verifyDeletionPolicyKept(t, db, "AWS::RDS::DBInstance", "Snapshot", map[string]interface{}{"Engine": "mysql"})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runTest(t, tt)
		})
	}
}

func Test_removeDeletionPolicyFromTemplate_KeepsDelete(t *testing.T) {
	tests := []testCase{
		{
			name: "YAML",
			template: `Resources:
  MyBucket:
    Type: AWS::S3::Bucket
    DeletionPolicy: Delete
    Properties:
      BucketName: test`,
			expectChanged: false,
			checkFn: func(t *testing.T, result string) {
				var data map[string]interface{}
				if err := yaml.Unmarshal([]byte(result), &data); err != nil {
					t.Fatalf("Failed to parse YAML: %v", err)
				}
				resources := data["Resources"].(map[string]interface{})
				bucket := resources["MyBucket"].(map[string]interface{})
				verifyDeletionPolicyKept(t, bucket, "AWS::S3::Bucket", "Delete", map[string]interface{}{"BucketName": "test"})
			},
		},
		{
			name: "JSON",
			template: `{
  "Resources": {
    "MyBucket": {
      "Type": "AWS::S3::Bucket",
      "DeletionPolicy": "Delete",
      "Properties": {
        "BucketName": "test"
      }
    }
  }
}`,
			expectChanged: false,
			checkFn: func(t *testing.T, result string) {
				var data map[string]interface{}
				if err := json.Unmarshal([]byte(result), &data); err != nil {
					t.Fatalf("Failed to parse JSON: %v", err)
				}
				resources := data["Resources"].(map[string]interface{})
				bucket := resources["MyBucket"].(map[string]interface{})
				verifyDeletionPolicyKept(t, bucket, "AWS::S3::Bucket", "Delete", map[string]interface{}{"BucketName": "test"})
			},
		},
		{
			name:          "JSON minified",
			template:      `{"Resources":{"MyBucket":{"Type":"AWS::S3::Bucket","DeletionPolicy":"Delete","Properties":{"BucketName":"test"}}}}`,
			expectChanged: false,
			checkFn: func(t *testing.T, result string) {
				var data map[string]interface{}
				if err := json.Unmarshal([]byte(result), &data); err != nil {
					t.Fatalf("Failed to parse JSON: %v", err)
				}
				resources := data["Resources"].(map[string]interface{})
				bucket := resources["MyBucket"].(map[string]interface{})
				verifyDeletionPolicyKept(t, bucket, "AWS::S3::Bucket", "Delete", map[string]interface{}{"BucketName": "test"})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runTest(t, tt)
		})
	}
}

func Test_removeDeletionPolicyFromTemplate_KeepsDeletionPolicyInProperties(t *testing.T) {
	tests := []testCase{
		{
			name: "YAML",
			template: `Resources:
  MyCustomResource:
    Type: Custom::MyResource
    DeletionPolicy: Retain
    Properties:
      ServiceToken: arn:aws:lambda:us-east-1:123456789012:function:MyFunction
      Config:
        DeletionPolicy: KeepOnDelete
        Rules:
          - Id: Rule1
            DeletionPolicy: RemoveOnDelete`,
			expectChanged: true,
			checkFn: func(t *testing.T, result string) {
				var data map[string]interface{}
				if err := yaml.Unmarshal([]byte(result), &data); err != nil {
					t.Fatalf("Failed to parse YAML: %v", err)
				}
				resources := data["Resources"].(map[string]interface{})
				resource := resources["MyCustomResource"].(map[string]interface{})
				verifyCustomResourceProperties(t, resource)
			},
		},
		{
			name: "JSON",
			template: `{
  "Resources": {
    "MyCustomResource": {
      "Type": "Custom::MyResource",
      "DeletionPolicy": "RetainExceptOnCreate",
      "Properties": {
        "ServiceToken": "arn:aws:lambda:us-east-1:123456789012:function:MyFunction",
        "Config": {
          "DeletionPolicy": "KeepOnDelete",
          "Rules": [
            {
              "Id": "Rule1",
              "DeletionPolicy": "RemoveOnDelete"
            }
          ]
        }
      }
    }
  }
}`,
			expectChanged: true,
			checkFn: func(t *testing.T, result string) {
				var data map[string]interface{}
				if err := json.Unmarshal([]byte(result), &data); err != nil {
					t.Fatalf("Failed to parse JSON: %v", err)
				}
				resources := data["Resources"].(map[string]interface{})
				resource := resources["MyCustomResource"].(map[string]interface{})
				verifyCustomResourceProperties(t, resource)
			},
		},
		{
			name:          "JSON minified",
			template:      `{"Resources":{"MyCustomResource":{"Type":"Custom::MyResource","DeletionPolicy":"Retain","Properties":{"ServiceToken":"arn:aws:lambda:us-east-1:123456789012:function:MyFunction","Config":{"DeletionPolicy":"KeepOnDelete","Rules":[{"Id":"Rule1","DeletionPolicy":"RemoveOnDelete"}]}}}}}`,
			expectChanged: true,
			checkFn: func(t *testing.T, result string) {
				var data map[string]interface{}
				if err := json.Unmarshal([]byte(result), &data); err != nil {
					t.Fatalf("Failed to parse JSON: %v", err)
				}
				resources := data["Resources"].(map[string]interface{})
				resource := resources["MyCustomResource"].(map[string]interface{})
				verifyCustomResourceProperties(t, resource)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runTest(t, tt)
		})
	}
}

func Test_removeDeletionPolicyFromTemplate_NoDeletionPolicy(t *testing.T) {
	tests := []testCase{
		{
			name: "YAML",
			template: `Resources:
  MyQueue:
    Type: AWS::SQS::Queue
    Properties:
      QueueName: test`,
			expectChanged: false,
			checkFn: func(t *testing.T, result string) {
				var data map[string]interface{}
				if err := yaml.Unmarshal([]byte(result), &data); err != nil {
					t.Fatalf("Failed to parse YAML: %v", err)
				}
				resources := data["Resources"].(map[string]interface{})
				queue := resources["MyQueue"].(map[string]interface{})
				verifyNoDeletionPolicy(t, queue, "AWS::SQS::Queue", map[string]interface{}{"QueueName": "test"})
			},
		},
		{
			name: "JSON",
			template: `{
  "Resources": {
    "MyQueue": {
      "Type": "AWS::SQS::Queue",
      "Properties": {
        "QueueName": "test"
      }
    }
  }
}`,
			expectChanged: false,
			checkFn: func(t *testing.T, result string) {
				var data map[string]interface{}
				if err := json.Unmarshal([]byte(result), &data); err != nil {
					t.Fatalf("Failed to parse JSON: %v", err)
				}
				resources := data["Resources"].(map[string]interface{})
				queue := resources["MyQueue"].(map[string]interface{})
				verifyNoDeletionPolicy(t, queue, "AWS::SQS::Queue", map[string]interface{}{"QueueName": "test"})
			},
		},
		{
			name:          "JSON minified",
			template:      `{"Resources":{"MyQueue":{"Type":"AWS::SQS::Queue","Properties":{"QueueName":"test"}}}}`,
			expectChanged: false,
			checkFn: func(t *testing.T, result string) {
				var data map[string]interface{}
				if err := json.Unmarshal([]byte(result), &data); err != nil {
					t.Fatalf("Failed to parse JSON: %v", err)
				}
				resources := data["Resources"].(map[string]interface{})
				queue := resources["MyQueue"].(map[string]interface{})
				verifyNoDeletionPolicy(t, queue, "AWS::SQS::Queue", map[string]interface{}{"QueueName": "test"})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runTest(t, tt)
		})
	}
}

func Test_removeDeletionPolicyFromTemplate_CompleteTemplate(t *testing.T) {
	tests := []testCase{
		{
			name: "YAML",
			template: `AWSTemplateFormatVersion: '2010-09-09'
Description: Test template
Resources:
  MyBucket:
    Type: AWS::S3::Bucket
    DeletionPolicy: Retain
    Properties:
      BucketName: test-bucket
  MyTable:
    Type: AWS::DynamoDB::Table
    DeletionPolicy: RetainExceptOnCreate
    Properties:
      TableName: test-table
  MyDB:
    Type: AWS::RDS::DBInstance
    DeletionPolicy: Snapshot
    Properties:
      Engine: mysql
  MyQueue:
    Type: AWS::SQS::Queue
    Properties:
      QueueName: test-queue
Outputs:
  BucketName:
    Value: !Ref MyBucket
  TableName:
    Value: !Ref MyTable`,
			expectChanged: true,
			checkFn: func(t *testing.T, result string) {
				var data map[string]interface{}
				if err := yaml.Unmarshal([]byte(result), &data); err != nil {
					t.Fatalf("Failed to parse YAML: %v", err)
				}
				resources := data["Resources"].(map[string]interface{})

				// MyBucket - Retain should be removed
				bucket := resources["MyBucket"].(map[string]interface{})
				verifyDeletionPolicyRemoved(t, bucket, "AWS::S3::Bucket", map[string]interface{}{"BucketName": "test-bucket"})

				// MyTable - RetainExceptOnCreate should be removed
				table := resources["MyTable"].(map[string]interface{})
				verifyDeletionPolicyRemoved(t, table, "AWS::DynamoDB::Table", map[string]interface{}{"TableName": "test-table"})

				// MyDB - Snapshot should be kept
				db := resources["MyDB"].(map[string]interface{})
				verifyDeletionPolicyKept(t, db, "AWS::RDS::DBInstance", "Snapshot", map[string]interface{}{"Engine": "mysql"})

				// MyQueue - no DeletionPolicy
				queue := resources["MyQueue"].(map[string]interface{})
				verifyNoDeletionPolicy(t, queue, "AWS::SQS::Queue", map[string]interface{}{"QueueName": "test-queue"})
			},
		},
		{
			name: "JSON",
			template: `{
  "AWSTemplateFormatVersion": "2010-09-09",
  "Description": "Test template",
  "Resources": {
    "MyBucket": {
      "Type": "AWS::S3::Bucket",
      "DeletionPolicy": "Retain",
      "Properties": {
        "BucketName": "test-bucket"
      }
    },
    "MyTable": {
      "Type": "AWS::DynamoDB::Table",
      "DeletionPolicy": "RetainExceptOnCreate",
      "Properties": {
        "TableName": "test-table"
      }
    },
    "MyDB": {
      "Type": "AWS::RDS::DBInstance",
      "DeletionPolicy": "Snapshot",
      "Properties": {
        "Engine": "mysql"
      }
    },
    "MyQueue": {
      "Type": "AWS::SQS::Queue",
      "Properties": {
        "QueueName": "test-queue"
      }
    }
  },
  "Outputs": {
    "BucketName": {
      "Value": {"Ref": "MyBucket"}
    },
    "TableName": {
      "Value": {"Ref": "MyTable"}
    }
  }
}`,
			expectChanged: true,
			checkFn: func(t *testing.T, result string) {
				var data map[string]interface{}
				if err := json.Unmarshal([]byte(result), &data); err != nil {
					t.Fatalf("Failed to parse JSON: %v", err)
				}
				resources := data["Resources"].(map[string]interface{})

				// MyBucket - Retain should be removed
				bucket := resources["MyBucket"].(map[string]interface{})
				verifyDeletionPolicyRemoved(t, bucket, "AWS::S3::Bucket", map[string]interface{}{"BucketName": "test-bucket"})

				// MyTable - RetainExceptOnCreate should be removed
				table := resources["MyTable"].(map[string]interface{})
				verifyDeletionPolicyRemoved(t, table, "AWS::DynamoDB::Table", map[string]interface{}{"TableName": "test-table"})

				// MyDB - Snapshot should be kept
				db := resources["MyDB"].(map[string]interface{})
				verifyDeletionPolicyKept(t, db, "AWS::RDS::DBInstance", "Snapshot", map[string]interface{}{"Engine": "mysql"})

				// MyQueue - no DeletionPolicy
				queue := resources["MyQueue"].(map[string]interface{})
				verifyNoDeletionPolicy(t, queue, "AWS::SQS::Queue", map[string]interface{}{"QueueName": "test-queue"})
			},
		},
		{
			name:          "JSON minified",
			template:      `{"AWSTemplateFormatVersion":"2010-09-09","Description":"Test template","Resources":{"MyBucket":{"Type":"AWS::S3::Bucket","DeletionPolicy":"Retain","Properties":{"BucketName":"test-bucket"}},"MyTable":{"Type":"AWS::DynamoDB::Table","DeletionPolicy":"RetainExceptOnCreate","Properties":{"TableName":"test-table"}},"MyDB":{"Type":"AWS::RDS::DBInstance","DeletionPolicy":"Snapshot","Properties":{"Engine":"mysql"}},"MyQueue":{"Type":"AWS::SQS::Queue","Properties":{"QueueName":"test-queue"}}},"Outputs":{"BucketName":{"Value":{"Ref":"MyBucket"}},"TableName":{"Value":{"Ref":"MyTable"}}}}`,
			expectChanged: true,
			checkFn: func(t *testing.T, result string) {
				var data map[string]interface{}
				if err := json.Unmarshal([]byte(result), &data); err != nil {
					t.Fatalf("Failed to parse JSON: %v", err)
				}
				resources := data["Resources"].(map[string]interface{})

				// MyBucket - Retain should be removed
				bucket := resources["MyBucket"].(map[string]interface{})
				verifyDeletionPolicyRemoved(t, bucket, "AWS::S3::Bucket", map[string]interface{}{"BucketName": "test-bucket"})

				// MyTable - RetainExceptOnCreate should be removed
				table := resources["MyTable"].(map[string]interface{})
				verifyDeletionPolicyRemoved(t, table, "AWS::DynamoDB::Table", map[string]interface{}{"TableName": "test-table"})

				// MyDB - Snapshot should be kept
				db := resources["MyDB"].(map[string]interface{})
				verifyDeletionPolicyKept(t, db, "AWS::RDS::DBInstance", "Snapshot", map[string]interface{}{"Engine": "mysql"})

				// MyQueue - no DeletionPolicy
				queue := resources["MyQueue"].(map[string]interface{})
				verifyNoDeletionPolicy(t, queue, "AWS::SQS::Queue", map[string]interface{}{"QueueName": "test-queue"})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runTest(t, tt)
		})
	}
}

// Helper function to verify DeletionPolicy is removed and other properties are preserved
func verifyDeletionPolicyRemoved(t *testing.T, resource map[string]interface{}, expectedType string, expectedProps map[string]interface{}) {
	t.Helper()
	if _, policyExists := resource["DeletionPolicy"]; policyExists {
		t.Error("DeletionPolicy should be removed")
	}
	if resourceType, typeExists := resource["Type"]; !typeExists || resourceType != expectedType {
		t.Errorf("Type should be %s, got %v", expectedType, resourceType)
	}
	props, propsExist := resource["Properties"]
	if !propsExist {
		t.Error("Properties should be preserved")
		return
	}
	propsMap := props.(map[string]interface{})
	for key, expectedValue := range expectedProps {
		if actualValue, propExists := propsMap[key]; !propExists {
			t.Errorf("Properties[%s] should exist", key)
		} else if actualValue != expectedValue {
			t.Errorf("Properties[%s] = %v, want %v", key, actualValue, expectedValue)
		}
	}
}

// Helper function to verify DeletionPolicy is kept with expected value
func verifyDeletionPolicyKept(t *testing.T, resource map[string]interface{}, expectedType, expectedPolicy string, expectedProps map[string]interface{}) {
	t.Helper()
	policy, policyExists := resource["DeletionPolicy"]
	if !policyExists {
		t.Errorf("DeletionPolicy should exist")
		return
	}
	if policy != expectedPolicy {
		t.Errorf("DeletionPolicy = %v, want %s", policy, expectedPolicy)
	}
	if resourceType, typeExists := resource["Type"]; !typeExists || resourceType != expectedType {
		t.Errorf("Type should be %s, got %v", expectedType, resourceType)
	}
	props, propsExist := resource["Properties"]
	if !propsExist {
		t.Error("Properties should be preserved")
		return
	}
	propsMap := props.(map[string]interface{})
	for key, expectedValue := range expectedProps {
		if actualValue, propExists := propsMap[key]; !propExists {
			t.Errorf("Properties[%s] should exist", key)
		} else if actualValue != expectedValue {
			t.Errorf("Properties[%s] = %v, want %v", key, actualValue, expectedValue)
		}
	}
}

// Helper function to verify DeletionPolicy does not exist
//
//nolint:unparam
func verifyNoDeletionPolicy(t *testing.T, resource map[string]interface{}, expectedType string, expectedProps map[string]interface{}) {
	t.Helper()
	if _, policyExists := resource["DeletionPolicy"]; policyExists {
		t.Error("DeletionPolicy should not exist")
	}
	if resourceType, typeExists := resource["Type"]; !typeExists || resourceType != expectedType {
		t.Errorf("Type should be %s, got %v", expectedType, resourceType)
	}
	props, propsExist := resource["Properties"]
	if !propsExist {
		t.Error("Properties should be preserved")
		return
	}
	propsMap := props.(map[string]interface{})
	for key, expectedValue := range expectedProps {
		if actualValue, propExists := propsMap[key]; !propExists {
			t.Errorf("Properties[%s] should exist", key)
		} else if actualValue != expectedValue {
			t.Errorf("Properties[%s] = %v, want %v", key, actualValue, expectedValue)
		}
	}
}

// Helper function to verify custom resource properties with nested DeletionPolicy
func verifyCustomResourceProperties(t *testing.T, resource map[string]interface{}) {
	t.Helper()
	// Resource level DeletionPolicy should be removed
	verifyDeletionPolicyRemoved(t, resource, "Custom::MyResource", nil)
	// Properties level DeletionPolicy should be kept intact
	props := resource["Properties"].(map[string]interface{})
	// Check ServiceToken
	if props["ServiceToken"] != "arn:aws:lambda:us-east-1:123456789012:function:MyFunction" {
		t.Errorf("Properties.ServiceToken = %v, want arn:aws:lambda:us-east-1:123456789012:function:MyFunction", props["ServiceToken"])
	}
	config := props["Config"].(map[string]interface{})
	if config["DeletionPolicy"] != "KeepOnDelete" {
		t.Errorf("Properties.Config.DeletionPolicy = %v, want KeepOnDelete", config["DeletionPolicy"])
	}
	rules := config["Rules"].([]interface{})
	if len(rules) != 1 {
		t.Fatalf("Config.Rules should have 1 rule, got %d", len(rules))
	}
	rule := rules[0].(map[string]interface{})
	if rule["Id"] != "Rule1" {
		t.Errorf("Rule.Id = %v, want Rule1", rule["Id"])
	}
	if rule["DeletionPolicy"] != "RemoveOnDelete" {
		t.Errorf("Rule.DeletionPolicy = %v, want RemoveOnDelete", rule["DeletionPolicy"])
	}
}
