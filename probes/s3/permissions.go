package s3

type ReadPermissions struct {
	GetObject               bool
	GetObjectVersion        bool
	GetObjectTagging        bool
	GetObjectVersionTagging bool
	GetObjectAttributes     bool
	GetObjectRetention      bool
	GetObjectLegalHold      bool
	GetObjectTorrent        bool
	ListBucket              bool
	ListBucketVersions      bool
	GetBucketLocation       bool
	GetObjectAcl            bool
	GetObjectVersionAcl     bool
}

type WritePermissions struct {
	PutObject                bool
	PutObjectTagging         bool
	PutObjectVersionTagging  bool
	PutObjectRetention       bool
	PutObjectLegalHold       bool
	PutObjectAcl             bool
	PutObjectVersionAcl      bool
	RestoreObject            bool
	ListMultipartUploadParts bool
}

type DeletePermissions struct {
	DeleteObject               bool
	DeleteObjectVersion        bool
	DeleteObjectTagging        bool
	DeleteObjectVersionTagging bool
	AbortMultipartUpload       bool
	BypassGovernanceRetention  bool
}

type BucketConfigPermissions struct {
	CreateBucket               bool
	DeleteBucket               bool
	ListBucketMultipartUploads bool
	GetBucketVersioning        bool
	PutBucketVersioning        bool
	GetBucketRequestPayment    bool
	PutBucketRequestPayment    bool
	GetBucketLogging           bool
	PutBucketLogging           bool
	GetBucketNotification      bool
	PutBucketNotification      bool
	GetBucketWebsite           bool
	PutBucketWebsite           bool
	DeleteBucketWebsite        bool
	GetBucketCORS              bool
	PutBucketCORS              bool
	DeleteBucketCORS           bool
	GetBucketTagging           bool
	PutBucketTagging           bool
	DeleteBucketTagging        bool
	GetBucketOwnershipControls bool
	PutBucketOwnershipControls bool
}

type BucketPolicyPermissions struct {
	GetBucketPolicy       bool
	PutBucketPolicy       bool
	DeleteBucketPolicy    bool
	GetBucketPolicyStatus bool
	GetBucketAcl          bool
	PutBucketAcl          bool
}

type SecurityPermissions struct {
	GetEncryptionConfiguration  bool
	PutEncryptionConfiguration  bool
	GetBucketPublicAccessBlock  bool
	PutBucketPublicAccessBlock  bool
	GetAccountPublicAccessBlock bool
	PutAccountPublicAccessBlock bool
}

type LifecyclePermissions struct {
	GetLifecycleConfiguration             bool
	PutLifecycleConfiguration             bool
	GetReplicationConfiguration           bool
	PutReplicationConfiguration           bool
	DeleteBucketReplication               bool
	GetIntelligentTieringConfiguration    bool
	PutIntelligentTieringConfiguration    bool
	DeleteIntelligentTieringConfiguration bool
	GetStorageLensConfiguration           bool
	PutStorageLensConfiguration           bool
	DeleteStorageLensConfiguration        bool
	GetStorageLensConfigurationTagging    bool
	PutStorageLensConfigurationTagging    bool
	DeleteStorageLensConfigurationTagging bool
}

type AnalyticsPermissions struct {
	GetAnalyticsConfiguration    bool
	PutAnalyticsConfiguration    bool
	DeleteAnalyticsConfiguration bool
	GetMetricsConfiguration      bool
	PutMetricsConfiguration      bool
	DeleteMetricsConfiguration   bool
	GetInventoryConfiguration    bool
	PutInventoryConfiguration    bool
	DeleteInventoryConfiguration bool
	GetStorageLensDashboard      bool
}

type AccessPointPermissions struct {
	CreateAccessPoint                          bool
	DeleteAccessPoint                          bool
	GetAccessPoint                             bool
	ListAccessPoints                           bool
	GetAccessPointPolicy                       bool
	PutAccessPointPolicy                       bool
	DeleteAccessPointPolicy                    bool
	GetAccessPointPolicyStatus                 bool
	GetAccessPointConfigurationForObjectLambda bool
	PutAccessPointConfigurationForObjectLambda bool
	CreateMultiRegionAccessPoint               bool
	DeleteMultiRegionAccessPoint               bool
	GetMultiRegionAccessPoint                  bool
	ListMultiRegionAccessPoints                bool
	GetMultiRegionAccessPointPolicy            bool
	PutMultiRegionAccessPointPolicy            bool
	GetMultiRegionAccessPointPolicyStatus      bool
	GetMultiRegionAccessPointRoutes            bool
	SubmitMultiRegionAccessPointRoutes         bool
}

type ObjectLockPermissions struct {
	GetObjectLockConfiguration       bool
	PutObjectLockConfiguration       bool
	GetBucketObjectLockConfiguration bool
	PutBucketObjectLockConfiguration bool
}

type BatchPermissions struct {
	CreateJob         bool
	DescribeJob       bool
	ListJobs          bool
	UpdateJobPriority bool
	UpdateJobStatus   bool
}

type AccountPermissions struct {
	ListAllMyBuckets bool
	HeadBucket       bool
}

func prepareActionsMap(pc *PermissionsCheck) map[string]bool {
	return map[string]bool{
		// Read Permissions
		"s3:GetObject":               pc.Read.GetObject,
		"s3:GetObjectVersion":        pc.Read.GetObjectVersion,
		"s3:GetObjectTagging":        pc.Read.GetObjectTagging,
		"s3:GetObjectVersionTagging": pc.Read.GetObjectVersionTagging,
		"s3:GetObjectAttributes":     pc.Read.GetObjectAttributes,
		"s3:GetObjectRetention":      pc.Read.GetObjectRetention,
		"s3:GetObjectLegalHold":      pc.Read.GetObjectLegalHold,
		"s3:GetObjectTorrent":        pc.Read.GetObjectTorrent,
		"s3:ListBucket":              pc.Read.ListBucket,
		"s3:ListBucketVersions":      pc.Read.ListBucketVersions,
		"s3:GetBucketLocation":       pc.Read.GetBucketLocation,
		"s3:GetObjectAcl":            pc.Read.GetObjectAcl,
		"s3:GetObjectVersionAcl":     pc.Read.GetObjectVersionAcl,

		// Write Permissions
		"s3:PutObject":                pc.Write.PutObject,
		"s3:PutObjectTagging":         pc.Write.PutObjectTagging,
		"s3:PutObjectVersionTagging":  pc.Write.PutObjectVersionTagging,
		"s3:PutObjectRetention":       pc.Write.PutObjectRetention,
		"s3:PutObjectLegalHold":       pc.Write.PutObjectLegalHold,
		"s3:PutObjectAcl":             pc.Write.PutObjectAcl,
		"s3:PutObjectVersionAcl":      pc.Write.PutObjectVersionAcl,
		"s3:RestoreObject":            pc.Write.RestoreObject,
		"s3:ListMultipartUploadParts": pc.Write.ListMultipartUploadParts,

		// Delete Permissions
		"s3:DeleteObject":               pc.Delete.DeleteObject,
		"s3:DeleteObjectVersion":        pc.Delete.DeleteObjectVersion,
		"s3:DeleteObjectTagging":        pc.Delete.DeleteObjectTagging,
		"s3:DeleteObjectVersionTagging": pc.Delete.DeleteObjectVersionTagging,
		"s3:AbortMultipartUpload":       pc.Delete.AbortMultipartUpload,
		"s3:BypassGovernanceRetention":  pc.Delete.BypassGovernanceRetention,

		// Bucket Config Permissions
		"s3:CreateBucket":               pc.BucketConfig.CreateBucket,
		"s3:DeleteBucket":               pc.BucketConfig.DeleteBucket,
		"s3:ListBucketMultipartUploads": pc.BucketConfig.ListBucketMultipartUploads,
		"s3:GetBucketVersioning":        pc.BucketConfig.GetBucketVersioning,
		"s3:PutBucketVersioning":        pc.BucketConfig.PutBucketVersioning,
		"s3:GetBucketRequestPayment":    pc.BucketConfig.GetBucketRequestPayment,
		"s3:PutBucketRequestPayment":    pc.BucketConfig.PutBucketRequestPayment,
		"s3:GetBucketLogging":           pc.BucketConfig.GetBucketLogging,
		"s3:PutBucketLogging":           pc.BucketConfig.PutBucketLogging,
		"s3:GetBucketNotification":      pc.BucketConfig.GetBucketNotification,
		"s3:PutBucketNotification":      pc.BucketConfig.PutBucketNotification,
		"s3:GetBucketWebsite":           pc.BucketConfig.GetBucketWebsite,
		"s3:PutBucketWebsite":           pc.BucketConfig.PutBucketWebsite,
		"s3:DeleteBucketWebsite":        pc.BucketConfig.DeleteBucketWebsite,
		"s3:GetBucketCORS":              pc.BucketConfig.GetBucketCORS,
		"s3:PutBucketCORS":              pc.BucketConfig.PutBucketCORS,
		"s3:DeleteBucketCORS":           pc.BucketConfig.DeleteBucketCORS,
		"s3:GetBucketTagging":           pc.BucketConfig.GetBucketTagging,
		"s3:PutBucketTagging":           pc.BucketConfig.PutBucketTagging,
		"s3:DeleteBucketTagging":        pc.BucketConfig.DeleteBucketTagging,
		"s3:GetBucketOwnershipControls": pc.BucketConfig.GetBucketOwnershipControls,
		"s3:PutBucketOwnershipControls": pc.BucketConfig.PutBucketOwnershipControls,

		// Bucket Policy Permissions
		"s3:GetBucketPolicy":       pc.BucketPolicy.GetBucketPolicy,
		"s3:PutBucketPolicy":       pc.BucketPolicy.PutBucketPolicy,
		"s3:DeleteBucketPolicy":    pc.BucketPolicy.DeleteBucketPolicy,
		"s3:GetBucketPolicyStatus": pc.BucketPolicy.GetBucketPolicyStatus,
		"s3:GetBucketAcl":          pc.BucketPolicy.GetBucketAcl,
		"s3:PutBucketAcl":          pc.BucketPolicy.PutBucketAcl,

		// Security Permissions
		"s3:GetEncryptionConfiguration":  pc.Security.GetEncryptionConfiguration,
		"s3:PutEncryptionConfiguration":  pc.Security.PutEncryptionConfiguration,
		"s3:GetBucketPublicAccessBlock":  pc.Security.GetBucketPublicAccessBlock,
		"s3:PutBucketPublicAccessBlock":  pc.Security.PutBucketPublicAccessBlock,
		"s3:GetAccountPublicAccessBlock": pc.Security.GetAccountPublicAccessBlock,
		"s3:PutAccountPublicAccessBlock": pc.Security.PutAccountPublicAccessBlock,

		// Lifecycle Permissions
		"s3:GetLifecycleConfiguration":             pc.Lifecycle.GetLifecycleConfiguration,
		"s3:PutLifecycleConfiguration":             pc.Lifecycle.PutLifecycleConfiguration,
		"s3:GetReplicationConfiguration":           pc.Lifecycle.GetReplicationConfiguration,
		"s3:PutReplicationConfiguration":           pc.Lifecycle.PutReplicationConfiguration,
		"s3:DeleteBucketReplication":               pc.Lifecycle.DeleteBucketReplication,
		"s3:GetIntelligentTieringConfiguration":    pc.Lifecycle.GetIntelligentTieringConfiguration,
		"s3:PutIntelligentTieringConfiguration":    pc.Lifecycle.PutIntelligentTieringConfiguration,
		"s3:DeleteIntelligentTieringConfiguration": pc.Lifecycle.DeleteIntelligentTieringConfiguration,
		"s3:GetStorageLensConfiguration":           pc.Lifecycle.GetStorageLensConfiguration,
		"s3:PutStorageLensConfiguration":           pc.Lifecycle.PutStorageLensConfiguration,
		"s3:DeleteStorageLensConfiguration":        pc.Lifecycle.DeleteStorageLensConfiguration,
		"s3:GetStorageLensConfigurationTagging":    pc.Lifecycle.GetStorageLensConfigurationTagging,
		"s3:PutStorageLensConfigurationTagging":    pc.Lifecycle.PutStorageLensConfigurationTagging,
		"s3:DeleteStorageLensConfigurationTagging": pc.Lifecycle.DeleteStorageLensConfigurationTagging,

		// Analytics Permissions
		"s3:GetAnalyticsConfiguration":    pc.Analytics.GetAnalyticsConfiguration,
		"s3:PutAnalyticsConfiguration":    pc.Analytics.PutAnalyticsConfiguration,
		"s3:DeleteAnalyticsConfiguration": pc.Analytics.DeleteAnalyticsConfiguration,
		"s3:GetMetricsConfiguration":      pc.Analytics.GetMetricsConfiguration,
		"s3:PutMetricsConfiguration":      pc.Analytics.PutMetricsConfiguration,
		"s3:DeleteMetricsConfiguration":   pc.Analytics.DeleteMetricsConfiguration,
		"s3:GetInventoryConfiguration":    pc.Analytics.GetInventoryConfiguration,
		"s3:PutInventoryConfiguration":    pc.Analytics.PutInventoryConfiguration,
		"s3:DeleteInventoryConfiguration": pc.Analytics.DeleteInventoryConfiguration,
		"s3:GetStorageLensDashboard":      pc.Analytics.GetStorageLensDashboard,

		// Access Point Permissions
		"s3:CreateAccessPoint":                          pc.AccessPoint.CreateAccessPoint,
		"s3:DeleteAccessPoint":                          pc.AccessPoint.DeleteAccessPoint,
		"s3:GetAccessPoint":                             pc.AccessPoint.GetAccessPoint,
		"s3:ListAccessPoints":                           pc.AccessPoint.ListAccessPoints,
		"s3:GetAccessPointPolicy":                       pc.AccessPoint.GetAccessPointPolicy,
		"s3:PutAccessPointPolicy":                       pc.AccessPoint.PutAccessPointPolicy,
		"s3:DeleteAccessPointPolicy":                    pc.AccessPoint.DeleteAccessPointPolicy,
		"s3:GetAccessPointPolicyStatus":                 pc.AccessPoint.GetAccessPointPolicyStatus,
		"s3:GetAccessPointConfigurationForObjectLambda": pc.AccessPoint.GetAccessPointConfigurationForObjectLambda,
		"s3:PutAccessPointConfigurationForObjectLambda": pc.AccessPoint.PutAccessPointConfigurationForObjectLambda,
		"s3:CreateMultiRegionAccessPoint":               pc.AccessPoint.CreateMultiRegionAccessPoint,
		"s3:DeleteMultiRegionAccessPoint":               pc.AccessPoint.DeleteMultiRegionAccessPoint,
		"s3:GetMultiRegionAccessPoint":                  pc.AccessPoint.GetMultiRegionAccessPoint,
		"s3:ListMultiRegionAccessPoints":                pc.AccessPoint.ListMultiRegionAccessPoints,
		"s3:GetMultiRegionAccessPointPolicy":            pc.AccessPoint.GetMultiRegionAccessPointPolicy,
		"s3:PutMultiRegionAccessPointPolicy":            pc.AccessPoint.PutMultiRegionAccessPointPolicy,
		"s3:GetMultiRegionAccessPointPolicyStatus":      pc.AccessPoint.GetMultiRegionAccessPointPolicyStatus,
		"s3:GetMultiRegionAccessPointRoutes":            pc.AccessPoint.GetMultiRegionAccessPointRoutes,
		"s3:SubmitMultiRegionAccessPointRoutes":         pc.AccessPoint.SubmitMultiRegionAccessPointRoutes,

		// Object Lock Permissions
		"s3:GetObjectLockConfiguration":       pc.ObjectLock.GetObjectLockConfiguration,
		"s3:PutObjectLockConfiguration":       pc.ObjectLock.PutObjectLockConfiguration,
		"s3:GetBucketObjectLockConfiguration": pc.ObjectLock.GetBucketObjectLockConfiguration,
		"s3:PutBucketObjectLockConfiguration": pc.ObjectLock.PutBucketObjectLockConfiguration,

		// Batch Permissions
		"s3:CreateJob":         pc.Batch.CreateJob,
		"s3:DescribeJob":       pc.Batch.DescribeJob,
		"s3:ListJobs":          pc.Batch.ListJobs,
		"s3:UpdateJobPriority": pc.Batch.UpdateJobPriority,
		"s3:UpdateJobStatus":   pc.Batch.UpdateJobStatus,

		// Account Permissions
		"s3:ListAllMyBuckets": pc.Account.ListAllMyBuckets,
		"s3:HeadBucket":       pc.Account.HeadBucket,
	}
}
