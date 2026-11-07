package product

const Shipping = 1500

func Devices() []Device {
	return []Device{{
		Code:  "x1",
		Name:  "Syncloud X1",
		Board: "amd64",
		Price: 58200,
		Specs: []Spec{
			{Name: "Base", Value: "GMKtec NucBox G10"},
			{Name: "CPU", Value: "AMD Ryzen 5 3500U"},
			{Name: "CPU cores", Value: "4 cores, 8 threads"},
			{Name: "CPU frequency", Value: "2.1 GHz, up to 3.7 GHz"},
			{Name: "RAM", Value: "16 GB DDR4, up to 64 GB"},
			{Name: "Disk", Value: "1 TB NVMe SSD"},
			{Name: "Disk slots", Value: "2 x M.2 2280, PCIe 3.0 x4"},
			{Name: "Ethernet", Value: "2.5 Gb"},
			{Name: "USB", Value: "USB 3.2 and USB-C"},
			{Name: "Video", Value: "Three displays, 4K"},
			{Name: "Wireless", Value: "Wi-Fi 5, Bluetooth 5.0"},
			{Name: "Size", Value: "127 x 127 x 45 mm"},
		},
		Options: []Option{
			{Code: "1t", Name: "1 TB NVMe SSD", Extra: 0},
		},
	}}
}
