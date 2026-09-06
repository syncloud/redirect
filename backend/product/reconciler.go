package product

import (
	"errors"
	"time"

	"go.uber.org/zap"
)

type Settler interface {
	Unpaid(before time.Time) ([]*Order, error)
	Settle(order *Order) error
	Abandon(order *Order) error
	Stuck(order *Order, reason string) error
}

type Reconciler struct {
	orders   Settler
	interval time.Duration
	settle   time.Duration
	expire   time.Duration
	now      func() time.Time
	logger   *zap.Logger
}

func NewReconciler(orders Settler, interval, settle, expire time.Duration, logger *zap.Logger) *Reconciler {
	return &Reconciler{
		orders:   orders,
		interval: interval,
		settle:   settle,
		expire:   expire,
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
	giveUpBefore := r.now().Add(-r.expire)
	for _, order := range orders {
		err := r.orders.Settle(order)
		switch {
		case err == nil:
			r.logger.Info("order paid after all",
				zap.Int64("order", order.Id),
				zap.String("provider", order.Provider))
		case errors.Is(err, ErrNotPaid):
			if order.CreatedAt.After(giveUpBefore) {
				r.logger.Info("order still not paid", zap.Int64("order", order.Id))
				continue
			}
			if err := r.orders.Abandon(order); err != nil {
				r.logger.Error("cannot abandon an order",
					zap.Int64("order", order.Id), zap.Error(err))
				continue
			}
			r.logger.Info("order abandoned, it was never paid", zap.Int64("order", order.Id))
		default:
			r.logger.Error("cannot settle an order",
				zap.Int64("order", order.Id), zap.Error(err))
			if err := r.orders.Stuck(order, err.Error()); err != nil {
				r.logger.Error("cannot report a stuck order",
					zap.Int64("order", order.Id), zap.Error(err))
			}
		}
	}
}
