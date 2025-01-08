package conf

import "github.com/Falokut/go-kit/config"

type LocalConfig struct {
	BaseLocalStoragePath string `yaml:"base_local_storage_path" env:"BASE_LOCAL_STORAGE_PATH"`
	StorageMode          string `yaml:"storage_mode" env:"STORAGE_MODE" validate:"required,oneof=LOCAL MINIO"` // MINIO or LOCAL
	MinioConfig          struct {
		Endpoint          string `yaml:"endpoint" env:"MINIO_ENDPOINT"`
		AccessKeyID       string `yaml:"access_key_id" env:"MINIO_ACCESS_KEY_ID"`
		SecretAccessKey   string `yaml:"secret_access_key" env:"MINIO_SECRET_ACCESS_KEY"`
		Secure            bool   `yaml:"secure" env:"MINIO_SECURE"`
		Token             string `yaml:"token" env:"MINIO_TOKEN"`
		UploadFileThreads uint   `yaml:"upload_file_threads" env:"MINIO_UPLOAD_FILE_THREADS"`
	} `yaml:"minio"`
	MaxImageSizeMb     int64         `yaml:"max_file_size" env:"MAX_FILE_SIZE"` // in mb
	SupportedFileTypes []string      `yaml:"supported_file_types"`
	Listen             config.Listen `yaml:"listen"`
}
