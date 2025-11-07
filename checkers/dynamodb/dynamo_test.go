package dynamodb_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	dynamot "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	iamt "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	dynamodbc "github.com/botchris/go-health/checkers/dynamodb"
	"github.com/stretchr/testify/assert"
)

func TestDynamoProbe_Check_Success(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client := &mockDynamoClient{
		describeTableFunc: func(ctx context.Context, params *dynamodb.DescribeTableInput, _ ...func(*dynamodb.Options)) (*dynamodb.DescribeTableOutput, error) {
			return &dynamodb.DescribeTableOutput{
				Table: &dynamot.TableDescription{
					TableStatus: dynamot.TableStatusActive,
					TableArn:    aws.String("arn:aws:dynamodb:us-east-1:123456789012:table/test"),
				},
			}, nil
		},
	}

	probe, err := dynamodbc.New(client, "test")
	assert.NoError(t, err)
	assert.NoError(t, probe.Check(ctx))
}

func TestDynamoProbe_Check_TableNotActive(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client := &mockDynamoClient{
		describeTableFunc: func(ctx context.Context, params *dynamodb.DescribeTableInput, _ ...func(*dynamodb.Options)) (*dynamodb.DescribeTableOutput, error) {
			return &dynamodb.DescribeTableOutput{
				Table: &dynamot.TableDescription{
					TableStatus: dynamot.TableStatusCreating,
					TableArn:    aws.String("arn:aws:dynamodb:us-east-1:123456789012:table/test"),
				},
			}, nil
		},
	}

	probe, err := dynamodbc.New(client, "test")
	assert.NoError(t, err)

	err = probe.Check(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "is not active")
}

func TestDynamoProbe_Check_DescribeTableError(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client := &mockDynamoClient{
		describeTableFunc: func(ctx context.Context, params *dynamodb.DescribeTableInput, _ ...func(*dynamodb.Options)) (*dynamodb.DescribeTableOutput, error) {
			return nil, errors.New("describe error")
		},
	}

	probe, err := dynamodbc.New(client, "test")
	assert.NoError(t, err)

	err = probe.Check(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "connectivity failed")
}

func TestDynamoProbe_Check_GlobalSecondaryIndexNotActive(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client := &mockDynamoClient{
		describeTableFunc: func(ctx context.Context, params *dynamodb.DescribeTableInput, _ ...func(*dynamodb.Options)) (*dynamodb.DescribeTableOutput, error) {
			return &dynamodb.DescribeTableOutput{
				Table: &dynamot.TableDescription{
					TableStatus: dynamot.TableStatusActive,
					TableArn:    aws.String("arn:aws:dynamodb:us-east-1:123456789012:table/test"),
					GlobalSecondaryIndexes: []dynamot.GlobalSecondaryIndexDescription{
						{
							IndexName:   aws.String("gsi1"),
							IndexStatus: dynamot.IndexStatusCreating,
						},
					},
				},
			}, nil
		},
	}

	probe, err := dynamodbc.New(client, "test")
	assert.NoError(t, err)

	err = probe.Check(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-active global secondary index")
}

func TestDynamoProbe_Check_Permissions_AllAllowed(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client := &mockDynamoClient{
		describeTableFunc: func(ctx context.Context, params *dynamodb.DescribeTableInput, _ ...func(*dynamodb.Options)) (*dynamodb.DescribeTableOutput, error) {
			return &dynamodb.DescribeTableOutput{
				Table: &dynamot.TableDescription{
					TableStatus: dynamot.TableStatusActive,
					TableArn:    aws.String("arn:aws:dynamodb:us-east-1:123456789012:table/test"),
				},
			}, nil
		},
	}

	iamClient := &mockIAMClient{
		simulateFunc: func(ctx context.Context, params *iam.SimulatePrincipalPolicyInput, _ ...func(*iam.Options)) (*iam.SimulatePrincipalPolicyOutput, error) {
			return &iam.SimulatePrincipalPolicyOutput{
				EvaluationResults: []iamt.EvaluationResult{
					{EvalDecision: iamt.PolicyEvaluationDecisionTypeAllowed},
				},
			}, nil
		},
	}

	stsClient := &mockSTSClient{
		getCallerFunc: func(ctx context.Context, params *sts.GetCallerIdentityInput, _ ...func(*sts.Options)) (*sts.GetCallerIdentityOutput, error) {
			return &sts.GetCallerIdentityOutput{Arn: aws.String("arn:aws:iam::123456789012:user/test")}, nil
		},
	}

	perm := dynamodbc.PermissionsCheck{
		IAM:        iamClient,
		STS:        stsClient,
		Get:        true,
		BatchGet:   true,
		Query:      true,
		Scan:       true,
		Put:        true,
		BatchWrite: true,
		Delete:     true,
	}

	probe, err := dynamodbc.New(client, "test", dynamodbc.WithPermissionsCheck(perm))
	assert.NoError(t, err)
	assert.NoError(t, probe.Check(ctx))
}

func TestDynamoProbe_Check_Permissions_Denied(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client := &mockDynamoClient{
		describeTableFunc: func(ctx context.Context, params *dynamodb.DescribeTableInput, _ ...func(*dynamodb.Options)) (*dynamodb.DescribeTableOutput, error) {
			return &dynamodb.DescribeTableOutput{
				Table: &dynamot.TableDescription{
					TableStatus: dynamot.TableStatusActive,
					TableArn:    aws.String("arn:aws:dynamodb:us-east-1:123456789012:table/test"),
				},
			}, nil
		},
	}

	iamClient := &mockIAMClient{
		simulateFunc: func(ctx context.Context, params *iam.SimulatePrincipalPolicyInput, _ ...func(*iam.Options)) (*iam.SimulatePrincipalPolicyOutput, error) {
			return &iam.SimulatePrincipalPolicyOutput{
				EvaluationResults: []iamt.EvaluationResult{
					{EvalDecision: iamt.PolicyEvaluationDecisionTypeExplicitDeny},
				},
			}, nil
		},
	}

	stsClient := &mockSTSClient{
		getCallerFunc: func(ctx context.Context, params *sts.GetCallerIdentityInput, _ ...func(*sts.Options)) (*sts.GetCallerIdentityOutput, error) {
			return &sts.GetCallerIdentityOutput{Arn: aws.String("arn:aws:iam::123456789012:user/test")}, nil
		},
	}

	perm := dynamodbc.PermissionsCheck{
		IAM: iamClient,
		STS: stsClient,
		Get: true,
		Put: true,
	}

	probe, err := dynamodbc.New(client, "test", dynamodbc.WithPermissionsCheck(perm))
	assert.NoError(t, err)

	err = probe.Check(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "permission denied")
}

func TestDynamoProbe_Check_Permissions_SimulateError(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client := &mockDynamoClient{
		describeTableFunc: func(ctx context.Context, params *dynamodb.DescribeTableInput, _ ...func(*dynamodb.Options)) (*dynamodb.DescribeTableOutput, error) {
			return &dynamodb.DescribeTableOutput{
				Table: &dynamot.TableDescription{
					TableStatus: dynamot.TableStatusActive,
					TableArn:    aws.String("arn:aws:dynamodb:us-east-1:123456789012:table/test"),
				},
			}, nil
		},
	}

	iamClient := &mockIAMClient{
		simulateFunc: func(ctx context.Context, params *iam.SimulatePrincipalPolicyInput, _ ...func(*iam.Options)) (*iam.SimulatePrincipalPolicyOutput, error) {
			return nil, errors.New("simulate error")
		},
	}

	stsClient := &mockSTSClient{
		getCallerFunc: func(ctx context.Context, params *sts.GetCallerIdentityInput, _ ...func(*sts.Options)) (*sts.GetCallerIdentityOutput, error) {
			return &sts.GetCallerIdentityOutput{Arn: aws.String("arn:aws:iam::123456789012:user/test")}, nil
		},
	}

	perm := dynamodbc.PermissionsCheck{
		IAM: iamClient,
		STS: stsClient,
		Get: true,
	}

	probe, err := dynamodbc.New(client, "test", dynamodbc.WithPermissionsCheck(perm))
	assert.NoError(t, err)

	err = probe.Check(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "simulate policy for dynamodb:GetItem failed")
}

func TestDynamoProbe_Check_Permissions_STSError(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client := &mockDynamoClient{
		describeTableFunc: func(ctx context.Context, params *dynamodb.DescribeTableInput, _ ...func(*dynamodb.Options)) (*dynamodb.DescribeTableOutput, error) {
			return &dynamodb.DescribeTableOutput{
				Table: &dynamot.TableDescription{
					TableStatus: dynamot.TableStatusActive,
					TableArn:    aws.String("arn:aws:dynamodb:us-east-1:123456789012:table/test"),
				},
			}, nil
		},
	}

	iamClient := &mockIAMClient{
		simulateFunc: func(ctx context.Context, params *iam.SimulatePrincipalPolicyInput, _ ...func(*iam.Options)) (*iam.SimulatePrincipalPolicyOutput, error) {
			return &iam.SimulatePrincipalPolicyOutput{}, nil
		},
	}

	stsClient := &mockSTSClient{
		getCallerFunc: func(ctx context.Context, params *sts.GetCallerIdentityInput, _ ...func(*sts.Options)) (*sts.GetCallerIdentityOutput, error) {
			return nil, errors.New("sts error")
		},
	}

	perm := dynamodbc.PermissionsCheck{
		IAM: iamClient,
		STS: stsClient,
		Get: true,
	}

	probe, err := dynamodbc.New(client, "test", dynamodbc.WithPermissionsCheck(perm))
	assert.NoError(t, err)

	err = probe.Check(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get caller identity")
}

type mockDynamoClient struct {
	describeTableFunc func(ctx context.Context, params *dynamodb.DescribeTableInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DescribeTableOutput, error)
}

func (m *mockDynamoClient) DescribeTable(ctx context.Context, params *dynamodb.DescribeTableInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DescribeTableOutput, error) {
	return m.describeTableFunc(ctx, params, optFns...)
}

type mockIAMClient struct {
	simulateFunc func(ctx context.Context, params *iam.SimulatePrincipalPolicyInput, optFns ...func(*iam.Options)) (*iam.SimulatePrincipalPolicyOutput, error)
}

func (m *mockIAMClient) SimulatePrincipalPolicy(ctx context.Context, params *iam.SimulatePrincipalPolicyInput, optFns ...func(*iam.Options)) (*iam.SimulatePrincipalPolicyOutput, error) {
	return m.simulateFunc(ctx, params, optFns...)
}

type mockSTSClient struct {
	getCallerFunc func(ctx context.Context, params *sts.GetCallerIdentityInput, optFns ...func(*sts.Options)) (*sts.GetCallerIdentityOutput, error)
}

func (m *mockSTSClient) GetCallerIdentity(ctx context.Context, params *sts.GetCallerIdentityInput, optFns ...func(*sts.Options)) (*sts.GetCallerIdentityOutput, error) {
	return m.getCallerFunc(ctx, params, optFns...)
}
