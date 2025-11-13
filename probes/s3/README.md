# AWS S3 Probe

The S3 Probe allows you to monitor the health of an AWS S3 bucket.
It checks the availability of your S3 buckets by requesting the bucket's attributes,
and optionally simulates IAM policies to verify permissions. For example, you
can check if your application has the necessary permissions to read or write
objects from the bucket.

---

## Prerequisites

The application running this probe must have the necessary AWS permissions to interact with SQS.
Ensure that the following IAM permissions are granted:

- `iam:SimulatePrincipalPolicy`
- `sts:GetCallerIdentity`
- `s3:HeadBucket`
