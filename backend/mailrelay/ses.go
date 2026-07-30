package mailrelay

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
	"go.uber.org/zap"
)

type Sender interface {
	Send(from string, recipients []string, message []byte) error
}

type SesSender struct {
	ses           *ses.SES
	configuration string
	logger        *zap.Logger
}

func NewSesSender(awsSession *session.Session, region string, configuration string, logger *zap.Logger) *SesSender {
	return &SesSender{
		ses:           ses.New(awsSession, aws.NewConfig().WithRegion(region)),
		configuration: configuration,
		logger:        logger,
	}
}

func (s *SesSender) Send(from string, recipients []string, message []byte) error {
	input := &ses.SendRawEmailInput{
		Source:       aws.String(from),
		Destinations: aws.StringSlice(recipients),
		RawMessage:   &ses.RawMessage{Data: message},
	}
	if s.configuration != "" {
		input.ConfigurationSetName = aws.String(s.configuration)
	}
	_, err := s.ses.SendRawEmail(input)
	return err
}
