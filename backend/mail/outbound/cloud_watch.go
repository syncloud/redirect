package outbound

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
)

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
