package s3_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	iamt "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	s3p "github.com/botchris/go-health/probes/s3"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Run("creates probe successfully with custom client", func(t *testing.T) {
		client := &mockS3Client{}
		probe, err := s3p.New("test-bucket", s3p.WithClient(client))

		require.NoError(t, err)
		require.NotNil(t, probe)
	})

	t.Run("creates probe with permissions check", func(t *testing.T) {
		client := &mockS3Client{}
		iamClient := &mockIAMClient{}
		stsClient := &mockSTSClient{}

		permCheck := s3p.PermissionsCheck{
			IAM: iamClient,
			STS: stsClient,
			Read: s3p.ReadPermissions{
				GetObject: true,
			},
		}

		probe, err := s3p.New("test-bucket",
			s3p.WithClient(client),
			s3p.WithPermissionsCheck(permCheck),
		)

		require.NoError(t, err)
		require.NotNil(t, probe)
	})
}

func TestS3Probe_Check(t *testing.T) {
	t.Run("successful health check without permissions", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		client := &mockS3Client{
			headBucketFunc: func(ctx context.Context, params *s3.HeadBucketInput, optFns ...func(*s3.Options)) (*s3.HeadBucketOutput, error) {
				if *params.Bucket != "test-bucket" {
					t.Errorf("expected bucket name 'test-bucket', got %s", *params.Bucket)
				}

				return &s3.HeadBucketOutput{
					BucketArn: aws.String("arn:aws:s3:::test-bucket"),
				}, nil
			},
		}

		probe, err := s3p.New("test-bucket", s3p.WithClient(client))
		require.NoError(t, err)
		require.NoError(t, probe.Check(ctx))
	})

	t.Run("HeadBucket fails", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		expectedErr := errors.New("bucket not found")
		client := &mockS3Client{
			headBucketFunc: func(ctx context.Context, params *s3.HeadBucketInput, optFns ...func(*s3.Options)) (*s3.HeadBucketOutput, error) {
				return nil, expectedErr
			},
		}

		probe, err := s3p.New("test-bucket", s3p.WithClient(client))
		require.NoError(t, err)

		err = probe.Check(ctx)
		require.Error(t, err)
		errors.Is(err, expectedErr)
	})

	t.Run("successful with permission check", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		client := &mockS3Client{}
		iamClient := &mockIAMClient{}
		stsClient := &mockSTSClient{}

		permCheck := s3p.PermissionsCheck{
			IAM: iamClient,
			STS: stsClient,
			Read: s3p.ReadPermissions{
				GetObject:  true,
				ListBucket: true,
			},
		}

		probe, err := s3p.New("test-bucket",
			s3p.WithClient(client),
			s3p.WithPermissionsCheck(permCheck),
		)

		require.NoError(t, err)
		require.NoError(t, probe.Check(ctx))
	})

	t.Run("permission check fails on GetCallerIdentity", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		client := &mockS3Client{}
		expectedErr := errors.New("sts error")
		stsClient := &mockSTSClient{
			getCallerIdentityFunc: func(ctx context.Context, params *sts.GetCallerIdentityInput, optFns ...func(*sts.Options)) (*sts.GetCallerIdentityOutput, error) {
				return nil, expectedErr
			},
		}
		iamClient := &mockIAMClient{}

		permCheck := s3p.PermissionsCheck{
			IAM: iamClient,
			STS: stsClient,
			Read: s3p.ReadPermissions{
				GetObject: true,
			},
		}

		probe, err := s3p.New("test-bucket",
			s3p.WithClient(client),
			s3p.WithPermissionsCheck(permCheck),
		)
		require.NoError(t, err)

		err = probe.Check(ctx)
		require.Error(t, err)
		errors.Is(err, expectedErr)
	})

	t.Run("permission denied for action", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		client := &mockS3Client{}
		stsClient := &mockSTSClient{}
		iamClient := &mockIAMClient{
			simulatePrincipalPolicyFunc: func(ctx context.Context, params *iam.SimulatePrincipalPolicyInput, optFns ...func(*iam.Options)) (*iam.SimulatePrincipalPolicyOutput, error) {
				return &iam.SimulatePrincipalPolicyOutput{
					EvaluationResults: []iamt.EvaluationResult{
						{EvalDecision: iamt.PolicyEvaluationDecisionTypeExplicitDeny},
					},
				}, nil
			},
		}

		permCheck := s3p.PermissionsCheck{
			IAM: iamClient,
			STS: stsClient,
			Read: s3p.ReadPermissions{
				GetObject: true,
			},
		}

		probe, err := s3p.New("test-bucket",
			s3p.WithClient(client),
			s3p.WithPermissionsCheck(permCheck),
		)

		require.NoError(t, err)
		require.Error(t, probe.Check(ctx))
	})

	t.Run("SimulatePrincipalPolicy fails", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		client := &mockS3Client{}
		stsClient := &mockSTSClient{}
		expectedErr := errors.New("iam simulation error")
		iamClient := &mockIAMClient{
			simulatePrincipalPolicyFunc: func(ctx context.Context, params *iam.SimulatePrincipalPolicyInput, optFns ...func(*iam.Options)) (*iam.SimulatePrincipalPolicyOutput, error) {
				return nil, expectedErr
			},
		}

		permCheck := s3p.PermissionsCheck{
			IAM: iamClient,
			STS: stsClient,
			Write: s3p.WritePermissions{
				PutObject: true,
			},
		}

		probe, err := s3p.New("test-bucket",
			s3p.WithClient(client),
			s3p.WithPermissionsCheck(permCheck),
		)
		require.NoError(t, err)

		err = probe.Check(ctx)
		require.Error(t, err)
		errors.Is(err, expectedErr)
	})

	t.Run("multiple permission checks", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		client := &mockS3Client{}
		stsClient := &mockSTSClient{}

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

		permCheck := s3p.PermissionsCheck{
			IAM: iamClient,
			STS: stsClient,
			Read: s3p.ReadPermissions{
				GetObject:  true,
				ListBucket: true,
			},
			Write: s3p.WritePermissions{
				PutObject: true,
			},
			Delete: s3p.DeletePermissions{
				DeleteObject: true,
			},
		}

		probe, err := s3p.New("test-bucket",
			s3p.WithClient(client),
			s3p.WithPermissionsCheck(permCheck),
		)
		require.NoError(t, err)

		require.NoError(t, probe.Check(ctx))
		require.EqualValues(t, 4, callCount)
	})
}

func TestWithClient(t *testing.T) {
	client := &mockS3Client{}
	probe, err := s3p.New("test-bucket", s3p.WithClient(client))

	require.NoError(t, err)
	require.NotNil(t, probe)
}

func TestWithPermissionsCheck(t *testing.T) {
	t.Run("with IAM and STS clients provided", func(t *testing.T) {
		client := &mockS3Client{}
		iamClient := &mockIAMClient{}
		stsClient := &mockSTSClient{}

		permCheck := s3p.PermissionsCheck{
			IAM: iamClient,
			STS: stsClient,
		}

		probe, err := s3p.New("test-bucket",
			s3p.WithClient(client),
			s3p.WithPermissionsCheck(permCheck),
		)

		require.NoError(t, err)
		require.NotNil(t, probe)
	})

	t.Run("verifies all permission categories", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		client := &mockS3Client{}
		iamClient := &mockIAMClient{}
		stsClient := &mockSTSClient{}

		permCheck := s3p.PermissionsCheck{
			IAM: iamClient,
			STS: stsClient,
			BucketConfig: s3p.BucketConfigPermissions{
				GetBucketVersioning: true,
			},
			BucketPolicy: s3p.BucketPolicyPermissions{
				GetBucketPolicy: true,
			},
			Security: s3p.SecurityPermissions{
				GetEncryptionConfiguration: true,
			},
			Lifecycle: s3p.LifecyclePermissions{
				GetLifecycleConfiguration: true,
			},
			Analytics: s3p.AnalyticsPermissions{
				GetAnalyticsConfiguration: true,
			},
			AccessPoint: s3p.AccessPointPermissions{
				CreateAccessPoint: true,
			},
			ObjectLock: s3p.ObjectLockPermissions{
				GetObjectLockConfiguration: true,
			},
			Batch: s3p.BatchPermissions{
				CreateJob: true,
			},
			Account: s3p.AccountPermissions{
				ListAllMyBuckets: true,
			},
		}

		probe, err := s3p.New("test-bucket",
			s3p.WithClient(client),
			s3p.WithPermissionsCheck(permCheck),
		)

		require.NoError(t, err)
		require.NoError(t, probe.Check(ctx))
	})
}

type mockS3Client struct {
	headBucketFunc func(ctx context.Context, params *s3.HeadBucketInput, optFns ...func(*s3.Options)) (*s3.HeadBucketOutput, error)
}

func (m *mockS3Client) HeadBucket(ctx context.Context, params *s3.HeadBucketInput, optFns ...func(*s3.Options)) (*s3.HeadBucketOutput, error) {
	if m.headBucketFunc != nil {
		return m.headBucketFunc(ctx, params, optFns...)
	}

	return &s3.HeadBucketOutput{
		BucketArn: aws.String("arn:aws:s3:::test-bucket"),
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
