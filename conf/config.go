package conf

import (
	"github.com/Falokut/go-kit/config"
)

type LocalConfig struct {
	MinioConfig           config.Minio  `yaml:"minio"`
	MaxFileSizeMb         int64         `validate:"required,gte=1" yaml:"max_file_size" env:"MAX_FILE_SIZE"` // in mb
	SupportedFileTypes    []string      `yaml:"supported_file_types"`
	MaxRangeRequestLength int64         `validate:"required" yaml:"max_range_request_length" env:"MAX_RANGE_REQUEST_LENGTH"` // in kb
	Listen                config.Listen `yaml:"listen"`
	HealthcheckPort       int           `yaml:"healthcheck_port"`
}
