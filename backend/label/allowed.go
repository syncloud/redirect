package label

const Other = "other"

type Allowed struct {
	known map[string]bool
}

func NewAllowed(values []string) *Allowed {
	known := map[string]bool{}
	for _, value := range values {
		known[value] = true
	}
	return &Allowed{known: known}
}

func (a *Allowed) Label(value string) string {
	if a.known[value] {
		return value
	}
	return Other
}
