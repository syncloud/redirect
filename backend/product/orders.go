package product

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Mail interface {
	SendDeviceOrder(order *Order, device, option string) error
}

type Store interface {
	InsertOrder(order *Order) (int64, error)
	GetUnpaidOrders(before time.Time) ([]*Order, error)
	SetOrderProviderReference(id int64, providerReference string) error
	GetOrderByReference(reference string) (*Order, error)
	MarkOrderPaid(id int64) error
}

type Orders struct {
	catalog   *Catalog
	checkouts *Checkouts
	store     Store
	mail      Mail
	logger    *zap.Logger
}

func NewOrders(catalog *Catalog, checkouts *Checkouts, store Store, mail Mail, logger *zap.Logger) *Orders {
	return &Orders{catalog: catalog, checkouts: checkouts, store: store, mail: mail, logger: logger}
}

func (o *Orders) Catalog() []Device {
	return o.catalog.Devices()
}

func (o *Orders) Shipping() int {
	return o.catalog.Shipping()
}

func (o *Orders) Start(order *Order, provider string) (string, error) {
	if missing := order.Missing(); len(missing) > 0 {
		return "", fmt.Errorf("the address needs %v", missing)
	}
	checkout, err := o.checkouts.Get(provider)
	if err != nil {
		return "", err
	}
	device, option, err := o.describe(order.Device, order.Option)
	if err != nil {
		return "", err
	}
	total, err := o.catalog.Total(order.Device, order.Option)
	if err != nil {
		return "", err
	}

	order.Total = total
	order.Provider = provider
	order.Reference = uuid.New().String()

	id, err := o.store.InsertOrder(order)
	if err != nil {
		return "", err
	}
	order.Id = id

	providerReference, url, err := checkout.Start(order, fmt.Sprintf("%s, %s", device, option))
	if err != nil {
		return "", err
	}
	if err := o.store.SetOrderProviderReference(id, providerReference); err != nil {
		return "", err
	}
	order.ProviderReference = providerReference
	order.Url = url
	return order.Reference, nil
}

func (o *Orders) Complete(userId int64, reference string) error {
	order, err := o.store.GetOrderByReference(reference)
	if err != nil {
		return err
	}
	if order == nil || order.UserId != userId {
		return ErrNoOrder
	}
	return o.Settle(order)
}

func (o *Orders) Settle(order *Order) error {
	if order.Paid {
		return nil
	}

	checkout, err := o.checkouts.Get(order.Provider)
	if err != nil {
		return err
	}
	paid, amount, currency, err := checkout.Paid(order.ProviderReference)
	if err != nil {
		return err
	}
	if !paid {
		return ErrNotPaid
	}
	if currency != Currency {
		o.logger.Error("wrong currency",
			zap.String("paid", currency), zap.String("expected", Currency),
			zap.String("reference", order.Reference))
		return ErrWrongCurrency
	}
	if amount != order.Total {
		o.logger.Error("wrong amount",
			zap.Int("paid", amount), zap.Int("expected", order.Total),
			zap.String("reference", order.Reference))
		return ErrWrongAmount
	}

	if err := o.store.MarkOrderPaid(order.Id); err != nil {
		return err
	}
	device, option, err := o.describe(order.Device, order.Option)
	if err != nil {
		return err
	}
	return o.mail.SendDeviceOrder(order, device, option)
}

func (o *Orders) Unpaid(before time.Time) ([]*Order, error) {
	return o.store.GetUnpaidOrders(before)
}

func (o *Orders) describe(deviceCode, optionCode string) (string, string, error) {
	for _, device := range o.catalog.Devices() {
		if device.Code != deviceCode {
			continue
		}
		for _, option := range device.Options {
			if option.Code == optionCode {
				return device.Name, option.Name, nil
			}
		}
	}
	return "", "", fmt.Errorf("no %s with %s", deviceCode, optionCode)
}
