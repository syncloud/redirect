package product

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Mail interface {
	SendDeviceOrder(order *Order, device, option string) error
	SendDeviceOrderCustomer(order *Order, device, option string) error
	SendDeviceOrderStatus(order *Order, device, option, message string) error
	SendDeviceOrderStuck(order *Order, reason string) error
}

type Store interface {
	InsertOrder(order *Order) (int64, error)
	GetUnpaidOrders(before time.Time) ([]*Order, error)
	SetOrderProviderReference(id int64, providerReference string) error
	GetOrderByReference(reference string) (*Order, error)
	GetOrderById(id int64) (*Order, error)
	MarkOrderPaid(id int64) error
	RedactOrders(userId int64) error
	GetOrdersByUser(userId int64) ([]*Order, error)
	GetAllOrders() ([]*Order, error)
	SetOrderStatus(id int64, status string) error
	MarkOrderAbandoned(id int64) error
	SetOrderProvider(id int64, provider string, providerReference string) error
	GetUnfinishedOrders(userId int64) ([]*Order, error)
	InsertOrderEvent(orderId int64, status string, comment string) error
	GetOrderEvents(orderId int64) ([]*OrderEvent, error)
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
	device, option, err := o.Describe(order.Device, order.Option)
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

func (o *Orders) Complete(userId int64, reference string) (int64, error) {
	order, err := o.store.GetOrderByReference(reference)
	if err != nil {
		return 0, err
	}
	if order == nil || order.UserId != userId {
		return 0, ErrNoOrder
	}
	return order.Id, o.Settle(order)
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
	device, option, err := o.Describe(order.Device, order.Option)
	if err != nil {
		return err
	}
	if err := o.store.InsertOrderEvent(order.Id, order.Status, ""); err != nil {
		return err
	}
	if err := o.mail.SendDeviceOrder(order, device, option); err != nil {
		return err
	}
	if order.Email == "" {
		o.logger.Error("order has no account to confirm to",
			zap.String("reference", order.Reference))
		return nil
	}
	return o.mail.SendDeviceOrderCustomer(order, device, option)
}

func (o *Orders) Redact(userId int64) error {
	return o.store.RedactOrders(userId)
}

func (o *Orders) Unpaid(before time.Time) ([]*Order, error) {
	return o.store.GetUnpaidOrders(before)
}

func (o *Orders) Describe(deviceCode, optionCode string) (string, string, error) {
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

var Statuses = []string{"ordered", "sent"}

var statusMessages = map[string]string{
	"ordered": "We have your order and are getting it ready.",
	"sent":    "Your device is on its way.",
}

func ValidStatus(status string) bool {
	for _, each := range Statuses {
		if each == status {
			return true
		}
	}
	return false
}

func (o *Orders) Mine(userId int64) ([]*Order, error) {
	return o.store.GetOrdersByUser(userId)
}

func (o *Orders) All() ([]*Order, error) {
	return o.store.GetAllOrders()
}

func (o *Orders) SetStatus(id int64, status string, comment string) error {
	if !ValidStatus(status) {
		return fmt.Errorf("%w: %s", ErrBadStatus, status)
	}
	order, err := o.store.GetOrderById(id)
	if err != nil {
		return err
	}
	if order == nil {
		return ErrNoOrder
	}
	if order.Status == status && comment == "" {
		return nil
	}
	if err := o.store.SetOrderStatus(order.Id, status); err != nil {
		return err
	}
	if err := o.store.InsertOrderEvent(order.Id, status, comment); err != nil {
		return err
	}
	order.Status = status
	device, option, err := o.Describe(order.Device, order.Option)
	if err != nil {
		return err
	}
	message := statusMessages[status]
	if comment != "" {
		message = message + "\n\n" + comment
	}
	return o.mail.SendDeviceOrderStatus(order, device, option, message)
}

func (o *Orders) Detail(id int64, userId int64, admin bool) (*Order, []*OrderEvent, error) {
	order, err := o.store.GetOrderById(id)
	if err != nil {
		return nil, nil, err
	}
	if order == nil || (!admin && order.UserId != userId) {
		return nil, nil, ErrNoOrder
	}
	events, err := o.store.GetOrderEvents(order.Id)
	if err != nil {
		return nil, nil, err
	}
	return order, events, nil
}

func (o *Orders) Unfinished(userId int64) ([]*Order, error) {
	return o.store.GetUnfinishedOrders(userId)
}

func (o *Orders) Abandon(order *Order) error {
	return o.store.MarkOrderAbandoned(order.Id)
}

func (o *Orders) Stuck(order *Order, reason string) error {
	return o.mail.SendDeviceOrderStuck(order, reason)
}

func (o *Orders) Retry(userId int64, id int64, provider string) (*Order, error) {
	order, err := o.store.GetOrderById(id)
	if err != nil {
		return nil, err
	}
	if order == nil || order.UserId != userId {
		return nil, ErrNoOrder
	}
	if order.Paid {
		return nil, ErrAlreadyPaid
	}
	if order.Abandoned {
		return nil, ErrAbandoned
	}
	checkout, err := o.checkouts.Get(provider)
	if err != nil {
		return nil, err
	}
	device, option, err := o.Describe(order.Device, order.Option)
	if err != nil {
		return nil, err
	}
	order.Provider = provider
	providerReference, url, err := checkout.Start(order, fmt.Sprintf("%s, %s", device, option))
	if err != nil {
		return nil, err
	}
	if err := o.store.SetOrderProvider(order.Id, provider, providerReference); err != nil {
		return nil, err
	}
	order.ProviderReference = providerReference
	order.Url = url
	return order, nil
}
