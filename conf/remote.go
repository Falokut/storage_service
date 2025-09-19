package conf

import (
	"reflect"

	"github.com/Falokut/go-kit/db"
	"github.com/Falokut/go-kit/log"
	"github.com/Falokut/go-kit/miniox"
	"github.com/Falokut/go-kit/remote/schema"
)

// nolint: gochecknoinits
func init() {
	schema.CustomGenerators.Register("logLevel", func(field reflect.StructField, t *schema.Schema) {
		t.Type = "string"
		t.Enum = []any{"debug", "info", "warn", "error", "fatal"}
	})
}

type Remote struct {
	LogLevel           log.Level     `schemaGen:"logLevel" schema:"Уровень логирования"`
	DB                 db.Config     `schema:"Настройка подключения к db"`
	Minio              miniox.Config `schema:"Настройка подключения к minio"`
	MaxFileSizeMb      int64         `schema:"Максимальный размер файла, в мегабайтах" validate:"required,gte=1"`
	SupportedFileTypes []string      `schema:"Разрешённые content-type файлов, если пустой, разрешены все"`
	Pending            Pending       `schema:"Настройка воркера"`
}

type Pending struct {
	FileLifetimeInMin int `schema:"Время, через которое незакоммиченный файл удаляется, в минутах" validate:"required,gte=1"`
	MaxFilesToDelete  int `schema:"Максимальное количество файлов для удаления за 1 срабатывание джобы" validate:"required,gte=1"`
}
