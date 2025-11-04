package dynamodb

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/iam"
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
	dsc, dErr := c.opts.client.DescribeTable(ctx, &dynamodb.DescribeTableInput{TableName: aws.String(c.opts.table)})
	if dErr != nil {
		return fmt.Errorf("dynamodb connectivity failed: %w", dErr)
	}

	if dsc.Table.TableStatus != types.TableStatusActive {
		return fmt.Errorf("dynamodb table %s is not active", c.opts.table)
	}

	indexStatus := make([]error, 0)

	for i := range dsc.Table.GlobalSecondaryIndexes {
		if dsc.Table.GlobalSecondaryIndexes[i].IndexStatus != types.IndexStatusActive {
			indexStatus = append(indexStatus, fmt.Errorf("dynamodb table %s has non-active global secondary index %s", c.opts.table, aws.ToString(dsc.Table.GlobalSecondaryIndexes[i].IndexName)))
		}
	}

	if len(indexStatus) > 0 {
		return errors.Join(indexStatus...)
	}

	if c.opts.pChecker != nil {
		if err := c.checkDynamoPermissions(ctx, *dsc.Table.TableArn, c.opts.pChecker); err != nil {
			return fmt.Errorf("%w: dynamodb permissions check failed", err)
		}
	}

	return nil
}

func (c *dynamoChecker) checkDynamoPermissions(ctx context.Context, tableARN string, pChecker *PermissionsCheck) error {
	var errs []error

	actions := map[string]bool{
		"dynamodb:GetItem":        pChecker.Get,
		"dynamodb:BatchGetItem":   pChecker.BatchGet,
		"dynamodb:Query":          pChecker.Query,
		"dynamodb:Scan":           pChecker.Scan,
		"dynamodb:PutItem":        pChecker.Put,
		"dynamodb:BatchWriteItem": pChecker.BatchWrite,
		"dynamodb:DeleteItem":     pChecker.Delete,
	}

	resource := tableARN

	for action, enabled := range actions {
		if !enabled {
			continue
		}

		input := &iam.SimulatePrincipalPolicyInput{
			PolicySourceArn: aws.String(pChecker.PrincipalARN),
			ActionNames:     []string{action},
			ResourceArns:    []string{resource},
		}

		out, err := pChecker.IAM.SimulatePrincipalPolicy(ctx, input)
		if err != nil {
			errs = append(errs, fmt.Errorf("simulate policy for %s failed: %w", action, err))

			continue
		}

		allowed := false

		for x := range out.EvaluationResults {
			if out.EvaluationResults[x].EvalDecision == "allowed" || out.EvaluationResults[x].EvalDecision == "Allowed" {
				allowed = true

				break
			}
		}

		if !allowed {
			errs = append(errs, fmt.Errorf("permission denied for action %s on %s", action, resource))
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}
