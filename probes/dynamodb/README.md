# AWS DynamoDB Probe

The DynamoDB probe allows you to monitor the health of an AWS DynamoDB table.
It checks the availability of your DynamoDB tables by requesting the table's attributes,
and optionally simulates IAM policies to verify permissions. For example, you
can check if your application has the necessary permissions to read or write
items from the table.

---

## Prerequisites

The application running this probe must have the necessary AWS permissions to interact with SQS.
Ensure that the following IAM permissions are granted:

- `iam:SimulatePrincipalPolicy`
- `sts:GetCallerIdentity`
- `dynamodb:DescribeTable`
