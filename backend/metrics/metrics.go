package metrics

import "github.com/prometheus/client_golang/prometheus"

type Metrics struct {
	requests  *prometheus.CounterVec
	dnsClient *prometheus.CounterVec
	cleaner   *prometheus.CounterVec
	download  *prometheus.CounterVec
	rogue     *RogueDevices
}

func New() *Metrics {
	return &Metrics{
		requests: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "redirect_requests_total",
				Help: "HTTP handler invocations on the redirect api/www surface, by handler.",
			},
			[]string{"handler"},
		),
		dnsClient: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "redirect_dns_client_actions_total",
				Help: "Route53 client actions, by action (connect|commit|error).",
			},
			[]string{"action"},
		),
		cleaner: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "redirect_cleaner_actions_total",
				Help: "Domain cleaner actions, by result (delete|error).",
			},
			[]string{"result"},
		),
		download: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "redirect_download_total",
				Help: "Device image download links followed, by board and whether the visitor arrived from an ad.",
			},
			[]string{"board", "source"},
		),
		rogue: NewRogueDevices(),
	}
}

func (m *Metrics) Request(handler string) {
	m.requests.WithLabelValues(handler).Inc()
}

func (m *Metrics) DnsClient(action string) {
	m.dnsClient.WithLabelValues(action).Inc()
}

func (m *Metrics) Cleaner(result string) {
	m.cleaner.WithLabelValues(result).Inc()
}

func (m *Metrics) Download(board, source string) {
	m.download.WithLabelValues(board, source).Inc()
}

func (m *Metrics) Rogue(platformVersion, token string) {
	m.rogue.Mark(platformVersion, token)
}

func (m *Metrics) Describe(ch chan<- *prometheus.Desc) {
	m.requests.Describe(ch)
	m.dnsClient.Describe(ch)
	m.cleaner.Describe(ch)
	m.download.Describe(ch)
	m.rogue.Describe(ch)
}

func (m *Metrics) Collect(ch chan<- prometheus.Metric) {
	m.requests.Collect(ch)
	m.dnsClient.Collect(ch)
	m.cleaner.Collect(ch)
	m.download.Collect(ch)
	m.rogue.Collect(ch)
}
