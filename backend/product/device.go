package product

type Device struct {
	Code    string   `json:"code"`
	Name    string   `json:"name"`
	Board   string   `json:"board"`
	Price   int      `json:"price"`
	Specs   []Spec   `json:"specs"`
	Options []Option `json:"options"`
}

type Spec struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}
