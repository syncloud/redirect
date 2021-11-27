package dns

import (
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/smira/go-statsd"
	"github.com/stretchr/testify/assert"
	"testing"
)

type StatsdClientStub struct {
}

func (n StatsdClientStub) Incr(_ string, _ int64, _ ...statsd.Tag) {
}

type Route53Stub struct {
	resourceRecordSetsInput *route53.ChangeResourceRecordSetsInput
	createHostedZoneInput   *route53.CreateHostedZoneInput
	deleteHostedZoneInput   *route53.DeleteHostedZoneInput
	getHostedZoneInput      *route53.GetHostedZoneInput
}

func (r *Route53Stub) ChangeResourceRecordSets(input *route53.ChangeResourceRecordSetsInput) (*route53.ChangeResourceRecordSetsOutput, error) {
	r.resourceRecordSetsInput = input
	return nil, nil
}

func (r *Route53Stub) CreateHostedZone(input *route53.CreateHostedZoneInput) (*route53.CreateHostedZoneOutput, error) {
	r.createHostedZoneInput = input
	return nil, nil
}

func (r *Route53Stub) DeleteHostedZone(input *route53.DeleteHostedZoneInput) (*route53.DeleteHostedZoneOutput, error) {
	r.deleteHostedZoneInput = input
	return nil, nil
}

func (r *Route53Stub) GetHostedZone(input *route53.GetHostedZoneInput) (*route53.GetHostedZoneOutput, error) {
	r.getHostedZoneInput = input
	return nil, nil
}

func TestAmazonDns_CreateCertbotRecord_QuoteValue(t *testing.T) {
	client := &Route53Stub{}
	amazonDns := New(&StatsdClientStub{}, client)
	err := amazonDns.CreateCertbotRecord("", "name", []string{"value"})
	assert.Nil(t, err)
	record := client.resourceRecordSetsInput.ChangeBatch.Changes[0].ResourceRecordSet.ResourceRecords[0]
	assert.Equal(t, `"value"`, *record.Value)

}
