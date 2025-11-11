package operation

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
)

func Test_removeDeletionPolicyFromTemplate_YAMLInline(t *testing.T) {
	tests := []struct {
		name     string
		template string
		want     string
	}{
		{
			name: "basic",
			template: `Resources:
  MyTopic:
    DeletionPolicy: Retain
    Properties:`,
			want: `Resources:
  MyTopic:
    Properties:`,
		},
		{
			name: "with double quotes on value",
			template: `Resources:
  MyTopic:
    DeletionPolicy: "Retain"
    Properties:`,
			want: `Resources:
  MyTopic:
    Properties:`,
		},
		{
			name: "with single quotes on value",
			template: `Resources:
  MyTopic:
    DeletionPolicy: 'Retain'
    Properties:`,
			want: `Resources:
  MyTopic:
    Properties:`,
		},
		{
			name: "with double quoted key",
			template: `Resources:
  MyTopic:
    "DeletionPolicy": "Retain"
    Properties:`,
			want: `Resources:
  MyTopic:
    Properties:`,
		},
		{
			name: "with single quoted key",
			template: `Resources:
  MyTopic:
    'DeletionPolicy': 'Retain'
    Properties:`,
			want: `Resources:
  MyTopic:
    Properties:`,
		},
		{
			name: "deletion policy at last",
			template: `Resources:
  MyTopic:
    Properties:
      Key1: Value1
    DeletionPolicy: Retain`,
			want: `Resources:
  MyTopic:
    Properties:
      Key1: Value1`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := removeDeletionPolicyFromTemplate(aws.String(tt.template))
			if got != tt.want {
				t.Errorf("removeDeletionPolicyFromTemplate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_removeDeletionPolicyFromTemplate_YAMLBlock(t *testing.T) {
	tests := []struct {
		name     string
		template string
		want     string
	}{
		{
			name: "basic",
			template: `Resources:
  MyTopic:
    DeletionPolicy:
      Retain
    Properties:`,
			want: `Resources:
  MyTopic:
    Properties:`,
		},
		{
			name: "with double quotes",
			template: `Resources:
  MyTopic:
    "DeletionPolicy":
      "Retain"
    Properties:`,
			want: `Resources:
  MyTopic:
    Properties:`,
		},
		{
			name: "with single quotes",
			template: `Resources:
  MyTopic:
    'DeletionPolicy':
      'Retain'
    Properties:`,
			want: `Resources:
  MyTopic:
    Properties:`,
		},
		{
			name: "deletion policy at last",
			template: `Resources:
  MyTopic:
    Properties:
      Key1: Value1
    DeletionPolicy:
      Retain`,
			want: `Resources:
  MyTopic:
    Properties:
      Key1: Value1`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := removeDeletionPolicyFromTemplate(aws.String(tt.template))
			if got != tt.want {
				t.Errorf("removeDeletionPolicyFromTemplate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_removeDeletionPolicyFromTemplate_JSONFormatted(t *testing.T) {
	tests := []struct {
		name     string
		template string
		want     string
	}{
		{
			name: "deletion policy at first",
			template: `{
  "Resources": {
    "MyTopic": {
      "DeletionPolicy": "Retain",
      "Type":"AWS::SecretsManager::Secret"
    }
  }
}`,
			want: `{
  "Resources": {
    "MyTopic": {
      "Type":"AWS::SecretsManager::Secret"
    }
  }
}`,
		},
		{
			name: "deletion policy at last",
			template: `{
  "Resources": {
    "MyTopic": {
      "Type":"AWS::SecretsManager::Secret",
      "DeletionPolicy": "Retain"
    }
  }
}`,
			want: `{
  "Resources": {
    "MyTopic": {
      "Type":"AWS::SecretsManager::Secret"
    }
  }
}`,
		},
		{
			name: "deletion policy in the middle",
			template: `{
  "Resources": {
    "MyTopic": {
      "UpdatePolicy": "Retain",
      "DeletionPolicy": "Retain",
      "Type":"AWS::SecretsManager::Secret"
    }
  }
}`,
			want: `{
  "Resources": {
    "MyTopic": {
      "UpdatePolicy": "Retain",
      "Type":"AWS::SecretsManager::Secret"
    }
  }
}`,
		},
		{
			name: "trailing comma handling",
			template: `{
  "Resources": {
    "MyBucket": {
      "Type": "AWS::S3::Bucket",
      "Properties": {
        "BucketName": "test"
      },
      "DeletionPolicy": "Retain"
    }
  }
}`,
			want: `{
  "Resources": {
    "MyBucket": {
      "Type": "AWS::S3::Bucket",
      "Properties": {
        "BucketName": "test"
      }
    }
  }
}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := removeDeletionPolicyFromTemplate(aws.String(tt.template))
			if got != tt.want {
				t.Errorf("removeDeletionPolicyFromTemplate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_removeDeletionPolicyFromTemplate_JSONMinified(t *testing.T) {
	tests := []struct {
		name     string
		template string
		want     string
	}{
		{
			name:     "deletion policy at first",
			template: `{"Resources":{"MyTopic":{"DeletionPolicy":"Retain","Type":"AWS::SecretsManager::Secret"}}}`,
			want:     `{"Resources":{"MyTopic":{"Type":"AWS::SecretsManager::Secret"}}}`,
		},
		{
			name:     "deletion policy at last",
			template: `{"Resources":{"MyTopic":{"Type":"AWS::SecretsManager::Secret","DeletionPolicy":"Retain"}}}`,
			want:     `{"Resources":{"MyTopic":{"Type":"AWS::SecretsManager::Secret"}}}`,
		},
		{
			name:     "deletion policy in the middle",
			template: `{"Resources":{"MyTopic":{"UpdatePolicy":"Retain","DeletionPolicy":"Retain","Type":"AWS::SecretsManager::Secret"}}}`,
			want:     `{"Resources":{"MyTopic":{"UpdatePolicy":"Retain","Type":"AWS::SecretsManager::Secret"}}}`,
		},
		{
			name:     "with escaped newline in string value",
			template: `{"Resources":{"MyResource":{"Type":"AWS::EC2::Instance","Properties":{"UserData":"#!/bin/bash\necho \"Hello\""},"DeletionPolicy":"Retain"}}}`,
			want:     `{"Resources":{"MyResource":{"Type":"AWS::EC2::Instance","Properties":{"UserData":"#!/bin/bash\necho \"Hello\""}}}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := removeDeletionPolicyFromTemplate(aws.String(tt.template))
			if got != tt.want {
				t.Errorf("removeDeletionPolicyFromTemplate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_removeDeletionPolicyFromTemplate_RetainExceptOnCreate(t *testing.T) {
	tests := []struct {
		name     string
		template string
		want     string
	}{
		{
			name: "yaml format",
			template: `Resources:
  MyTopic:
    DeletionPolicy: RetainExceptOnCreate
    Properties:`,
			want: `Resources:
  MyTopic:
    Properties:`,
		},
		{
			name: "json format at first",
			template: `{
  "Resources": {
    "MyTopic": {
      "DeletionPolicy": "RetainExceptOnCreate",
      "Type":"AWS::SecretsManager::Secret"
    }
  }
}`,
			want: `{
  "Resources": {
    "MyTopic": {
      "Type":"AWS::SecretsManager::Secret"
    }
  }
}`,
		},
		{
			name: "json format at last",
			template: `{
  "Resources": {
    "MyRole": {
      "Type": "AWS::IAM::Role",
      "DeletionPolicy": "RetainExceptOnCreate"
    }
  }
}`,
			want: `{
  "Resources": {
    "MyRole": {
      "Type": "AWS::IAM::Role"
    }
  }
}`,
		},
		{
			name:     "minified json format",
			template: `{"Resources":{"MyTopic":{"DeletionPolicy":"RetainExceptOnCreate","Type":"AWS::SecretsManager::Secret"}}}`,
			want:     `{"Resources":{"MyTopic":{"Type":"AWS::SecretsManager::Secret"}}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := removeDeletionPolicyFromTemplate(aws.String(tt.template))
			if got != tt.want {
				t.Errorf("removeDeletionPolicyFromTemplate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_removeDeletionPolicyFromTemplate_NegativeCases(t *testing.T) {
	tests := []struct {
		name     string
		template string
		want     string
	}{
		{
			name: "do not remove Delete",
			template: `{
  "Resources": {
    "MyTopic": {
      "DeletionPolicy": "Delete",
      "Type":"AWS::SecretsManager::Secret"
    }
  }
}`,
			want: `{
  "Resources": {
    "MyTopic": {
      "DeletionPolicy": "Delete",
      "Type":"AWS::SecretsManager::Secret"
    }
  }
}`,
		},
		{
			name: "do not remove Snapshot",
			template: `{
  "Resources": {
    "MyTopic": {
      "DeletionPolicy": "Snapshot",
      "Type":"AWS::RDS::DBInstance"
    }
  }
}`,
			want: `{
  "Resources": {
    "MyTopic": {
      "DeletionPolicy": "Snapshot",
      "Type":"AWS::RDS::DBInstance"
    }
  }
}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := removeDeletionPolicyFromTemplate(aws.String(tt.template))
			if got != tt.want {
				t.Errorf("removeDeletionPolicyFromTemplate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_removeDeletionPolicyFromTemplate_FormattingPreservation(t *testing.T) {
	tests := []struct {
		name     string
		template string
		want     string
	}{
		{
			name: "yaml indentation with 2 spaces",
			template: `AWSTemplateFormatVersion: '2010-09-09'
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
      TableName: test-table`,
			want: `AWSTemplateFormatVersion: '2010-09-09'
Resources:
  MyBucket:
    Type: AWS::S3::Bucket
    Properties:
      BucketName: test-bucket
  MyTable:
    Type: AWS::DynamoDB::Table
    Properties:
      TableName: test-table`,
		},
		{
			name: "yaml indentation with 4 spaces",
			template: `AWSTemplateFormatVersion: '2010-09-09'
Resources:
    MyBucket:
        Type: AWS::S3::Bucket
        DeletionPolicy: Retain
        Properties:
            BucketName: test-bucket`,
			want: `AWSTemplateFormatVersion: '2010-09-09'
Resources:
    MyBucket:
        Type: AWS::S3::Bucket
        Properties:
            BucketName: test-bucket`,
		},
		{
			name: "json indentation and order",
			template: `{
  "AWSTemplateFormatVersion": "2010-09-09",
  "Resources": {
    "MyBucket": {
      "Type": "AWS::S3::Bucket",
      "DeletionPolicy": "Retain",
      "Properties": {
        "BucketName": "test-bucket"
      }
    },
    "MyRole": {
      "Type": "AWS::IAM::Role",
      "DeletionPolicy": "RetainExceptOnCreate"
    }
  }
}`,
			want: `{
  "AWSTemplateFormatVersion": "2010-09-09",
  "Resources": {
    "MyBucket": {
      "Type": "AWS::S3::Bucket",
      "Properties": {
        "BucketName": "test-bucket"
      }
    },
    "MyRole": {
      "Type": "AWS::IAM::Role"
    }
  }
}`,
		},
		{
			name: "json order - DeletionPolicy in middle",
			template: `{
  "Resources": {
    "MyTopic": {
      "Type": "AWS::SNS::Topic",
      "DeletionPolicy": "Retain",
      "Properties": {
        "TopicName": "test-topic"
      }
    }
  }
}`,
			want: `{
  "Resources": {
    "MyTopic": {
      "Type": "AWS::SNS::Topic",
      "Properties": {
        "TopicName": "test-topic"
      }
    }
  }
}`,
		},
		{
			name: "yaml order - multiple resources",
			template: `Resources:
  FirstResource:
    Type: AWS::S3::Bucket
    DeletionPolicy: Retain
  SecondResource:
    Type: AWS::IAM::Role
    DeletionPolicy: RetainExceptOnCreate
  ThirdResource:
    Type: AWS::Lambda::Function`,
			want: `Resources:
  FirstResource:
    Type: AWS::S3::Bucket
  SecondResource:
    Type: AWS::IAM::Role
  ThirdResource:
    Type: AWS::Lambda::Function`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := removeDeletionPolicyFromTemplate(aws.String(tt.template))
			if got != tt.want {
				t.Errorf("removeDeletionPolicyFromTemplate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_removeDeletionPolicyFromTemplate_MultilineStrings(t *testing.T) {
	tests := []struct {
		name     string
		template string
		want     string
	}{
		{
			name: "yaml with multiline string",
			template: `Resources:
  MyResource:
    Type: AWS::EC2::Instance
    DeletionPolicy: Retain
    Properties:
      UserData: |
        #!/bin/bash
        echo "Hello"
        echo "World"`,
			want: `Resources:
  MyResource:
    Type: AWS::EC2::Instance
    Properties:
      UserData: |
        #!/bin/bash
        echo "Hello"
        echo "World"`,
		},
		{
			name: "json with multiline string",
			template: `{
  "Resources": {
    "MyResource": {
      "Type": "AWS::EC2::Instance",
      "DeletionPolicy": "Retain",
      "Properties": {
        "UserData": "#!/bin/bash
echo \"Hello\"
echo \"World\""
      }
    }
  }
}`,
			want: `{
  "Resources": {
    "MyResource": {
      "Type": "AWS::EC2::Instance",
      "Properties": {
        "UserData": "#!/bin/bash
echo \"Hello\"
echo \"World\""
      }
    }
  }
}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := removeDeletionPolicyFromTemplate(aws.String(tt.template))
			if got != tt.want {
				t.Errorf("removeDeletionPolicyFromTemplate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_removeDeletionPolicyFromTemplate_CompleteTemplates(t *testing.T) {
	tests := []struct {
		name     string
		template string
		want     string
	}{
		{
			name: "complete YAML template with Retain",
			template: `AWSTemplateFormatVersion: '2010-09-09'
Description: Sample template
Parameters:
  Env:
    Type: String
Resources:
  MyBucket:
    Type: AWS::S3::Bucket
    DeletionPolicy: Retain
    Properties:
      BucketName: !Ref Env
  MyTable:
    Type: AWS::DynamoDB::Table
    DeletionPolicy: RetainExceptOnCreate
    Properties:
      TableName: my-table
  MyDB:
    Type: AWS::RDS::DBInstance
    DeletionPolicy: Snapshot
    Properties:
      Engine: mysql
  MyQueue:
    Type: AWS::SQS::Queue
    Properties:
      QueueName: my-queue
Outputs:
  BucketName:
    Value: !Ref MyBucket`,
			want: `AWSTemplateFormatVersion: '2010-09-09'
Description: Sample template
Parameters:
  Env:
    Type: String
Resources:
  MyBucket:
    Type: AWS::S3::Bucket
    Properties:
      BucketName: !Ref Env
  MyTable:
    Type: AWS::DynamoDB::Table
    Properties:
      TableName: my-table
  MyDB:
    Type: AWS::RDS::DBInstance
    DeletionPolicy: Snapshot
    Properties:
      Engine: mysql
  MyQueue:
    Type: AWS::SQS::Queue
    Properties:
      QueueName: my-queue
Outputs:
  BucketName:
    Value: !Ref MyBucket`,
		},
		{
			name: "complete JSON template with Retain",
			template: `{
  "AWSTemplateFormatVersion": "2010-09-09",
  "Description": "Sample template",
  "Parameters": {
    "Env": {
      "Type": "String"
    }
  },
  "Resources": {
    "MyBucket": {
      "Type": "AWS::S3::Bucket",
      "DeletionPolicy": "Retain",
      "Properties": {
        "BucketName": {"Ref": "Env"}
      }
    },
    "MyTable": {
      "Type": "AWS::DynamoDB::Table",
      "DeletionPolicy": "RetainExceptOnCreate",
      "Properties": {
        "TableName": "my-table"
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
        "QueueName": "my-queue"
      }
    }
  },
  "Outputs": {
    "BucketName": {
      "Value": {"Ref": "MyBucket"}
    }
  }
}`,
			want: `{
  "AWSTemplateFormatVersion": "2010-09-09",
  "Description": "Sample template",
  "Parameters": {
    "Env": {
      "Type": "String"
    }
  },
  "Resources": {
    "MyBucket": {
      "Type": "AWS::S3::Bucket",
      "Properties": {
        "BucketName": {"Ref": "Env"}
      }
    },
    "MyTable": {
      "Type": "AWS::DynamoDB::Table",
      "Properties": {
        "TableName": "my-table"
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
        "QueueName": "my-queue"
      }
    }
  },
  "Outputs": {
    "BucketName": {
      "Value": {"Ref": "MyBucket"}
    }
  }
}`,
		},
		{
			name:     "complete minified JSON template with Retain",
			template: `{"AWSTemplateFormatVersion":"2010-09-09","Description":"Sample template","Parameters":{"Env":{"Type":"String"}},"Resources":{"MyBucket":{"Type":"AWS::S3::Bucket","DeletionPolicy":"Retain","Properties":{"BucketName":{"Ref":"Env"}}},"MyTable":{"Type":"AWS::DynamoDB::Table","DeletionPolicy":"RetainExceptOnCreate","Properties":{"TableName":"my-table"}},"MyDB":{"Type":"AWS::RDS::DBInstance","DeletionPolicy":"Snapshot","Properties":{"Engine":"mysql"}},"MyQueue":{"Type":"AWS::SQS::Queue","Properties":{"QueueName":"my-queue"}}},"Outputs":{"BucketName":{"Value":{"Ref":"MyBucket"}}}}`,
			want:     `{"AWSTemplateFormatVersion":"2010-09-09","Description":"Sample template","Parameters":{"Env":{"Type":"String"}},"Resources":{"MyBucket":{"Type":"AWS::S3::Bucket","Properties":{"BucketName":{"Ref":"Env"}}},"MyTable":{"Type":"AWS::DynamoDB::Table","Properties":{"TableName":"my-table"}},"MyDB":{"Type":"AWS::RDS::DBInstance","DeletionPolicy":"Snapshot","Properties":{"Engine":"mysql"}},"MyQueue":{"Type":"AWS::SQS::Queue","Properties":{"QueueName":"my-queue"}}},"Outputs":{"BucketName":{"Value":{"Ref":"MyBucket"}}}}`,
		},
		{
			name: "complete YAML template without Retain",
			template: `AWSTemplateFormatVersion: '2010-09-09'
Description: Sample template
Parameters:
  Env:
    Type: String
Resources:
  MyBucket:
    Type: AWS::S3::Bucket
    DeletionPolicy: Delete
    Properties:
      BucketName: !Ref Env
  MyTable:
    Type: AWS::DynamoDB::Table
    Properties:
      TableName: my-table
  MyDB:
    Type: AWS::RDS::DBInstance
    DeletionPolicy: Snapshot
    Properties:
      Engine: mysql
  MyQueue:
    Type: AWS::SQS::Queue
    Properties:
      QueueName: my-queue
Outputs:
  BucketName:
    Value: !Ref MyBucket`,
			want: `AWSTemplateFormatVersion: '2010-09-09'
Description: Sample template
Parameters:
  Env:
    Type: String
Resources:
  MyBucket:
    Type: AWS::S3::Bucket
    DeletionPolicy: Delete
    Properties:
      BucketName: !Ref Env
  MyTable:
    Type: AWS::DynamoDB::Table
    Properties:
      TableName: my-table
  MyDB:
    Type: AWS::RDS::DBInstance
    DeletionPolicy: Snapshot
    Properties:
      Engine: mysql
  MyQueue:
    Type: AWS::SQS::Queue
    Properties:
      QueueName: my-queue
Outputs:
  BucketName:
    Value: !Ref MyBucket`,
		},
		{
			name: "complete JSON template without Retain",
			template: `{
  "AWSTemplateFormatVersion": "2010-09-09",
  "Description": "Sample template",
  "Parameters": {
    "Env": {
      "Type": "String"
    }
  },
  "Resources": {
    "MyBucket": {
      "Type": "AWS::S3::Bucket",
      "DeletionPolicy": "Delete",
      "Properties": {
        "BucketName": {"Ref": "Env"}
      }
    },
    "MyTable": {
      "Type": "AWS::DynamoDB::Table",
      "Properties": {
        "TableName": "my-table"
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
        "QueueName": "my-queue"
      }
    }
  },
  "Outputs": {
    "BucketName": {
      "Value": {"Ref": "MyBucket"}
    }
  }
}`,
			want: `{
  "AWSTemplateFormatVersion": "2010-09-09",
  "Description": "Sample template",
  "Parameters": {
    "Env": {
      "Type": "String"
    }
  },
  "Resources": {
    "MyBucket": {
      "Type": "AWS::S3::Bucket",
      "DeletionPolicy": "Delete",
      "Properties": {
        "BucketName": {"Ref": "Env"}
      }
    },
    "MyTable": {
      "Type": "AWS::DynamoDB::Table",
      "Properties": {
        "TableName": "my-table"
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
        "QueueName": "my-queue"
      }
    }
  },
  "Outputs": {
    "BucketName": {
      "Value": {"Ref": "MyBucket"}
    }
  }
}`,
		},
		{
			name:     "complete minified JSON template without Retain",
			template: `{"AWSTemplateFormatVersion":"2010-09-09","Description":"Sample template","Parameters":{"Env":{"Type":"String"}},"Resources":{"MyBucket":{"Type":"AWS::S3::Bucket","DeletionPolicy":"Delete","Properties":{"BucketName":{"Ref":"Env"}}},"MyTable":{"Type":"AWS::DynamoDB::Table","Properties":{"TableName":"my-table"}},"MyDB":{"Type":"AWS::RDS::DBInstance","DeletionPolicy":"Snapshot","Properties":{"Engine":"mysql"}},"MyQueue":{"Type":"AWS::SQS::Queue","Properties":{"QueueName":"my-queue"}}},"Outputs":{"BucketName":{"Value":{"Ref":"MyBucket"}}}}`,
			want:     `{"AWSTemplateFormatVersion":"2010-09-09","Description":"Sample template","Parameters":{"Env":{"Type":"String"}},"Resources":{"MyBucket":{"Type":"AWS::S3::Bucket","DeletionPolicy":"Delete","Properties":{"BucketName":{"Ref":"Env"}}},"MyTable":{"Type":"AWS::DynamoDB::Table","Properties":{"TableName":"my-table"}},"MyDB":{"Type":"AWS::RDS::DBInstance","DeletionPolicy":"Snapshot","Properties":{"Engine":"mysql"}},"MyQueue":{"Type":"AWS::SQS::Queue","Properties":{"QueueName":"my-queue"}}},"Outputs":{"BucketName":{"Value":{"Ref":"MyBucket"}}}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := removeDeletionPolicyFromTemplate(aws.String(tt.template))
			if got != tt.want {
				t.Errorf("removeDeletionPolicyFromTemplate() = %v, want %v", got, tt.want)
			}
		})
	}
}
