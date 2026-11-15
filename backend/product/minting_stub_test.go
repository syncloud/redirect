package product

import (
	"fmt"

	"github.com/syncloud/redirect/grant"
)

type mintingStub struct {
	minted []*grant.Grant
	err    error
}

func (m *mintingStub) Mint(source, plan string, uses int) (*grant.Grant, error) {
	if m.err != nil {
		return nil, m.err
	}
	made := &grant.Grant{
		Id:       int64(len(m.minted) + 1),
		Code:     fmt.Sprintf("CODE-%d", len(m.minted)+1),
		Plan:     plan,
		Months:   12,
		UsesLeft: uses,
		Source:   source,
	}
	m.minted = append(m.minted, made)
	return made, nil
}
