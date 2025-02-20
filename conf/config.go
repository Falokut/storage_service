package conf

import (
	"github.com/Falokut/go-kit/config"
)

type LocalConfig struct {
	BaseLocalStoragePath  string        `yaml:"base_local_storage_path" env:"BASE_LOCAL_STORAGE_PATH"`
	MinioConfig           config.Minio  `yaml:"minio"`
	MaxImageSizeMb        int64         `yaml:"max_file_size" env:"MAX_FILE_SIZE"` // in mb
	SupportedFileTypes    []string      `yaml:"supported_file_types"`
	MaxRangeRequestLength int64         `yaml:"max_range_request_length" env:"MAX_RANGE_REQUEST_LENGTH"` // in mb
	Listen                config.Listen `yaml:"listen"`
	HealthcheckPort       int           `yaml:"healthcheck_port"`
}
