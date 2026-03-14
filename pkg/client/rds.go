//go:generate mockgen -source=$GOFILE -destination=rds_mock.go -package=$GOPACKAGE -write_package_comment=false
package client

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
)

type IRDS interface {
	CheckDBInstanceDeletionProtection(ctx context.Context, dbInstanceId *string) (bool, error)
	DisableDBInstanceDeletionProtection(ctx context.Context, dbInstanceId *string) error
	CheckDBClusterDeletionProtection(ctx context.Context, dbClusterId *string) (bool, error)
	DisableDBClusterDeletionProtection(ctx context.Context, dbClusterId *string) error
}

var _ IRDS = (*RDS)(nil)

type RDS struct {
	client *rds.Client
}

func NewRDS(client *rds.Client) *RDS {
	return &RDS{
		client: client,
	}
}

func (r *RDS) CheckDBInstanceDeletionProtection(ctx context.Context, dbInstanceId *string) (bool, error) {
	input := &rds.DescribeDBInstancesInput{
		DBInstanceIdentifier: dbInstanceId,
	}

	output, err := r.client.DescribeDBInstances(ctx, input)
	if err != nil {
		return false, &ClientError{
			ResourceName: dbInstanceId,
			Err:          err,
		}
	}

	if len(output.DBInstances) == 0 {
		return false, nil
	}

	return aws.ToBool(output.DBInstances[0].DeletionProtection), nil
}

func (r *RDS) DisableDBInstanceDeletionProtection(ctx context.Context, dbInstanceId *string) error {
	input := &rds.ModifyDBInstanceInput{
		DBInstanceIdentifier: dbInstanceId,
		DeletionProtection:   aws.Bool(false),
	}

	_, err := r.client.ModifyDBInstance(ctx, input)
	if err != nil {
		return &ClientError{
			ResourceName: dbInstanceId,
			Err:          err,
		}
	}

	return nil
}

func (r *RDS) CheckDBClusterDeletionProtection(ctx context.Context, dbClusterId *string) (bool, error) {
	input := &rds.DescribeDBClustersInput{
		DBClusterIdentifier: dbClusterId,
	}

	output, err := r.client.DescribeDBClusters(ctx, input)
	if err != nil {
		return false, &ClientError{
			ResourceName: dbClusterId,
			Err:          err,
		}
	}

	if len(output.DBClusters) == 0 {
		return false, nil
	}

	return aws.ToBool(output.DBClusters[0].DeletionProtection), nil
}

func (r *RDS) DisableDBClusterDeletionProtection(ctx context.Context, dbClusterId *string) error {
	input := &rds.ModifyDBClusterInput{
		DBClusterIdentifier: dbClusterId,
		DeletionProtection:  aws.Bool(false),
	}

	_, err := r.client.ModifyDBCluster(ctx, input)
	if err != nil {
		return &ClientError{
			ResourceName: dbClusterId,
			Err:          err,
		}
	}

	return nil
}
