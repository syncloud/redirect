package product

import (
	"errors"
	"time"

	"go.uber.org/zap"
)

type Settler interface {
	Unpaid(before time.Time) ([]*Order, error)
	Settle(order *Order) error
}

type Reconciler struct {
	orders   Settler
	interval time.Duration
	settle   time.Duration
	now      func() time.Time
	logger   *zap.Logger
}

func NewReconciler(orders Settler, interval, settle time.Duration, logger *zap.Logger) *Reconciler {
	return &Reconciler{
		orders:   orders,
		interval: interval,
		settle:   settle,
		now:      time.Now,
		logger:   logger,
	}
}

func (r *Reconciler) Start() error {
	go r.loop()
	return nil
}

func (r *Reconciler) loop() {
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()
	for range ticker.C {
		r.Run()
	}
}

func (r *Reconciler) Run() {
	orders, err := r.orders.Unpaid(r.now().Add(-r.settle))
	if err != nil {
		r.logger.Warn("cannot read unpaid orders", zap.Error(err))
		return
	}
	for _, order := range orders {
		err := r.orders.Settle(order)
		switch {
		case err == nil:
			r.logger.Info("order paid after all",
				zap.String("reference", order.Reference),
				zap.String("provider", order.Provider))
		case errors.Is(err, ErrNotPaid):
			r.logger.Info("order still not paid",
				zap.String("reference", order.Reference))
		default:
			r.logger.Error("cannot settle an order",
				zap.String("reference", order.Reference), zap.Error(err))
		}
	}
}
