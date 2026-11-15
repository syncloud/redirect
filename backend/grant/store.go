package grant

type Store interface {
	InsertGrant(g *Grant) (int64, error)
	GetGrantByCode(code string) (*Grant, error)
	ClaimGrantUse(id int64) (bool, error)
	ReleaseGrantUse(id int64) error
	InsertRedemption(r *Redemption) (int64, error)
	DeleteRedemption(id int64) error
	GetRedemption(id int64) (*Redemption, error)
}
