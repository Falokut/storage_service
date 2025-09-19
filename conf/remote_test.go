package conf_test

import (
	"github.com/Falokut/go-kit/test/rct"
	"storage-service/conf"
	"testing"
)

func TestDefaultRemoteConfig(t *testing.T) {
	t.Parallel()
	rct.Test(t, "default_remote_config.json", conf.Remote{})
}
