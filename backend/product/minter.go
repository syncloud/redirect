package product

import "github.com/syncloud/redirect/grant"

type Minter interface {
	Mint(source, plan string, uses int) (*grant.Grant, error)
}
