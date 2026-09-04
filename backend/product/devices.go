package product

const Shipping = 1500

func Devices() []Device {
	return []Device{{
		Code:  "h4",
		Name:  "Syncloud H4",
		Board: "odroid-hc4",
		Price: 22900,
		Specs: []Spec{
			{Name: "CPU", Value: "Amlogic S905X3, Cortex-A55"},
			{Name: "RAM", Value: "4 GB DDR4"},
			{Name: "Drive bays", Value: "2 x SATA, 3.5 or 2.5 inch"},
			{Name: "Ethernet", Value: "1 Gb"},
			{Name: "CPU cores", Value: "4"},
			{Name: "CPU frequency", Value: "1.8 GHz"},
			{Name: "Boot media", Value: "Micro SD card, included"},
			{Name: "USB", Value: "USB 2.0 x 1"},
			{Name: "Video", Value: "HDMI 2.0, 4K at 60 Hz"},
			{Name: "Size", Value: "84 x 90.5 x 25 mm"},
		},
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
