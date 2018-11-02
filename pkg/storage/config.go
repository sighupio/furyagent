package storage

// Config represent a configuration for working with an object storage
type Config struct {
	Provider string
	S3
	Local
}

// Local is configuration needed to save/open file from disk
type Local struct {
	Path         string
	BackupFolder string
}

// S3 is the configuration needed to save/open file from s3
type S3 struct {
	AccessKey  string
	SecretKey  string
	URL        string
	Region     string
	BucketName string
}
