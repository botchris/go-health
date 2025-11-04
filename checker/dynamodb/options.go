package dynamodb

import "github.com/aws/aws-sdk-go-v2/service/dynamodb"

// Option is a function that configures options for the DynamoDB checker.
type Option func(*options) error

type options struct {
	client          *dynamodb.Client
	table           string
	getCheck        bool
	batchGetCheck   bool
	queryCheck      bool
	scanCheck       bool
	putItemCheck    bool
	batchWriteCheck bool
	deleteCheck     bool
}

func WithGetCheck() Option {
	return func(o *options) error {
		o.getCheck = true

		return nil
	}
}

func WithBatchGetCheck() Option {
	return func(o *options) error {
		o.batchGetCheck = true

		return nil
	}
}

func WithPutCheck() Option {
	return func(o *options) error {
		o.putItemCheck = true

		return nil
	}
}
