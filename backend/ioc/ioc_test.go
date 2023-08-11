package ioc

import (
	"github.com/stretchr/testify/assert"
	"github.com/syncloud/redirect/rest"
	"os"
	"path"
	"testing"
)

func TestNewContainer(t *testing.T) {
	pwd, err := os.Getwd()
	assert.NoError(t, err)
	configPath := path.Join(pwd, "../../config/env/integration/config.cfg")
	secretPath := path.Join(pwd, "../../config/env/integration/secret.cfg")
	c, err := NewContainer(configPath, secretPath, "")
	assert.NoError(t, err)
	err = c.Call(func(www *rest.Www) error { return nil })
	assert.NoError(t, err)
}
