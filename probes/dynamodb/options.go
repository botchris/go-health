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

	Get                                 bool
	BatchGet                            bool
	Query                               bool
	Scan                                bool
	Put                                 bool
	BatchWrite                          bool
	Delete                              bool
	Update                              bool
	UpdateTable                         bool
	UpdateItem                          bool
	CreateTable                         bool
	DeleteTable                         bool
	DescribeTable                       bool
	DescribeStream                      bool
	ListTables                          bool
	ListStreams                         bool
	TagResource                         bool
	UntagResource                       bool
	TransactGet                         bool
	TransactWrite                       bool
	RestoreTable                        bool
	ExportTable                         bool
	ImportTable                         bool
	DescribeLimits                      bool
	DescribeBackups                     bool
	DescribeContinuousBackups           bool
	DescribeContributorInsights         bool
	DescribeKinesisStreamingDestination bool
	DescribeReservedCapacity            bool
	DescribeReservedCapacityOfferings   bool
	DescribeTableReplicaAutoScaling     bool
	DescribeTimeToLive                  bool
	EnableKinesisStreamingDestination   bool
	DisableKinesisStreamingDestination  bool
	EnableReplication                   bool
	DisableReplication                  bool
	RestoreTableFromBackup              bool
	RestoreTableToPointInTime           bool
	UpdateContinuousBackups             bool
	UpdateContributorInsights           bool
	UpdateGlobalTable                   bool
	UpdateGlobalTableSettings           bool
	UpdateKinesisStreamingDestination   bool
	UpdateTableReplicaAutoScaling       bool
	UpdateTimeToLive                    bool
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
