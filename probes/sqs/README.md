# AWS SQS Probe

The SQS Probe allows you to monitor the health of a AWS SQS queue.
It checks the availability your SQS queues by requesting the queue attributes,
and optionally simulates IAM policies to verify permissions. For example, you
can check if your application has the necessary permissions to send or receive
messages from the queue.

---

## Prerequisites

The application running this probe must have the necessary AWS permissions to interact with SQS.
Ensure that the following IAM permissions are granted:

- `iam:SimulatePrincipalPolicy`
- `sts:GetCallerIdentity`
- `sqs:GetQueueAttributes`
