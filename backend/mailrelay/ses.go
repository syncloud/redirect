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

func NewSesSender(region string, configuration string, logger *zap.Logger) (*SesSender, error) {
	awsSession, err := session.NewSession(&aws.Config{Region: aws.String(region)})
	if err != nil {
		return nil, err
	}
	return &SesSender{ses: ses.New(awsSession), configuration: configuration, logger: logger}, nil
}

// Send hands the message to SES as is, so the device keeps its own From header
// and DKIM signature and recipients see mail from the user's own domain.
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
