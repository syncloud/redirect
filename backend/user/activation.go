package user

import (
	"github.com/syncloud/redirect/model"
	"go.uber.org/zap"
	"time"
)

const (
	ActivationActionTypeId = 1
	ActivationMaxAttempts  = 5
	ActivationBatchSize    = 20
	ActivationPollInterval = 2 * time.Second
)

type ActivationDatabase interface {
	GetPendingActivations(actionTypeId uint64, maxAttempts int, limit int) ([]*model.PendingActivation, error)
	MarkActionSent(actionId uint64, now time.Time) error
	IncrementActionAttempts(actionId uint64) error
}

type ActivationMail interface {
	SendActivate(to string, token string) error
}

type ActivationSender struct {
	database ActivationDatabase
	mail     ActivationMail
	enabled  bool
	interval time.Duration
	logger   *zap.Logger
}

func NewActivationSender(
	database ActivationDatabase,
	mail ActivationMail,
	enabled bool,
	logger *zap.Logger) *ActivationSender {
	return &ActivationSender{
		database: database,
		mail:     mail,
		enabled:  enabled,
		interval: ActivationPollInterval,
		logger:   logger,
	}
}

func (s *ActivationSender) Start() error {
	if !s.enabled {
		s.logger.Warn("activation sender is disabled")
		return nil
	}
	go func() {
		for {
			sent, err := s.SendPending(time.Now())
			if err != nil {
				s.logger.Error("unable to send activations", zap.Error(err))
			}
			if sent > 0 {
				s.logger.Info("activations sent", zap.Int("count", sent))
			}
			time.Sleep(s.interval)
		}
	}()
	return nil
}

func (s *ActivationSender) SendPending(now time.Time) (int, error) {
	pending, err := s.database.GetPendingActivations(
		ActivationActionTypeId, ActivationMaxAttempts, ActivationBatchSize)
	if err != nil {
		return 0, err
	}
	sent := 0
	for _, p := range pending {
		err := s.mail.SendActivate(p.Email, p.Token)
		if err != nil {
			s.logger.Error("unable to send activation",
				zap.Uint64("action", p.ActionId),
				zap.Int("attempts", p.Attempts+1),
				zap.Error(err))
			if err := s.database.IncrementActionAttempts(p.ActionId); err != nil {
				return sent, err
			}
			continue
		}
		if err := s.database.MarkActionSent(p.ActionId, now); err != nil {
			return sent, err
		}
		sent++
	}
	return sent, nil
}
