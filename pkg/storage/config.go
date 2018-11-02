package storage

// Config represent a configuration for working with an object storage
type Config struct {
	Provider   string `yml:"provider"`
	AccessKey  string `yml:"aws_access_key"`
	SecretKey  string `yml:"aws_secret_key"`
	URL        string `yml:"url"`
	Region     string `yml:"region"`
	BucketName string `yml:"bucketName"`
	Local
}

// Local is configuration needed to save/open file from disk
type Local struct {
	Path         string `yml:"path"`
	BackupFolder string `yml:"backupFolder"`
}

// S3 is the configuration needed to save/open file from s3
type S3 struct {
	AccessKey  string `yml:"aws_access_key"`
	SecretKey  string `yml:"aws_secret_key"`
	URL        string `yml:"url"`
	Region     string `yml:"region"`
	BucketName string `yml:"bucketName"`
}
