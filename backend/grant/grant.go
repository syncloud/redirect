package grant

type Grant struct {
	Id       int64
	Code     string
	Plan     string
	Months   int
	UsesLeft int
	Source   string
}

func (g *Grant) Available() bool {
	return g.UsesLeft > 0
}
