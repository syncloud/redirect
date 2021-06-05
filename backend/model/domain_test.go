package model

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBackwardCompatibleDomain_Free(t *testing.T) {
	domain := &Domain{Domain: "test123.syncloud.it"}
	domain.BackwardCompatibleDomain("syncloud.it")
	assert.Equal(t, "test123", domain.DeprecatedUserDomain)
}

func TestBackwardCompatibleDomain_Managed(t *testing.T) {
	domain := &Domain{Domain: "test123.com"}
	domain.BackwardCompatibleDomain("syncloud.it")
	assert.Equal(t, "", domain.DeprecatedUserDomain)
}
