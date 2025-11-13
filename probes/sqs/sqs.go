package sqs

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	iamt "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/botchris/go-health"
)

const queueARNAttribute = "QueueArn"

type sqsProbe struct {
	opts *options
}

// New creates a new SQS probe with the given options.
// You must ensure that the provided Client has at least
// permissions to perform the `sqs:GetQueueAttributes` operation.
// Additional permissions can be checked by providing a PermissionsCheck option.
func New(queueURL string, o ...Option) (health.Probe, error) {
	opts, err := prepareOptions(queueURL, o...)
	if err != nil {
		return nil, fmt.Errorf("sqs probe: preparing options failed: %w", err)
	}

	return sqsProbe{opts: opts}, nil
}

func (s sqsProbe) Check(ctx context.Context) error {
	res, hErr := s.opts.client.GetQueueAttributes(ctx, &sqs.GetQueueAttributesInput{
		QueueUrl:       &s.opts.queueURL,
		AttributeNames: []types.QueueAttributeName{queueARNAttribute},
	})

	if hErr != nil {
		return fmt.Errorf("sqs connectivity failed: %w", hErr)
	}

	queueARN := res.Attributes[queueARNAttribute]

	if s.opts.permissions != nil {
		if pErr := s.checkQueuePermissions(ctx, queueARN, s.opts.permissions); pErr != nil {
			return fmt.Errorf("%w: sqs permissions check failed", pErr)
		}
	}

	return nil
}

func (s sqsProbe) checkQueuePermissions(ctx context.Context, queueARN string, pc *PermissionsCheck) error {
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
			ResourceArns:    []string{queueARN},
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
			errs = append(errs, fmt.Errorf("permission denied for action %s on %s", action, queueARN))
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

func prepareOptions(queueURL string, o ...Option) (*options, error) {
	opts := &options{queueURL: queueURL}

	for i := range o {
		if err := o[i](opts); err != nil {
			return nil, fmt.Errorf("sqs: applying option %d failed: %w", i, err)
		}
	}

	if opts.client == nil {
		cfg, err := config.LoadDefaultConfig(context.TODO())
		if err != nil {
			return nil, fmt.Errorf("sqs: loading default AWS config failed: %w", err)
		}

		opts.client = sqs.NewFromConfig(cfg)
	}

	return opts, nil
}
