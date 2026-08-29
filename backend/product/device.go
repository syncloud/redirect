package product

type Device struct {
	Code    string   `json:"code"`
	Name    string   `json:"name"`
	Board   string   `json:"board"`
	Price   int      `json:"price"`
	Options []Option `json:"options"`
}
