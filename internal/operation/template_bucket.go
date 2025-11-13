package operation

import (
	"fmt"
	"math/rand/v2"
	"strings"
)

// generateTemplateBucketName generates a unique S3 bucket name for temporary template storage.
// The bucket name format: delstack-tpl-{stack}-{account:12}-{region:max14}-{random:4}
// Maximum length: 63 characters (S3 bucket name limit)
func generateTemplateBucketName(stackName, accountID, region string) string {
	// Generate random suffix to avoid bucket name collision (4 digits: 0000-9999)
	//nolint:gosec // G404: This is not cryptographically sensitive, just for bucket name uniqueness
	randomSuffix := fmt.Sprintf("%04d", rand.IntN(10000))

	// S3 bucket name must be lowercase, so convert stack name to lowercase and replace invalid characters
	sanitizedStackName := strings.ToLower(stackName)
	sanitizedStackName = strings.ReplaceAll(sanitizedStackName, "_", "-")

	// Truncate stack name to avoid exceeding S3 bucket name limit (63 chars)
	// Format: delstack-tpl-{stack}-{account:12}-{region:max14}-{random:4}
	// Calculation: 13 (prefix) + 17 (stack) + 1 + 12 + 1 + 14 + 1 + 4 = 63 chars
	// Region max length is 14 (e.g., ap-southeast-1, ap-northeast-1)
	if len(sanitizedStackName) > 17 {
		sanitizedStackName = sanitizedStackName[:17]
	}

	return fmt.Sprintf("delstack-tpl-%s-%s-%s-%s", sanitizedStackName, accountID, region, randomSuffix)
}
