package product

type Order struct {
	Id                int64
	UserId            int64
	Email             string
	Device            string
	Option            string
	Total             int
	Provider          string
	Reference         string
	ProviderReference string
	Name              string
	Address           string
	City              string
	Postcode          string
	Country           string
	Paid              bool
	Url               string
}

func (o *Order) Missing() []string {
	missing := []string{}
	for _, field := range []struct {
		name  string
		value string
	}{
		{"name", o.Name},
		{"address", o.Address},
		{"city", o.City},
		{"postcode", o.Postcode},
		{"country", o.Country},
	} {
		if field.value == "" {
			missing = append(missing, field.name)
		}
	}
	return missing
}
