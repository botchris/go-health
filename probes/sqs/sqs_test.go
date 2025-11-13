package sqs_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	iamt "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	sqsprobe "github.com/botchris/go-health/probes/sqs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Run("creates probe with valid queue URL", func(t *testing.T) {
		client := &mockSQSClient{}
		probe, err := sqsprobe.New(
			"https://sqs.us-east-1.amazonaws.com/123456789012/test-queue",
			sqsprobe.WithClient(client),
		)

		require.NoError(t, err)
		assert.NotNil(t, probe)
	})

	t.Run("creates probe without client (uses default)", func(t *testing.T) {
		// This test will attempt to load default AWS config
		// In a real environment, this would succeed if AWS credentials are configured
		probe, err := sqsprobe.New("https://sqs.us-east-1.amazonaws.com/123456789012/test-queue")

		// We expect either success or an error related to AWS config loading
		if err != nil {
			assert.Contains(t, err.Error(), "sqs probe: preparing options failed")
		} else {
			assert.NotNil(t, probe)
		}
	})
}

func TestCheck_Connectivity(t *testing.T) {
	t.Run("successful connectivity check", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		client := &mockSQSClient{
			getQueueAttributesFunc: func(ctx context.Context, params *sqs.GetQueueAttributesInput, optFns ...func(*sqs.Options)) (*sqs.GetQueueAttributesOutput, error) {
				return &sqs.GetQueueAttributesOutput{
					Attributes: map[string]string{
						"QueueArn": "arn:aws:sqs:us-east-1:123456789012:test-queue",
					},
				}, nil
			},
		}

		probe, err := sqsprobe.New(
			"https://sqs.us-east-1.amazonaws.com/123456789012/test-queue",
			sqsprobe.WithClient(client),
		)
		require.NoError(t, err)

		assert.NoError(t, probe.Check(ctx))
	})

	t.Run("failed connectivity check", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		expectedErr := errors.New("queue not found")
		client := &mockSQSClient{
			getQueueAttributesFunc: func(ctx context.Context, params *sqs.GetQueueAttributesInput, optFns ...func(*sqs.Options)) (*sqs.GetQueueAttributesOutput, error) {
				return nil, expectedErr
			},
		}

		probe, err := sqsprobe.New(
			"https://sqs.us-east-1.amazonaws.com/123456789012/test-queue",
			sqsprobe.WithClient(client),
		)
		require.NoError(t, err)

		err = probe.Check(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "sqs connectivity failed")
	})
}

func TestCheck_Permissions(t *testing.T) {
	t.Run("successful permission check", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		sqsClient := &mockSQSClient{}
		iamClient := &mockIAMClient{
			simulatePrincipalPolicyFunc: func(ctx context.Context, params *iam.SimulatePrincipalPolicyInput, optFns ...func(*iam.Options)) (*iam.SimulatePrincipalPolicyOutput, error) {
				return &iam.SimulatePrincipalPolicyOutput{
					EvaluationResults: []iamt.EvaluationResult{
						{EvalDecision: iamt.PolicyEvaluationDecisionTypeAllowed},
					},
				}, nil
			},
		}
		stsClient := &mockSTSClient{}

		permCheck := sqsprobe.PermissionsCheck{
			IAM: iamClient,
			STS: stsClient,
			Message: sqsprobe.MessagePermissions{
				SendMessage:    true,
				ReceiveMessage: true,
			},
		}

		probe, err := sqsprobe.New(
			"https://sqs.us-east-1.amazonaws.com/123456789012/test-queue",
			sqsprobe.WithClient(sqsClient),
			sqsprobe.WithPermissionsCheck(permCheck),
		)
		require.NoError(t, err)

		assert.NoError(t, probe.Check(ctx))
	})

	t.Run("permission denied", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		sqsClient := &mockSQSClient{}
		iamClient := &mockIAMClient{
			simulatePrincipalPolicyFunc: func(ctx context.Context, params *iam.SimulatePrincipalPolicyInput, optFns ...func(*iam.Options)) (*iam.SimulatePrincipalPolicyOutput, error) {
				return &iam.SimulatePrincipalPolicyOutput{
					EvaluationResults: []iamt.EvaluationResult{
						{EvalDecision: iamt.PolicyEvaluationDecisionTypeExplicitDeny},
					},
				}, nil
			},
		}
		stsClient := &mockSTSClient{}

		permCheck := sqsprobe.PermissionsCheck{
			IAM: iamClient,
			STS: stsClient,
			Message: sqsprobe.MessagePermissions{
				SendMessage: true,
			},
		}

		probe, err := sqsprobe.New(
			"https://sqs.us-east-1.amazonaws.com/123456789012/test-queue",
			sqsprobe.WithClient(sqsClient),
			sqsprobe.WithPermissionsCheck(permCheck),
		)
		require.NoError(t, err)

		err = probe.Check(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "permission denied")
	})

	t.Run("failed to get caller identity", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		sqsClient := &mockSQSClient{}
		iamClient := &mockIAMClient{}
		stsClient := &mockSTSClient{
			getCallerIdentityFunc: func(ctx context.Context, params *sts.GetCallerIdentityInput, optFns ...func(*sts.Options)) (*sts.GetCallerIdentityOutput, error) {
				return nil, errors.New("sts error")
			},
		}

		permCheck := sqsprobe.PermissionsCheck{
			IAM: iamClient,
			STS: stsClient,
			Message: sqsprobe.MessagePermissions{
				SendMessage: true,
			},
		}

		probe, err := sqsprobe.New(
			"https://sqs.us-east-1.amazonaws.com/123456789012/test-queue",
			sqsprobe.WithClient(sqsClient),
			sqsprobe.WithPermissionsCheck(permCheck),
		)
		require.NoError(t, err)

		err = probe.Check(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get caller identity")
	})

	t.Run("simulate policy error", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		sqsClient := &mockSQSClient{}
		iamClient := &mockIAMClient{
			simulatePrincipalPolicyFunc: func(ctx context.Context, params *iam.SimulatePrincipalPolicyInput, optFns ...func(*iam.Options)) (*iam.SimulatePrincipalPolicyOutput, error) {
				return nil, errors.New("iam error")
			},
		}
		stsClient := &mockSTSClient{}

		permCheck := sqsprobe.PermissionsCheck{
			IAM: iamClient,
			STS: stsClient,
			Message: sqsprobe.MessagePermissions{
				SendMessage: true,
			},
		}

		probe, err := sqsprobe.New(
			"https://sqs.us-east-1.amazonaws.com/123456789012/test-queue",
			sqsprobe.WithClient(sqsClient),
			sqsprobe.WithPermissionsCheck(permCheck),
		)
		require.NoError(t, err)

		err = probe.Check(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "simulate policy")
	})

	t.Run("multiple permission checks", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		sqsClient := &mockSQSClient{}
		callCount := 0
		iamClient := &mockIAMClient{
			simulatePrincipalPolicyFunc: func(ctx context.Context, params *iam.SimulatePrincipalPolicyInput, optFns ...func(*iam.Options)) (*iam.SimulatePrincipalPolicyOutput, error) {
				callCount++

				return &iam.SimulatePrincipalPolicyOutput{
					EvaluationResults: []iamt.EvaluationResult{
						{EvalDecision: iamt.PolicyEvaluationDecisionTypeAllowed},
					},
				}, nil
			},
		}
		stsClient := &mockSTSClient{}

		permCheck := sqsprobe.PermissionsCheck{
			IAM: iamClient,
			STS: stsClient,
			Message: sqsprobe.MessagePermissions{
				SendMessage:    true,
				ReceiveMessage: true,
				DeleteMessage:  true,
			},
			Queue: sqsprobe.QueuePermissions{
				GetQueueAttributes: true,
			},
		}

		probe, err := sqsprobe.New(
			"https://sqs.us-east-1.amazonaws.com/123456789012/test-queue",
			sqsprobe.WithClient(sqsClient),
			sqsprobe.WithPermissionsCheck(permCheck),
		)
		require.NoError(t, err)

		assert.NoError(t, probe.Check(ctx))
		assert.Equal(t, 4, callCount, "should check 4 permissions")
	})

	t.Run("failed to get queue ARN for permissions check", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		sqsClient := &mockSQSClient{
			getQueueAttributesFunc: func(ctx context.Context, params *sqs.GetQueueAttributesInput, optFns ...func(*sqs.Options)) (*sqs.GetQueueAttributesOutput, error) {
				return nil, errors.New("failed to get attributes")
			},
		}

		iamClient := &mockIAMClient{}
		stsClient := &mockSTSClient{}

		permCheck := sqsprobe.PermissionsCheck{
			IAM: iamClient,
			STS: stsClient,
			Message: sqsprobe.MessagePermissions{
				SendMessage: true,
			},
		}

		probe, err := sqsprobe.New(
			"https://sqs.us-east-1.amazonaws.com/123456789012/test-queue",
			sqsprobe.WithClient(sqsClient),
			sqsprobe.WithPermissionsCheck(permCheck),
		)
		require.NoError(t, err)

		err = probe.Check(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "sqs connectivity failed")
	})
}

func TestCheck_AllPermissionTypes(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	sqsClient := &mockSQSClient{}
	iamClient := &mockIAMClient{}
	stsClient := &mockSTSClient{}

	permCheck := sqsprobe.PermissionsCheck{
		IAM: iamClient,
		STS: stsClient,
		Message: sqsprobe.MessagePermissions{
			SendMessage:                  true,
			SendMessageBatch:             true,
			ReceiveMessage:               true,
			DeleteMessage:                true,
			DeleteMessageBatch:           true,
			ChangeMessageVisibility:      true,
			ChangeMessageVisibilityBatch: true,
			PurgeQueue:                   true,
		},
		Queue: sqsprobe.QueuePermissions{
			CreateQueue:        true,
			DeleteQueue:        true,
			GetQueueUrl:        true,
			GetQueueAttributes: true,
			SetQueueAttributes: true,
		},
		Policy: sqsprobe.PolicyPermissions{
			AddPermission:    true,
			RemovePermission: true,
		},
		DeadLetter: sqsprobe.DeadLetterPermissions{
			StartMessageMoveTask:       true,
			CancelMessageMoveTask:      true,
			ListMessageMoveTasks:       true,
			ListDeadLetterSourceQueues: true,
		},
		Tags: sqsprobe.TagPermissions{
			TagQueue:      true,
			UntagQueue:    true,
			ListQueueTags: true,
		},
		List: sqsprobe.ListPermissions{
			ListQueues: true,
		},
	}

	probe, err := sqsprobe.New(
		"https://sqs.us-east-1.amazonaws.com/123456789012/test-queue",
		sqsprobe.WithClient(sqsClient),
		sqsprobe.WithPermissionsCheck(permCheck),
	)
	require.NoError(t, err)

	assert.NoError(t, probe.Check(ctx))
}

type mockSQSClient struct {
	getQueueAttributesFunc func(ctx context.Context, params *sqs.GetQueueAttributesInput, optFns ...func(*sqs.Options)) (*sqs.GetQueueAttributesOutput, error)
}

func (m *mockSQSClient) GetQueueAttributes(ctx context.Context, params *sqs.GetQueueAttributesInput, optFns ...func(*sqs.Options)) (*sqs.GetQueueAttributesOutput, error) {
	if m.getQueueAttributesFunc != nil {
		return m.getQueueAttributesFunc(ctx, params, optFns...)
	}

	return &sqs.GetQueueAttributesOutput{
		Attributes: map[string]string{
			"QueueArn": "arn:aws:sqs:us-east-1:123456789012:test-queue",
		},
	}, nil
}

type mockIAMClient struct {
	simulatePrincipalPolicyFunc func(ctx context.Context, params *iam.SimulatePrincipalPolicyInput, optFns ...func(*iam.Options)) (*iam.SimulatePrincipalPolicyOutput, error)
}

func (m *mockIAMClient) SimulatePrincipalPolicy(ctx context.Context, params *iam.SimulatePrincipalPolicyInput, optFns ...func(*iam.Options)) (*iam.SimulatePrincipalPolicyOutput, error) {
	if m.simulatePrincipalPolicyFunc != nil {
		return m.simulatePrincipalPolicyFunc(ctx, params, optFns...)
	}

	return &iam.SimulatePrincipalPolicyOutput{
		EvaluationResults: []iamt.EvaluationResult{
			{EvalDecision: iamt.PolicyEvaluationDecisionTypeAllowed},
		},
	}, nil
}

type mockSTSClient struct {
	getCallerIdentityFunc func(ctx context.Context, params *sts.GetCallerIdentityInput, optFns ...func(*sts.Options)) (*sts.GetCallerIdentityOutput, error)
}

func (m *mockSTSClient) GetCallerIdentity(ctx context.Context, params *sts.GetCallerIdentityInput, optFns ...func(*sts.Options)) (*sts.GetCallerIdentityOutput, error) {
	if m.getCallerIdentityFunc != nil {
		return m.getCallerIdentityFunc(ctx, params, optFns...)
	}

	return &sts.GetCallerIdentityOutput{
		Arn: aws.String("arn:aws:iam::123456789012:user/test-user"),
	}, nil
}
