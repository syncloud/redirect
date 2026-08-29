package product

import (
	"fmt"

	"go.uber.org/zap"
)

type Mail interface {
	SendDeviceOrder(order *Order, device, option string) error
}

type Orders struct {
	catalog   *Catalog
	checkouts *Checkouts
	mail      Mail
	logger    *zap.Logger
}

func NewOrders(catalog *Catalog, checkouts *Checkouts, mail Mail, logger *zap.Logger) *Orders {
	return &Orders{catalog: catalog, checkouts: checkouts, mail: mail, logger: logger}
}

func (o *Orders) Start(order *Order, provider string) (string, error) {
	if missing := order.Missing(); len(missing) > 0 {
		return "", fmt.Errorf("the address needs %v", missing)
	}
	checkout, err := o.checkouts.Get(provider)
	if err != nil {
		return "", err
	}
	device, option, err := o.describe(order)
	if err != nil {
		return "", err
	}
	total, err := o.catalog.Total(order.Device, order.Option)
	if err != nil {
		return "", err
	}
	order.Total = total
	order.PaidWith = provider

	return checkout.Start(order, fmt.Sprintf("%s, %s", device, option))
}

func (o *Orders) Complete(order *Order, provider, reference string) error {
	checkout, err := o.checkouts.Get(provider)
	if err != nil {
		return err
	}
	total, err := o.catalog.Total(order.Device, order.Option)
	if err != nil {
		return err
	}
	paid, amount, err := checkout.Paid(reference)
	if err != nil {
		return err
	}
	if !paid {
		return ErrNotPaid
	}
	if amount != total {
		o.logger.Error("wrong amount",
			zap.Int("paid", amount), zap.Int("expected", total), zap.String("reference", reference))
		return ErrWrongAmount
	}

	device, option, err := o.describe(order)
	if err != nil {
		return err
	}
	order.Total = total
	order.PaidWith = provider
	order.Reference = reference
	return o.mail.SendDeviceOrder(order, device, option)
}

func (o *Orders) describe(order *Order) (string, string, error) {
	for _, device := range o.catalog.Devices() {
		if device.Code != order.Device {
			continue
		}
		for _, option := range device.Options {
			if option.Code == order.Option {
				return device.Name, option.Name, nil
			}
		}
	}
	return "", "", fmt.Errorf("no %s with %s", order.Device, order.Option)
}
