package metrics

import "github.com/prometheus/client_golang/prometheus"

type Metrics struct {
	requests  *prometheus.CounterVec
	dnsClient *prometheus.CounterVec
	cleaner   *prometheus.CounterVec
	events    *prometheus.CounterVec
	orders    *prometheus.CounterVec
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
		events: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "redirect_site_event_total",
				Help: "Steps visitors reached in the account site, by event name, whether they arrived from an ad and the language the page was shown in.",
			},
			[]string{"event", "source", "language"},
		),
		orders: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "redirect_device_order_total",
				Help: "Device checkouts begun, by payment provider and whether the provider accepted the checkout.",
			},
			[]string{"provider", "outcome"},
		),
		rogue: NewRogueDevices(),
	}
}

func (m *Metrics) Request(handler string) {
	m.requests.WithLabelValues(handler).Inc()
}

func (m *Metrics) Event(event, source, language string) {
	m.events.WithLabelValues(event, source, language).Inc()
}

func (m *Metrics) DeviceOrder(provider, outcome string) {
	m.orders.WithLabelValues(provider, outcome).Inc()
}

func (m *Metrics) DnsClient(action string) {
	m.dnsClient.WithLabelValues(action).Inc()
}

func (m *Metrics) Cleaner(result string) {
	m.cleaner.WithLabelValues(result).Inc()
}

func (m *Metrics) Rogue(platformVersion, token string) {
	m.rogue.Mark(platformVersion, token)
}

func (m *Metrics) Describe(ch chan<- *prometheus.Desc) {
	m.requests.Describe(ch)
	m.dnsClient.Describe(ch)
	m.cleaner.Describe(ch)
	m.events.Describe(ch)
	m.orders.Describe(ch)
	m.rogue.Describe(ch)
}

func (m *Metrics) Collect(ch chan<- prometheus.Metric) {
	m.requests.Collect(ch)
	m.dnsClient.Collect(ch)
	m.cleaner.Collect(ch)
	m.events.Collect(ch)
	m.orders.Collect(ch)
	m.rogue.Collect(ch)
}
