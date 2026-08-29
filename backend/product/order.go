package product

type Order struct {
	Email     string `json:"email"`
	Device    string `json:"device"`
	Option    string `json:"option"`
	Name      string `json:"name"`
	Address   string `json:"address"`
	City      string `json:"city"`
	Postcode  string `json:"postcode"`
	Country   string `json:"country"`
	Total     int    `json:"total"`
	Reference string `json:"reference"`
	PaidWith  string `json:"paid_with"`
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
