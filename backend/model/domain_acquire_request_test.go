package model

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestForwardCompatibleDomain_Free(t *testing.T) {
	userDomain := "test123"
	domain := &DomainAcquireRequest{DeprecatedUserDomain: &userDomain}
	domain.ForwardCompatibleDomain("syncloud.it")
	assert.Equal(t, "test123.syncloud.it", *domain.Domain)
}

func TestForwardCompatibleDomain_Free_Empty(t *testing.T) {
	domain := &DomainAcquireRequest{DeprecatedUserDomain: nil}
	domain.ForwardCompatibleDomain("syncloud.it")
	assert.Nil(t, domain.Domain)
}
