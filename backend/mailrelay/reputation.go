package mailrelay

import (
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

const (
	bounceMetric    = "Reputation.BounceRate"
	complaintMetric = "Reputation.ComplaintRate"
)

type MetricSource interface {
	Rate(metric string) (float64, error)
}

// Reputation publishes the two SES rates that decide whether our account keeps
// sending. AWS pauses sending at a 10% bounce rate and reviews an account at a
// 0.1% complaint rate, so these need to be visible long before they are hit.
type Reputation struct {
	source   MetricSource
	interval time.Duration
	logger   *zap.Logger

	mutex     sync.Mutex
	bounce    float64
	complaint float64

	bounceDesc    *prometheus.Desc
	complaintDesc *prometheus.Desc
}

func NewReputation(source MetricSource, interval time.Duration, logger *zap.Logger) *Reputation {
	return &Reputation{
		source:   source,
		interval: interval,
		logger:   logger,
		bounceDesc: prometheus.NewDesc(
			"redirect_ses_bounce_rate", "SES bounce rate, sending is paused at 0.1.", nil, nil),
		complaintDesc: prometheus.NewDesc(
			"redirect_ses_complaint_rate", "SES complaint rate, account is reviewed at 0.001.", nil, nil),
	}
}

func (r *Reputation) Start() error {
	r.poll()
	go func() {
		for range time.Tick(r.interval) {
			r.poll()
		}
	}()
	return nil
}

func (r *Reputation) poll() {
	bounce, err := r.source.Rate(bounceMetric)
	if err != nil {
		r.logger.Warn("cannot read ses bounce rate", zap.Error(err))
		return
	}
	complaint, err := r.source.Rate(complaintMetric)
	if err != nil {
		r.logger.Warn("cannot read ses complaint rate", zap.Error(err))
		return
	}
	r.mutex.Lock()
	r.bounce = bounce
	r.complaint = complaint
	r.mutex.Unlock()
}

func (r *Reputation) Describe(ch chan<- *prometheus.Desc) {
	ch <- r.bounceDesc
	ch <- r.complaintDesc
}

func (r *Reputation) Collect(ch chan<- prometheus.Metric) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	ch <- prometheus.MustNewConstMetric(r.bounceDesc, prometheus.GaugeValue, r.bounce)
	ch <- prometheus.MustNewConstMetric(r.complaintDesc, prometheus.GaugeValue, r.complaint)
}

type CloudWatch struct {
	client *cloudwatch.CloudWatch
	window time.Duration
	now    func() time.Time
}

func NewCloudWatch(awsSession *session.Session, region string, window time.Duration) *CloudWatch {
	return &CloudWatch{
		client: cloudwatch.New(awsSession, aws.NewConfig().WithRegion(region)),
		window: window,
		now:    time.Now,
	}
}

func (c *CloudWatch) Rate(metric string) (float64, error) {
	end := c.now().UTC()
	output, err := c.client.GetMetricStatistics(&cloudwatch.GetMetricStatisticsInput{
		Namespace:  aws.String("AWS/SES"),
		MetricName: aws.String(metric),
		StartTime:  aws.Time(end.Add(-c.window)),
		EndTime:    aws.Time(end),
		Period:     aws.Int64(int64(c.window.Seconds())),
		Statistics: aws.StringSlice([]string{cloudwatch.StatisticMaximum}),
	})
	if err != nil {
		return 0, err
	}
	rate := 0.0
	for _, point := range output.Datapoints {
		if point.Maximum != nil && *point.Maximum > rate {
			rate = *point.Maximum
		}
	}
	return rate, nil
}
