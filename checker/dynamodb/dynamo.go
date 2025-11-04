package dynamodb

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/botchris/go-health"
)

// dynamoChecker implements health.Checker for DynamoDB.
type dynamoChecker struct {
	opts *options
}

// NewChecker creates a new DynamoDB checker with the given options.
func NewChecker(client *dynamodb.Client, tableName string, o ...Option) (health.Checker, error) {
	opts := &options{
		client: client,
		table:  tableName,
	}

	for i := range o {
		if err := o[i](opts); err != nil {
			return nil, fmt.Errorf("dynamodb: applying option %d failed: %w", i, err)
		}
	}

	return &dynamoChecker{
		opts: opts,
	}, nil
}

// Check verifies connectivity and optionally permissions to DynamoDB.
func (c *dynamoChecker) Check(ctx context.Context) error {
	if _, err := c.opts.client.DescribeTable(ctx, &dynamodb.DescribeTableInput{TableName: aws.String(c.opts.table)}); err != nil {
		return fmt.Errorf("dynamodb connectivity failed: %w", err)
	}

	if c.opts.getCheck {
		_, err := c.opts.client.GetItem(ctx, &dynamodb.GetItemInput{
			TableName: aws.String(c.opts.table),
			Key:       map[string]types.AttributeValue{}, // empty key, will fail if table requires key, but will check permission
		})

		if err != nil && !isAccessDenied(err) {
			return fmt.Errorf("dynamodb GetItem check failed: %w", err)
		}
	}

	if c.opts.batchGetCheck {
		_, err := c.opts.client.BatchGetItem(ctx, &dynamodb.BatchGetItemInput{
			RequestItems: map[string]types.KeysAndAttributes{
				c.opts.table: {
					Keys: []map[string]types.AttributeValue{{}}, // empty key
				},
			},
		})

		if err != nil && !isAccessDenied(err) {
			return fmt.Errorf("dynamodb BatchGetItem check failed: %w", err)
		}
	}

	if c.opts.putItemCheck {
		_, err := c.opts.client.PutItem(ctx, &dynamodb.PutItemInput{
			TableName: aws.String(c.opts.table),
			Item:      map[string]types.AttributeValue{}, // empty item
		})

		if err != nil && !isAccessDenied(err) {
			return fmt.Errorf("dynamodb PutItem check failed: %w", err)
		}
	}

	return nil
}

// isAccessDenied returns true if the error is access denied error.
func isAccessDenied(err error) bool {
	// AWS SDK v2 does not export AccessDeniedException as a type, so check error string
	if err == nil {
		return false
	}

	msg := err.Error()

	return strings.Contains(msg, "AccessDenied") || strings.Contains(msg, "not authorized") || strings.Contains(msg, "access denied")
}
