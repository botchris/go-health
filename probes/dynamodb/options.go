package dynamodb

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

// Option is a function that configures options for the DynamoDB Probe.
type Option func(*options) error

type options struct {
	client      Client
	table       string
	permissions *PermissionsCheck
}

type PermissionsCheck struct {
	IAM IAMClient
	STS STSClient

	ItemRead     ItemReadPermissions
	ItemWrite    ItemWritePermissions
	Query        QueryPermissions
	Transaction  TransactionPermissions
	Table        TablePermissions
	Stream       StreamPermissions
	Backup       BackupPermissions
	GlobalTable  GlobalTablePermissions
	Replication  ReplicationPermissions
	TTL          TTLPermissions
	Kinesis      KinesisPermissions
	Contributors ContributorInsightsPermissions
	Capacity     CapacityPermissions
	DataTransfer DataTransferPermissions
	Tags         TagPermissions
	AutoScaling  AutoScalingPermissions

	actionsMap map[string]bool
}

// WithClient configures the checker to use the provided DynamoDB client.
// Otherwise, a default client will be created using the default AWS configuration.
func WithClient(c Client) Option {
	return func(o *options) error {
		o.client = c

		return nil
	}
}

// WithPermissionsCheck configures the checker to perform permissions checks
// using the provided PermissionsCheck settings.
//
// If the IAM or STS clients are not provided, they will be created using
// the default AWS configuration.
func WithPermissionsCheck(c PermissionsCheck) Option {
	return func(o *options) error {
		c.actionsMap = prepareActionsMap(&c)

		if c.IAM != nil && c.STS != nil {
			o.permissions = &c

			return nil
		}

		cfg, err := config.LoadDefaultConfig(context.TODO())
		if err != nil {
			return fmt.Errorf("unable to load AWS config: %w", err)
		}

		if c.IAM == nil {
			c.IAM = iam.NewFromConfig(cfg)
		}

		if c.STS == nil {
			c.STS = sts.NewFromConfig(cfg)
		}

		o.permissions = &c

		return nil
	}
}
