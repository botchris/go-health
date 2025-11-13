package sqs

// MessagePermissions defines the message-related permissions to be checked.
type MessagePermissions struct {
	SendMessage                  bool
	SendMessageBatch             bool
	ReceiveMessage               bool
	DeleteMessage                bool
	DeleteMessageBatch           bool
	ChangeMessageVisibility      bool
	ChangeMessageVisibilityBatch bool
	PurgeQueue                   bool
}

// QueuePermissions defines the queue-related permissions to be checked.
type QueuePermissions struct {
	CreateQueue        bool
	DeleteQueue        bool
	GetQueueUrl        bool
	GetQueueAttributes bool
	SetQueueAttributes bool
}

// PolicyPermissions defines the policy-related permissions to be checked.
type PolicyPermissions struct {
	AddPermission    bool
	RemovePermission bool
}

// DeadLetterPermissions defines the dead letter queue-related permissions to be checked.
type DeadLetterPermissions struct {
	StartMessageMoveTask       bool
	CancelMessageMoveTask      bool
	ListMessageMoveTasks       bool
	ListDeadLetterSourceQueues bool
}

// TagPermissions defines the tag-related permissions to be checked.
type TagPermissions struct {
	TagQueue      bool
	UntagQueue    bool
	ListQueueTags bool
}

// ListPermissions defines the list-related permissions to be checked.
type ListPermissions struct {
	ListQueues bool
}

// prepareActionsMap creates a map of SQS action names to their permission check values
func prepareActionsMap(pc *PermissionsCheck) map[string]bool {
	return map[string]bool{
		// Message Permissions
		"sqs:SendMessage":                  pc.Message.SendMessage,
		"sqs:SendMessageBatch":             pc.Message.SendMessageBatch,
		"sqs:ReceiveMessage":               pc.Message.ReceiveMessage,
		"sqs:DeleteMessage":                pc.Message.DeleteMessage,
		"sqs:DeleteMessageBatch":           pc.Message.DeleteMessageBatch,
		"sqs:ChangeMessageVisibility":      pc.Message.ChangeMessageVisibility,
		"sqs:ChangeMessageVisibilityBatch": pc.Message.ChangeMessageVisibilityBatch,
		"sqs:PurgeQueue":                   pc.Message.PurgeQueue,

		// Queue Permissions
		"sqs:CreateQueue":        pc.Queue.CreateQueue,
		"sqs:DeleteQueue":        pc.Queue.DeleteQueue,
		"sqs:GetQueueUrl":        pc.Queue.GetQueueUrl,
		"sqs:GetQueueAttributes": pc.Queue.GetQueueAttributes,
		"sqs:SetQueueAttributes": pc.Queue.SetQueueAttributes,

		// Policy Permissions
		"sqs:AddPermission":    pc.Policy.AddPermission,
		"sqs:RemovePermission": pc.Policy.RemovePermission,

		// Dead Letter Permissions
		"sqs:StartMessageMoveTask":       pc.DeadLetter.StartMessageMoveTask,
		"sqs:CancelMessageMoveTask":      pc.DeadLetter.CancelMessageMoveTask,
		"sqs:ListMessageMoveTasks":       pc.DeadLetter.ListMessageMoveTasks,
		"sqs:ListDeadLetterSourceQueues": pc.DeadLetter.ListDeadLetterSourceQueues,

		// Tag Permissions
		"sqs:TagQueue":      pc.Tags.TagQueue,
		"sqs:UntagQueue":    pc.Tags.UntagQueue,
		"sqs:ListQueueTags": pc.Tags.ListQueueTags,

		// List Permissions
		"sqs:ListQueues": pc.List.ListQueues,
	}
}
