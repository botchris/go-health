package s3

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	iamt "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/botchris/go-health"
)

type s3Probe struct {
	opts *options
}

func New(bucket string, o ...Option) (health.Probe, error) {
	opts, err := prepareOptions(bucket, o...)
	if err != nil {
		return nil, fmt.Errorf("dynamodb probe: preparing options failed: %w", err)
	}

	return s3Probe{opts: opts}, nil
}

func (s s3Probe) Check(ctx context.Context) error {
	res, hErr := s.opts.client.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: &s.opts.bucket})
	if hErr != nil {
		return fmt.Errorf("s3 connectivity failed: %w", hErr)
	}

	if s.opts.permissions != nil {
		if err := s.checkBucketPermissions(ctx, *res.BucketArn, s.opts.permissions); err != nil {
			return fmt.Errorf("%w: dynamodb permissions check failed", err)
		}
	}

	return nil
}

func (s s3Probe) checkBucketPermissions(ctx context.Context, bucketARN string, pc *PermissionsCheck) error {
	var errs []error

	identity, err := s.opts.permissions.STS.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
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
			ResourceArns:    []string{bucketARN},
		}

		result, sErr := s.opts.permissions.IAM.SimulatePrincipalPolicy(ctx, input)
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
			errs = append(errs, fmt.Errorf("permission denied for action %s on %s", action, bucketARN))
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

func prepareOptions(bucket string, o ...Option) (*options, error) {
	opts := &options{bucket: bucket}

	for i := range o {
		if err := o[i](opts); err != nil {
			return nil, fmt.Errorf("s3: applying option %d failed: %w", i, err)
		}
	}

	if opts.client == nil {
		cfg, err := config.LoadDefaultConfig(context.TODO())
		if err != nil {
			return nil, fmt.Errorf("s3: loading default AWS config failed: %w", err)
		}

		opts.client = s3.NewFromConfig(cfg)
	}

	return opts, nil
}
