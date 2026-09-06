package product

import "fmt"

type Catalog struct {
	devices  []Device
	shipping int
}

func NewCatalog(devices []Device, shipping int) *Catalog {
	return &Catalog{devices: devices, shipping: shipping}
}

func (c *Catalog) Devices() []Device {
	return c.devices
}

func (c *Catalog) Shipping() int {
	return c.shipping
}

func (c *Catalog) Total(deviceCode, optionCode string) (int, error) {
	for _, device := range c.devices {
		if device.Code != deviceCode {
			continue
		}
		for _, option := range device.Options {
			if option.Code == optionCode {
				return device.Price + option.Extra + c.shipping, nil
			}
		}
		return 0, fmt.Errorf("%s has no option %s", deviceCode, optionCode)
	}
	return 0, fmt.Errorf("no device %s", deviceCode)
}
