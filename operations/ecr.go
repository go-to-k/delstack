package operations

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/go-to-k/delstack/client"
)

func DeleteECRs(config aws.Config, resources []types.StackResourceSummary) error {
	// TODO: Concurrency Delete
	ecrClient := client.NewECR(config)
	for _, repository := range resources {
		err := DeleteECR(ecrClient, *repository.PhysicalResourceId)
		if err != nil {
			return err
		}
	}
	return nil
}

func DeleteECR(ecrClient *client.ECR, repositoryName string) error {
	if err := ecrClient.DeleteRepository(repositoryName); err != nil {
		return err
	}

	return nil
}
