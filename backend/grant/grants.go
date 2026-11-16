package grant

type Grants struct {
	store  Store
	months int
}

func NewGrants(store Store, months int) *Grants {
	return &Grants{store: store, months: months}
}

func (g *Grants) Mint(source, plan string, uses int) (*Grant, error) {
	code, err := NewCode()
	if err != nil {
		return nil, err
	}
	minted := &Grant{
		Code:     code,
		Plan:     plan,
		Months:   g.months,
		UsesLeft: uses,
		Source:   source,
	}
	id, err := g.store.InsertGrant(minted)
	if err != nil {
		return nil, err
	}
	minted.Id = id
	return minted, nil
}
