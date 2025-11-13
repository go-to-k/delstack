package operation

import (
	"strings"
	"testing"
)

func TestGenerateTemplateBucketName(t *testing.T) {
	type args struct {
		stackName string
		accountID string
		region    string
	}

	cases := []struct {
		name        string
		args        args
		wantPrefix  string
		validations []func(string) bool
	}{
		{
			name: "normal stack name",
			args: args{
				stackName: "MyTestStack",
				accountID: "123456789012",
				region:    "us-east-1",
			},
			wantPrefix: "delstack-tpl-myteststack-",
			validations: []func(string) bool{
				func(name string) bool { return strings.HasPrefix(name, "delstack-tpl-myteststack-") },
				func(name string) bool { return len(name) <= 63 },
				func(name string) bool { return name == strings.ToLower(name) },
				func(name string) bool { return strings.Contains(name, "123456789012") },
				func(name string) bool { return strings.Contains(name, "us-east-1") },
			},
		},
		{
			name: "long stack name (truncated to 17 chars)",
			args: args{
				stackName: "MyVeryLongProductionStackName",
				accountID: "987654321098",
				region:    "ap-northeast-1",
			},
			wantPrefix: "delstack-tpl-myverylongproduct-",
			validations: []func(string) bool{
				func(name string) bool { return strings.HasPrefix(name, "delstack-tpl-myverylongproduct-") },
				func(name string) bool { return len(name) <= 63 },
				func(name string) bool {
					// Extract stack name part
					parts := strings.Split(name, "-")
					stackParts := []string{}
					for i := 2; i < len(parts); i++ {
						if len(parts[i]) == 12 { // Found account ID
							break
						}
						stackParts = append(stackParts, parts[i])
					}
					stackName := strings.Join(stackParts, "-")
					return len(stackName) == 17
				},
				func(name string) bool { return strings.Contains(name, "987654321098") },
				func(name string) bool { return strings.Contains(name, "ap-northeast-1") },
			},
		},
		{
			name: "stack name with underscores",
			args: args{
				stackName: "My_Test_Stack",
				accountID: "111222333444",
				region:    "eu-west-1",
			},
			wantPrefix: "delstack-tpl-my-test-stack-",
			validations: []func(string) bool{
				func(name string) bool { return strings.HasPrefix(name, "delstack-tpl-my-test-stack-") },
				func(name string) bool { return !strings.Contains(name, "_") }, // underscores replaced
				func(name string) bool { return len(name) <= 63 },
				func(name string) bool { return strings.Contains(name, "111222333444") },
				func(name string) bool { return strings.Contains(name, "eu-west-1") },
			},
		},
		{
			name: "stack name with uppercase",
			args: args{
				stackName: "PRODUCTION-STACK",
				accountID: "555666777888",
				region:    "us-west-2",
			},
			wantPrefix: "delstack-tpl-production-stack-",
			validations: []func(string) bool{
				func(name string) bool { return name == strings.ToLower(name) }, // all lowercase
				func(name string) bool { return len(name) <= 63 },
				func(name string) bool { return strings.Contains(name, "555666777888") },
				func(name string) bool { return strings.Contains(name, "us-west-2") },
			},
		},
		{
			name: "longest region name",
			args: args{
				stackName: "TestStack",
				accountID: "999888777666",
				region:    "ap-southeast-1", // 14 chars (max)
			},
			wantPrefix: "delstack-tpl-teststack-",
			validations: []func(string) bool{
				func(name string) bool { return len(name) <= 63 },
				func(name string) bool { return strings.Contains(name, "ap-southeast-1") },
				func(name string) bool { return strings.Contains(name, "999888777666") },
			},
		},
		{
			name: "verify random suffix format (4 digits)",
			args: args{
				stackName: "Test",
				accountID: "123456789012",
				region:    "us-east-1",
			},
			wantPrefix: "delstack-tpl-test-",
			validations: []func(string) bool{
				func(name string) bool {
					// Extract last part (should be 4 digit random number)
					parts := strings.Split(name, "-")
					lastPart := parts[len(parts)-1]
					return len(lastPart) == 4 && isNumeric(lastPart)
				},
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			bucketName := generateTemplateBucketName(tt.args.stackName, tt.args.accountID, tt.args.region)

			if !strings.HasPrefix(bucketName, tt.wantPrefix) {
				t.Errorf("bucket name = %s, want prefix %s", bucketName, tt.wantPrefix)
			}

			for i, validate := range tt.validations {
				if !validate(bucketName) {
					t.Errorf("validation %d failed for bucket name: %s", i, bucketName)
				}
			}
		})
	}
}

func isNumeric(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}
