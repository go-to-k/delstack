package operation

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
)

func TestCloudFormationStackOperator_removeDeletionPolicyFromTemplate(t *testing.T) {
	type args struct {
		template *string
	}

	type want struct {
		modifiedTemplate string
	}

	cases := []struct {
		name string
		args args
		want want
	}{
		{
			name: "remove deletion policy from yaml format",
			args: args{
				template: aws.String(`Resources:
  MyTopic:
    DeletionPolicy: Retain
    Properties:`),
			},
			want: want{
				modifiedTemplate: `Resources:
  MyTopic:
    Properties:`},
		},
		{
			name: "remove deletion policy from yaml format with double quotes",
			args: args{
				template: aws.String(`Resources:
  MyTopic:
    DeletionPolicy: "Retain"
    Properties:`),
			},
			want: want{
				modifiedTemplate: `Resources:
  MyTopic:
    Properties:`},
		},
		{
			name: "remove deletion policy from yaml format with single quotes",
			args: args{
				template: aws.String(`Resources:
  MyTopic:
    DeletionPolicy: 'Retain'
    Properties:`),
			},
			want: want{
				modifiedTemplate: `Resources:
  MyTopic:
    Properties:`},
		},
		{
			name: "remove deletion policy from yaml format with double quoted key",
			args: args{
				template: aws.String(`Resources:
  MyTopic:
    "DeletionPolicy": "Retain"
    Properties:`),
			},
			want: want{
				modifiedTemplate: `Resources:
  MyTopic:
    Properties:`},
		},
		{
			name: "remove deletion policy from yaml format with single quoted key",
			args: args{
				template: aws.String(`Resources:
  MyTopic:
    'DeletionPolicy': 'Retain'
    Properties:`),
			},
			want: want{
				modifiedTemplate: `Resources:
  MyTopic:
    Properties:`},
		},
		{
			name: "remove deletion policy from yaml block format",
			args: args{
				template: aws.String(`Resources:
  MyTopic:
    DeletionPolicy:
      Retain
    Properties:`),
			},
			want: want{
				modifiedTemplate: `Resources:
  MyTopic:
    Properties:`},
		},
		{
			name: "remove deletion policy from yaml block format with double quotes",
			args: args{
				template: aws.String(`Resources:
  MyTopic:
    "DeletionPolicy":
      "Retain"
    Properties:`),
			},
			want: want{
				modifiedTemplate: `Resources:
  MyTopic:
    Properties:`},
		},
		{
			name: "remove deletion policy from yaml block format with single quotes",
			args: args{
				template: aws.String(`Resources:
  MyTopic:
    'DeletionPolicy':
      'Retain'
    Properties:`),
			},
			want: want{
				modifiedTemplate: `Resources:
  MyTopic:
    Properties:`},
		},
		{
			name: "remove deletion policy from yaml format with deletion policy at last",
			args: args{
				template: aws.String(`Resources:
  MyTopic:
    Properties:
      Key1: Value1
    DeletionPolicy: Retain`),
			},
			want: want{
				modifiedTemplate: `Resources:
  MyTopic:
    Properties:
      Key1: Value1`},
		},
		{
			name: "remove deletion policy from yaml block format with deletion policy at last",
			args: args{
				template: aws.String(`Resources:
  MyTopic:
    Properties:
      Key1: Value1
    DeletionPolicy:
      Retain`),
			},
			want: want{
				modifiedTemplate: `Resources:
  MyTopic:
    Properties:
      Key1: Value1`},
		},
		{
			name: "remove deletion policy from json format with deletion policy at first",
			args: args{
				template: aws.String(`{
  "Resources": {
    "MyTopic": {
      "DeletionPolicy": "Retain",
      "Type":"AWS::SecretsManager::Secret"
    }
  }
}`),
			},
			want: want{
				modifiedTemplate: `{
  "Resources": {
    "MyTopic": {
      "Type":"AWS::SecretsManager::Secret"
    }
  }
}`},
		},
		{
			name: "remove deletion policy from json format with deletion policy at last",
			args: args{
				template: aws.String(`{
  "Resources": {
    "MyTopic": {
      "Type":"AWS::SecretsManager::Secret",
      "DeletionPolicy": "Retain"
    }
  }
}`),
			},
			want: want{
				modifiedTemplate: `{
  "Resources": {
    "MyTopic": {
      "Type":"AWS::SecretsManager::Secret"
    }
  }
}`},
		},
		{
			name: "remove deletion policy from json format with deletion policy in the middle",
			args: args{
				template: aws.String(`{
  "Resources": {
    "MyTopic": {
      "UpdatePolicy": "Retain",
      "DeletionPolicy": "Retain",
      "Type":"AWS::SecretsManager::Secret"
    }
  }
}`),
			},
			want: want{
				modifiedTemplate: `{
  "Resources": {
    "MyTopic": {
      "UpdatePolicy": "Retain",
      "Type":"AWS::SecretsManager::Secret"
    }
  }
}`},
		},
		{
			name: "remove deletion policy from minified json format with deletion policy at first",
			args: args{
				template: aws.String(`{"Resources":{"MyTopic":{"DeletionPolicy":"Retain","Type":"AWS::SecretsManager::Secret"}}}`),
			},
			want: want{
				modifiedTemplate: `{"Resources":{"MyTopic":{"Type":"AWS::SecretsManager::Secret"}}}`},
		},
		{
			name: "remove deletion policy from minified json format with deletion policy at last",
			args: args{
				template: aws.String(`{"Resources":{"MyTopic":{"Type":"AWS::SecretsManager::Secret","DeletionPolicy":"Retain"}}}`),
			},
			want: want{
				modifiedTemplate: `{"Resources":{"MyTopic":{"Type":"AWS::SecretsManager::Secret"}}}`},
		},
		{
			name: "remove deletion policy from minified json format with deletion policy in the middle",
			args: args{
				template: aws.String(`{"Resources":{"MyTopic":{"UpdatePolicy":"Retain","DeletionPolicy":"Retain","Type":"AWS::SecretsManager::Secret"}}}`),
			},
			want: want{
				modifiedTemplate: `{"Resources":{"MyTopic":{"UpdatePolicy":"Retain","Type":"AWS::SecretsManager::Secret"}}}`},
		},
		{
			name: "remove deletion policy RetainExceptOnCreate from yaml format",
			args: args{
				template: aws.String(`Resources:
  MyTopic:
    DeletionPolicy: RetainExceptOnCreate
    Properties:`),
			},
			want: want{
				modifiedTemplate: `Resources:
  MyTopic:
    Properties:`},
		},
		{
			name: "remove deletion policy RetainExceptOnCreate from json format at first",
			args: args{
				template: aws.String(`{
  "Resources": {
    "MyTopic": {
      "DeletionPolicy": "RetainExceptOnCreate",
      "Type":"AWS::SecretsManager::Secret"
    }
  }
}`),
			},
			want: want{
				modifiedTemplate: `{
  "Resources": {
    "MyTopic": {
      "Type":"AWS::SecretsManager::Secret"
    }
  }
}`},
		},
		{
			name: "remove deletion policy RetainExceptOnCreate from json format at last",
			args: args{
				template: aws.String(`{
  "Resources": {
    "MyRole": {
      "Type": "AWS::IAM::Role",
      "DeletionPolicy": "RetainExceptOnCreate"
    }
  }
}`),
			},
			want: want{
				modifiedTemplate: `{
  "Resources": {
    "MyRole": {
      "Type": "AWS::IAM::Role"
    }
  }
}`},
		},
		{
			name: "remove deletion policy RetainExceptOnCreate from minified json format",
			args: args{
				template: aws.String(`{"Resources":{"MyTopic":{"DeletionPolicy":"RetainExceptOnCreate","Type":"AWS::SecretsManager::Secret"}}}`),
			},
			want: want{
				modifiedTemplate: `{"Resources":{"MyTopic":{"Type":"AWS::SecretsManager::Secret"}}}`},
		},
		{
			name: "do not remove deletion policy Delete from json format",
			args: args{
				template: aws.String(`{
  "Resources": {
    "MyTopic": {
      "DeletionPolicy": "Delete",
      "Type":"AWS::SecretsManager::Secret"
    }
  }
}`),
			},
			want: want{
				modifiedTemplate: `{
  "Resources": {
    "MyTopic": {
      "DeletionPolicy": "Delete",
      "Type":"AWS::SecretsManager::Secret"
    }
  }
}`},
		},
		{
			name: "do not remove deletion policy Snapshot from json format",
			args: args{
				template: aws.String(`{
  "Resources": {
    "MyTopic": {
      "DeletionPolicy": "Snapshot",
      "Type":"AWS::RDS::DBInstance"
    }
  }
}`),
			},
			want: want{
				modifiedTemplate: `{
  "Resources": {
    "MyTopic": {
      "DeletionPolicy": "Snapshot",
      "Type":"AWS::RDS::DBInstance"
    }
  }
}`},
		},
		{
			name: "preserve yaml indentation with 2 spaces",
			args: args{
				template: aws.String(`AWSTemplateFormatVersion: '2010-09-09'
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
      TableName: test-table`),
			},
			want: want{
				modifiedTemplate: `AWSTemplateFormatVersion: '2010-09-09'
Resources:
  MyBucket:
    Type: AWS::S3::Bucket
    Properties:
      BucketName: test-bucket
  MyTable:
    Type: AWS::DynamoDB::Table
    Properties:
      TableName: test-table`},
		},
		{
			name: "preserve yaml indentation with 4 spaces",
			args: args{
				template: aws.String(`AWSTemplateFormatVersion: '2010-09-09'
Resources:
    MyBucket:
        Type: AWS::S3::Bucket
        DeletionPolicy: Retain
        Properties:
            BucketName: test-bucket`),
			},
			want: want{
				modifiedTemplate: `AWSTemplateFormatVersion: '2010-09-09'
Resources:
    MyBucket:
        Type: AWS::S3::Bucket
        Properties:
            BucketName: test-bucket`},
		},
		{
			name: "preserve json indentation and order",
			args: args{
				template: aws.String(`{
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
}`),
			},
			want: want{
				modifiedTemplate: `{
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
}`},
		},
		{
			name: "preserve json order - DeletionPolicy in middle",
			args: args{
				template: aws.String(`{
  "Resources": {
    "MyTopic": {
      "Type": "AWS::SNS::Topic",
      "DeletionPolicy": "Retain",
      "Properties": {
        "TopicName": "test-topic"
      }
    }
  }
}`),
			},
			want: want{
				modifiedTemplate: `{
  "Resources": {
    "MyTopic": {
      "Type": "AWS::SNS::Topic",
      "Properties": {
        "TopicName": "test-topic"
      }
    }
  }
}`},
		},
		{
			name: "handle trailing comma correctly when DeletionPolicy is last property",
			args: args{
				template: aws.String(`{
  "Resources": {
    "MyBucket": {
      "Type": "AWS::S3::Bucket",
      "Properties": {
        "BucketName": "test"
      },
      "DeletionPolicy": "Retain"
    }
  }
}`),
			},
			want: want{
				modifiedTemplate: `{
  "Resources": {
    "MyBucket": {
      "Type": "AWS::S3::Bucket",
      "Properties": {
        "BucketName": "test"
      }
    }
  }
}`},
		},
		{
			name: "preserve yaml order - multiple resources",
			args: args{
				template: aws.String(`Resources:
  FirstResource:
    Type: AWS::S3::Bucket
    DeletionPolicy: Retain
  SecondResource:
    Type: AWS::IAM::Role
    DeletionPolicy: RetainExceptOnCreate
  ThirdResource:
    Type: AWS::Lambda::Function`),
			},
			want: want{
				modifiedTemplate: `Resources:
  FirstResource:
    Type: AWS::S3::Bucket
  SecondResource:
    Type: AWS::IAM::Role
  ThirdResource:
    Type: AWS::Lambda::Function`},
		},
		{
			name: "minified json with escaped newline in string value",
			args: args{
				template: aws.String(`{"Resources":{"MyResource":{"Type":"AWS::EC2::Instance","Properties":{"UserData":"#!/bin/bash\necho \"Hello\""},"DeletionPolicy":"Retain"}}}`),
			},
			want: want{
				modifiedTemplate: `{"Resources":{"MyResource":{"Type":"AWS::EC2::Instance","Properties":{"UserData":"#!/bin/bash\necho \"Hello\""}}}}`},
		},
		{
			name: "formatted yaml with multiline string containing actual newlines",
			args: args{
				template: aws.String(`Resources:
  MyResource:
    Type: AWS::EC2::Instance
    DeletionPolicy: Retain
    Properties:
      UserData: |
        #!/bin/bash
        echo "Hello"
        echo "World"`),
			},
			want: want{
				modifiedTemplate: `Resources:
  MyResource:
    Type: AWS::EC2::Instance
    Properties:
      UserData: |
        #!/bin/bash
        echo "Hello"
        echo "World"`},
		},
		{
			name: "formatted json with multiline string containing actual newlines",
			args: args{
				template: aws.String(`{
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
}`),
			},
			want: want{
				modifiedTemplate: `{
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
}`},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			cloudformationStackOperator := NewCloudFormationStackOperator(aws.Config{}, nil, []string{})
			got := cloudformationStackOperator.removeDeletionPolicyFromTemplate(tt.args.template)
			if !reflect.DeepEqual(got, tt.want.modifiedTemplate) {
				t.Errorf("output = %#v, want %#v", got, tt.want.modifiedTemplate)
			}
		})
	}
}
