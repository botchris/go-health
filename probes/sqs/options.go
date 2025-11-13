package sqs

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

// Option represents a configuration option for the SQS probe.
type Option func(*options) error

type options struct {
	queueURL    string
	client      Client
	permissions *PermissionsCheck
}

// PermissionsCheck defines the permissions to be checked for the SQS probe.
type PermissionsCheck struct {
	IAM IAMClient
	STS STSClient

	Message    MessagePermissions
	Queue      QueuePermissions
	Policy     PolicyPermissions
	DeadLetter DeadLetterPermissions
	Tags       TagPermissions
	List       ListPermissions

	actionsMap map[string]bool
}

// WithClient sets a custom SQS client for the probe.
func WithClient(client Client) Option {
	return func(o *options) error {
		o.client = client

		return nil
	}
}

// WithPermissionsCheck configures the checker to perform permissions checks
// using the provided PermissionsCheck settings.
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
