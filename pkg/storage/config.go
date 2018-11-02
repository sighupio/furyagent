package storage

// Config represent a configuration for working with an object storage
type Config struct {
	Provider string `yaml:"provider"`
	S3       `yaml:"s3"`
	Local    `yaml:"local"`
}

// Local is configuration needed to save/open file from disk
type Local struct {
	Path         string `yaml:"local"`
	BackupFolder string `yaml:"backupFolder"`
}

// S3 is the configuration needed to save/open file from s3
type S3 struct {
	AccessKey  string `yaml:"aws_access_key"`
	SecretKey  string `yaml:"aws_secret_key"`
	URL        string `yaml:"url"`
	Region     string `yaml:"region"`
	BucketName string `yaml:"bucketName"`
}
