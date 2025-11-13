package dynamodb

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	dynamot "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	iamt "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/botchris/go-health"
)

// dynamoProbe implements health.Probe for DynamoDB.
type dynamoProbe struct {
	opts *options
}

// New creates a new DynamoDB probe with the given options.
//
// You must ensure that the provided Client has at least
// permissions to perform the `dynamodb:DescribeTable` operation.
// Additional permissions can be checked by providing a PermissionsCheck
// option.
func New(tableName string, o ...Option) (health.Probe, error) {
	opts, err := prepareOptions(tableName, o...)
	if err != nil {
		return nil, fmt.Errorf("dynamodb probe: preparing options failed: %w", err)
	}

	return &dynamoProbe{opts: opts}, nil
}

// Check verifies connectivity and optionally permissions to DynamoDB.
func (c *dynamoProbe) Check(ctx context.Context) error {
	dsc, dErr := c.opts.client.DescribeTable(ctx, &dynamodb.DescribeTableInput{TableName: aws.String(c.opts.table)})
	if dErr != nil {
		return fmt.Errorf("dynamodb connectivity failed: %w", dErr)
	}

	if dsc.Table.TableStatus != dynamot.TableStatusActive {
		return fmt.Errorf("dynamodb table %s is not active", c.opts.table)
	}

	indexStatus := make([]error, 0)

	for i := range dsc.Table.GlobalSecondaryIndexes {
		if dsc.Table.GlobalSecondaryIndexes[i].IndexStatus != dynamot.IndexStatusActive {
			indexStatus = append(indexStatus, fmt.Errorf("dynamodb table %s has non-active global secondary index %s", c.opts.table, aws.ToString(dsc.Table.GlobalSecondaryIndexes[i].IndexName)))
		}
	}

	if len(indexStatus) > 0 {
		return errors.Join(indexStatus...)
	}

	if c.opts.permissions != nil {
		if err := c.checkDynamoPermissions(ctx, *dsc.Table.TableArn, c.opts.permissions); err != nil {
			return fmt.Errorf("dynamodb permissions check failed: %w", err)
		}
	}

	return nil
}

func (c *dynamoProbe) checkDynamoPermissions(ctx context.Context, tableARN string, pc *PermissionsCheck) error {
	var errs []error

	identity, err := c.opts.permissions.STS.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return fmt.Errorf("failed to get caller identity: %w", err)
	}

	principalARN := aws.ToString(identity.Arn)

	for action, enabled := range pc.actionsMap {
		if !enabled {
			continue
		}

		input := &iam.SimulatePrincipalPolicyInput{
			PolicySourceArn: aws.String(principalARN),
			ActionNames:     []string{action},
			ResourceArns:    []string{tableARN},
		}

		result, sErr := c.opts.permissions.IAM.SimulatePrincipalPolicy(ctx, input)
		if sErr != nil {
			errs = append(errs, fmt.Errorf("simulate policy for %s failed: %w", action, sErr))

			continue
		}

		allowed := false

		for x := range result.EvaluationResults {
			if result.EvaluationResults[x].EvalDecision == iamt.PolicyEvaluationDecisionTypeAllowed {
				allowed = true

				break
			}
		}

		if !allowed {
			errs = append(errs, fmt.Errorf("permission denied for action %s on %s", action, tableARN))
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

func prepareOptions(tableName string, o ...Option) (*options, error) {
	opts := &options{table: tableName}

	for i := range o {
		if err := o[i](opts); err != nil {
			return nil, fmt.Errorf("dynamodb: applying option %d failed: %w", i, err)
		}
	}

	if opts.client == nil {
		cfg, err := config.LoadDefaultConfig(context.TODO())
		if err != nil {
			return nil, fmt.Errorf("dynamodb: unable to load AWS config: %w", err)
		}

		opts.client = dynamodb.NewFromConfig(cfg)
	}

	return opts, nil
}
