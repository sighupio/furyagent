package storage

// Config represent a configuration for working with an object storage
type Config struct {
	Provider   string `mapstructure:"provider"`
	AccessKey  string `mapstructure:"aws_access_key"`
	SecretKey  string `mapstructure:"aws_secret_key"`
	URL        string `mapstructure:"url"`
	Region     string `mapstructure:"region"`
	BucketName string `mapstructure:"bucketName"`
	LocalPath  string `mapstructure:"path"`
}

// S3 is the configuration needed to save/open file from s3
type S3 struct {
	AccessKey  string `mapstructure:"aws_access_key"`
	SecretKey  string `mapstructure:"aws_secret_key"`
	URL        string `mapstructure:"url"`
	Region     string `mapstructure:"region"`
	BucketName string `mapstructure:"bucketName"`
}
