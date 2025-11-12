package s3

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

// Option represents a configuration option for the S3 probe.
type Option func(*options) error

type options struct {
	bucket      string
	client      Client
	permissions *PermissionsCheck
}

type PermissionsCheck struct {
	IAM IAMClient
	STS STSClient

	Read         ReadPermissions
	Write        WritePermissions
	Delete       DeletePermissions
	BucketConfig BucketConfigPermissions
	BucketPolicy BucketPolicyPermissions
	Security     SecurityPermissions
	Lifecycle    LifecyclePermissions
	Analytics    AnalyticsPermissions
	AccessPoint  AccessPointPermissions
	ObjectLock   ObjectLockPermissions
	Batch        BatchPermissions
	Account      AccountPermissions

	actionsMap map[string]bool
}

// WithClient sets a custom S3 client for the probe.
func WithClient(client Client) Option {
	return func(o *options) error {
		o.client = client

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
