package dynamodb

import (
	"errors"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/iam"
)

// Option is a function that configures options for the DynamoDB checker.
type Option func(*options) error

type PermissionsCheck struct {
	IAM          *iam.Client
	PrincipalARN string

	Get        bool
	BatchGet   bool
	Query      bool
	Scan       bool
	Put        bool
	BatchWrite bool
	Delete     bool
}

type options struct {
	client   *dynamodb.Client
	table    string
	pChecker *PermissionsCheck
}

// WithPermissionsCheck configures the checker to perform permissions checks
// using the provided PermissionsCheck settings.
func WithPermissionsCheck(c PermissionsCheck) Option {
	return func(o *options) error {
		if c.IAM == nil {
			return errors.New("IAM client is required for permissions check")
		}

		if c.PrincipalARN == "" {
			return errors.New("PrincipalARN is required for permissions check")
		}

		o.pChecker = &c

		return nil
	}
}
