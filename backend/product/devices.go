package product

const Shipping = 1500

func Devices() []Device {
	return []Device{{
		Code:  "h4",
		Name:  "Syncloud H4",
		Board: "odroid-hc4",
		Price: 22900,
		Options: []Option{
			{Code: "120", Name: "120 GB SSD", Extra: 0},
			{Code: "120x2", Name: "120 GB SSD x 2", Extra: 3000},
			{Code: "1t", Name: "1 TB SSD", Extra: 8000},
			{Code: "1tx2", Name: "1 TB SSD x 2", Extra: 18000},
			{Code: "2t", Name: "2 TB SSD", Extra: 20000},
			{Code: "2tx2", Name: "2 TB SSD x 2", Extra: 43000},
		},
	}}
}
