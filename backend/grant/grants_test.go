package grant

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/syncloud/redirect/model"
)

func TestMintMakesACodeNobodyHasToRead(t *testing.T) {
	store := newStoreStub()

	minted, err := NewGrants(store, 12).Mint("order:42", model.PlanPro, 1)

	assert.NoError(t, err)
	assert.Equal(t, 12, minted.Months)
	assert.Equal(t, 1, minted.UsesLeft)
	assert.Equal(t, "order:42", minted.Source)
	assert.NotContains(t, minted.Code, "O")
	assert.NotContains(t, minted.Code, "0")
	assert.NotContains(t, minted.Code, "I")
	assert.NotContains(t, minted.Code, "1")
	assert.Len(t, minted.Code, codeGroups*codeGroupSize+codeGroups-1)
}

func TestMintedCodesDoNotRepeat(t *testing.T) {
	store := newStoreStub()
	service := NewGrants(store, 12)
	seen := map[string]bool{}

	for i := 0; i < 200; i++ {
		minted, err := service.Mint("order:1", model.PlanPro, 1)
		assert.NoError(t, err)
		assert.False(t, seen[minted.Code])
		seen[minted.Code] = true
	}
}
