package user

import (
	"github.com/stretchr/testify/assert"
	"github.com/syncloud/redirect/log"
	"path"
	"testing"
)

func TestCleanerState_Get_NoStateYet_0(t *testing.T) {
	tempDir := t.TempDir()
	state := &CleanerState{
		file:   path.Join(tempDir, "state"),
		logger: log.Default(),
	}
	id, err := state.Get()
	assert.NoError(t, err)
	assert.Equal(t, int64(0), id)
}

func TestCleanerState_Set(t *testing.T) {
	tempDir := t.TempDir()
	state := &CleanerState{
		file:   path.Join(tempDir, "state"),
		logger: log.Default(),
	}
	err := state.Set(123)
	assert.NoError(t, err)

	id, err := state.Get()
	assert.NoError(t, err)
	assert.Equal(t, int64(123), id)
}
