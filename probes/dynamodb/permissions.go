package dynamodb

type ItemReadPermissions struct {
	GetItem      bool
	BatchGetItem bool
}

type ItemWritePermissions struct {
	PutItem        bool
	UpdateItem     bool
	DeleteItem     bool
	BatchWriteItem bool
}

type QueryPermissions struct {
	Query         bool
	Scan          bool
	PartiQLSelect bool
	PartiQLInsert bool
	PartiQLUpdate bool
	PartiQLDelete bool
}

type TransactionPermissions struct {
	TransactGetItems   bool
	TransactWriteItems bool
	ConditionCheckItem bool
}

type TablePermissions struct {
	CreateTable                    bool
	DeleteTable                    bool
	UpdateTable                    bool
	DescribeTable                  bool
	ListTables                     bool
	DescribeGlobalSecondaryIndexes bool
	DescribeLocalSecondaryIndexes  bool
}

type StreamPermissions struct {
	DescribeStream   bool
	GetShardIterator bool
	GetRecords       bool
	ListStreams      bool
}

type BackupPermissions struct {
	CreateBackup              bool
	DeleteBackup              bool
	DescribeBackup            bool
	ListBackups               bool
	RestoreTableFromBackup    bool
	RestoreTableToPointInTime bool
	DescribeContinuousBackups bool
	UpdateContinuousBackups   bool
}

type GlobalTablePermissions struct {
	CreateGlobalTable           bool
	UpdateGlobalTable           bool
	DescribeGlobalTable         bool
	ListGlobalTables            bool
	DescribeGlobalTableSettings bool
	UpdateGlobalTableSettings   bool
}

type ReplicationPermissions struct {
	CreateTableReplica bool
	DeleteTableReplica bool
	EnableReplication  bool
	DisableReplication bool
}

type TTLPermissions struct {
	DescribeTimeToLive bool
	UpdateTimeToLive   bool
}

type KinesisPermissions struct {
	DescribeKinesisStreamingDestination bool
	EnableKinesisStreamingDestination   bool
	DisableKinesisStreamingDestination  bool
	UpdateKinesisStreamingDestination   bool
}

type ContributorInsightsPermissions struct {
	DescribeContributorInsights bool
	UpdateContributorInsights   bool
}

type CapacityPermissions struct {
	DescribeLimits                    bool
	DescribeReservedCapacity          bool
	DescribeReservedCapacityOfferings bool
	PurchaseReservedCapacityOfferings bool
}

type DataTransferPermissions struct {
	ExportTableToPointInTime bool
	DescribeExport           bool
	ListExports              bool
	ImportTable              bool
	DescribeImport           bool
	ListImports              bool
}

type TagPermissions struct {
	TagResource        bool
	UntagResource      bool
	ListTagsOfResource bool
}

type AutoScalingPermissions struct {
	DescribeTableReplicaAutoScaling bool
	UpdateTableReplicaAutoScaling   bool
}

// prepareActionsMap creates a map of DynamoDB action names to their permission check values
func prepareActionsMap(pc *PermissionsCheck) map[string]bool {
	return map[string]bool{
		// Item Read Permissions
		"dynamodb:GetItem":      pc.ItemRead.GetItem,
		"dynamodb:BatchGetItem": pc.ItemRead.BatchGetItem,

		// Item Write Permissions
		"dynamodb:PutItem":        pc.ItemWrite.PutItem,
		"dynamodb:UpdateItem":     pc.ItemWrite.UpdateItem,
		"dynamodb:DeleteItem":     pc.ItemWrite.DeleteItem,
		"dynamodb:BatchWriteItem": pc.ItemWrite.BatchWriteItem,

		// Query Permissions
		"dynamodb:Query":         pc.Query.Query,
		"dynamodb:Scan":          pc.Query.Scan,
		"dynamodb:PartiQLSelect": pc.Query.PartiQLSelect,
		"dynamodb:PartiQLInsert": pc.Query.PartiQLInsert,
		"dynamodb:PartiQLUpdate": pc.Query.PartiQLUpdate,
		"dynamodb:PartiQLDelete": pc.Query.PartiQLDelete,

		// Transaction Permissions
		"dynamodb:TransactGetItems":   pc.Transaction.TransactGetItems,
		"dynamodb:TransactWriteItems": pc.Transaction.TransactWriteItems,
		"dynamodb:ConditionCheckItem": pc.Transaction.ConditionCheckItem,

		// Table Permissions
		"dynamodb:CreateTable":                    pc.Table.CreateTable,
		"dynamodb:DeleteTable":                    pc.Table.DeleteTable,
		"dynamodb:UpdateTable":                    pc.Table.UpdateTable,
		"dynamodb:DescribeTable":                  pc.Table.DescribeTable,
		"dynamodb:ListTables":                     pc.Table.ListTables,
		"dynamodb:DescribeGlobalSecondaryIndexes": pc.Table.DescribeGlobalSecondaryIndexes,
		"dynamodb:DescribeLocalSecondaryIndexes":  pc.Table.DescribeLocalSecondaryIndexes,

		// Stream Permissions
		"dynamodb:DescribeStream":   pc.Stream.DescribeStream,
		"dynamodb:GetShardIterator": pc.Stream.GetShardIterator,
		"dynamodb:GetRecords":       pc.Stream.GetRecords,
		"dynamodb:ListStreams":      pc.Stream.ListStreams,

		// Backup Permissions
		"dynamodb:CreateBackup":              pc.Backup.CreateBackup,
		"dynamodb:DeleteBackup":              pc.Backup.DeleteBackup,
		"dynamodb:DescribeBackup":            pc.Backup.DescribeBackup,
		"dynamodb:ListBackups":               pc.Backup.ListBackups,
		"dynamodb:RestoreTableFromBackup":    pc.Backup.RestoreTableFromBackup,
		"dynamodb:RestoreTableToPointInTime": pc.Backup.RestoreTableToPointInTime,
		"dynamodb:DescribeContinuousBackups": pc.Backup.DescribeContinuousBackups,
		"dynamodb:UpdateContinuousBackups":   pc.Backup.UpdateContinuousBackups,

		// Global Table Permissions
		"dynamodb:CreateGlobalTable":           pc.GlobalTable.CreateGlobalTable,
		"dynamodb:UpdateGlobalTable":           pc.GlobalTable.UpdateGlobalTable,
		"dynamodb:DescribeGlobalTable":         pc.GlobalTable.DescribeGlobalTable,
		"dynamodb:ListGlobalTables":            pc.GlobalTable.ListGlobalTables,
		"dynamodb:DescribeGlobalTableSettings": pc.GlobalTable.DescribeGlobalTableSettings,
		"dynamodb:UpdateGlobalTableSettings":   pc.GlobalTable.UpdateGlobalTableSettings,

		// Replication Permissions
		"dynamodb:CreateTableReplica": pc.Replication.CreateTableReplica,
		"dynamodb:DeleteTableReplica": pc.Replication.DeleteTableReplica,
		"dynamodb:EnableReplication":  pc.Replication.EnableReplication,
		"dynamodb:DisableReplication": pc.Replication.DisableReplication,

		// TTL Permissions
		"dynamodb:DescribeTimeToLive": pc.TTL.DescribeTimeToLive,
		"dynamodb:UpdateTimeToLive":   pc.TTL.UpdateTimeToLive,

		// Kinesis Permissions
		"dynamodb:DescribeKinesisStreamingDestination": pc.Kinesis.DescribeKinesisStreamingDestination,
		"dynamodb:EnableKinesisStreamingDestination":   pc.Kinesis.EnableKinesisStreamingDestination,
		"dynamodb:DisableKinesisStreamingDestination":  pc.Kinesis.DisableKinesisStreamingDestination,
		"dynamodb:UpdateKinesisStreamingDestination":   pc.Kinesis.UpdateKinesisStreamingDestination,

		// Contributor Insights Permissions
		"dynamodb:DescribeContributorInsights": pc.Contributors.DescribeContributorInsights,
		"dynamodb:UpdateContributorInsights":   pc.Contributors.UpdateContributorInsights,

		// Capacity Permissions
		"dynamodb:DescribeLimits":                    pc.Capacity.DescribeLimits,
		"dynamodb:DescribeReservedCapacity":          pc.Capacity.DescribeReservedCapacity,
		"dynamodb:DescribeReservedCapacityOfferings": pc.Capacity.DescribeReservedCapacityOfferings,
		"dynamodb:PurchaseReservedCapacityOfferings": pc.Capacity.PurchaseReservedCapacityOfferings,

		// Data Transfer Permissions
		"dynamodb:ExportTableToPointInTime": pc.DataTransfer.ExportTableToPointInTime,
		"dynamodb:DescribeExport":           pc.DataTransfer.DescribeExport,
		"dynamodb:ListExports":              pc.DataTransfer.ListExports,
		"dynamodb:ImportTable":              pc.DataTransfer.ImportTable,
		"dynamodb:DescribeImport":           pc.DataTransfer.DescribeImport,
		"dynamodb:ListImports":              pc.DataTransfer.ListImports,

		// Tag Permissions
		"dynamodb:TagResource":        pc.Tags.TagResource,
		"dynamodb:UntagResource":      pc.Tags.UntagResource,
		"dynamodb:ListTagsOfResource": pc.Tags.ListTagsOfResource,

		// Auto-Scaling Permissions
		"dynamodb:DescribeTableReplicaAutoScaling": pc.AutoScaling.DescribeTableReplicaAutoScaling,
		"dynamodb:UpdateTableReplicaAutoScaling":   pc.AutoScaling.UpdateTableReplicaAutoScaling,
	}
}
