package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

const activeDevicesWindow = time.Hour

type Database interface {
	GetOnlineDevicesCount() (int64, error)
	GetOnlineUsersCount() (int64, error)
	GetAllUsersCount() (int64, error)
	GetActiveUsersCount() (int64, error)
	GetSubscribedUsersCount() (int64, error)
	Get2MonthOldActiveUsersWithoutDomainCount() (int64, error)
	GetDomainCount() (int64, error)
	GetActiveDevicesByPlatformVersion(window time.Duration) (map[string]int64, error)
}

type DbGauges struct {
	db            Database
	logger        *zap.Logger
	devices       *prometheus.Desc
	users         *prometheus.Desc
	domains       *prometheus.Desc
	activeDevices *prometheus.Desc
}

func NewDbGauges(db Database, logger *zap.Logger) *DbGauges {
	return &DbGauges{
		db:      db,
		logger:  logger,
		devices: prometheus.NewDesc("redirect_db_devices", "Online devices count.", nil, nil),
		users:   prometheus.NewDesc("redirect_db_users", "User counts by state.", []string{"state"}, nil),
		domains: prometheus.NewDesc("redirect_db_domains", "Total domains count.", nil, nil),
		activeDevices: prometheus.NewDesc(
			"redirect_active_devices",
			"Devices whose last /domain/update is within the active window, by platform_version.",
			[]string{"platform_version"}, nil),
	}
}

func (g *DbGauges) Describe(ch chan<- *prometheus.Desc) {
	ch <- g.devices
	ch <- g.users
	ch <- g.domains
	ch <- g.activeDevices
}

func (g *DbGauges) Collect(ch chan<- prometheus.Metric) {
	g.emit(ch, g.devices, prometheus.GaugeValue, g.db.GetOnlineDevicesCount)
	g.emitLabeled(ch, g.users, "online", g.db.GetOnlineUsersCount)
	g.emitLabeled(ch, g.users, "all", g.db.GetAllUsersCount)
	g.emitLabeled(ch, g.users, "active", g.db.GetActiveUsersCount)
	g.emitLabeled(ch, g.users, "subscribed", g.db.GetSubscribedUsersCount)
	g.emitLabeled(ch, g.users, "dead", g.db.Get2MonthOldActiveUsersWithoutDomainCount)
	g.emit(ch, g.domains, prometheus.GaugeValue, g.db.GetDomainCount)

	counts, err := g.db.GetActiveDevicesByPlatformVersion(activeDevicesWindow)
	if err != nil {
		g.logger.Warn("active devices query failed", zap.Error(err))
		return
	}
	for platformVersion, n := range counts {
		ch <- prometheus.MustNewConstMetric(g.activeDevices, prometheus.GaugeValue, float64(n), platformVersion)
	}
}

func (g *DbGauges) emit(ch chan<- prometheus.Metric, desc *prometheus.Desc, kind prometheus.ValueType, query func() (int64, error)) {
	v, err := query()
	if err != nil {
		g.logger.Warn("db gauge query failed", zap.String("metric", desc.String()), zap.Error(err))
		return
	}
	ch <- prometheus.MustNewConstMetric(desc, kind, float64(v))
}

func (g *DbGauges) emitLabeled(ch chan<- prometheus.Metric, desc *prometheus.Desc, label string, query func() (int64, error)) {
	v, err := query()
	if err != nil {
		g.logger.Warn("db gauge query failed", zap.String("metric", desc.String()), zap.String("state", label), zap.Error(err))
		return
	}
	ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, float64(v), label)
}
