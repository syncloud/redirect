package dns

import (
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/stretchr/testify/assert"
	"github.com/syncloud/redirect/log"
	"github.com/syncloud/redirect/metrics"
	"github.com/syncloud/redirect/model"
	"testing"
)

type Route53Stub struct {
	resourceRecordSetsInput *route53.ChangeResourceRecordSetsInput
	createHostedZoneInput   *route53.CreateHostedZoneInput
	deleteHostedZoneInput   *route53.DeleteHostedZoneInput
	getHostedZoneInput      *route53.GetHostedZoneInput
	waitedFor               []string
	waitError               error
}

func (r *Route53Stub) ChangeResourceRecordSets(input *route53.ChangeResourceRecordSetsInput) (*route53.ChangeResourceRecordSetsOutput, error) {
	r.resourceRecordSetsInput = input
	return &route53.ChangeResourceRecordSetsOutput{ChangeInfo: &route53.ChangeInfo{Id: aws.String("/change/C1")}}, nil
}

func (r *Route53Stub) WaitUntilResourceRecordSetsChangedWithContext(_ aws.Context, input *route53.GetChangeInput, _ ...request.WaiterOption) error {
	r.waitedFor = append(r.waitedFor, *input.Id)
	return r.waitError
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

func TestAmazonDns_CreateCertbotRecord_WaitsForChangeInSync(t *testing.T) {
	client := &Route53Stub{}
	amazonDns := New(client, metrics.New(), 10, "syncloud.it", log.Default())
	err := amazonDns.CreateCertbotRecord("zone", "_acme-challenge.device.syncloud.it", []string{"a", "b"})
	assert.Nil(t, err)
	assert.Equal(t, []string{"/change/C1"}, client.waitedFor)
}

func TestAmazonDns_CreateCertbotRecord_FailsWhenNotInSync(t *testing.T) {
	client := &Route53Stub{waitError: errors.New("timeout")}
	amazonDns := New(client, metrics.New(), 10, "syncloud.it", log.Default())
	err := amazonDns.CreateCertbotRecord("zone", "_acme-challenge.device.syncloud.it", []string{"a"})
	assert.ErrorContains(t, err, "not in sync")
}

func TestAmazonDns_CreateCertbotRecord_QuoteValue(t *testing.T) {
	client := &Route53Stub{}
	amazonDns := New(client, metrics.New(), 10, "syncloud.it", log.Default())
	err := amazonDns.CreateCertbotRecord("", "name", []string{"value"})
	assert.Nil(t, err)
	record := client.resourceRecordSetsInput.ChangeBatch.Changes[0].ResourceRecordSet.ResourceRecords[0]
	assert.Equal(t, `"value"`, *record.Value)

}

func TestAmazonDns_UpdateDomainRecords_Dkim255CharsLimit(t *testing.T) {
	client := &Route53Stub{}
	amazonDns := New(client, metrics.New(), 10, "syncloud.it", log.Default())
	dkimKey := "12345abcde"
	domain := &model.Domain{DkimKey: &dkimKey}
	err := amazonDns.UpdateDomainRecords(domain)
	assert.Nil(t, err)
	record := client.resourceRecordSetsInput.ChangeBatch.Changes[0].ResourceRecordSet.ResourceRecords[0]
	assert.Equal(t, `"v=DKIM1; k" "=rsa; p=12" "345abcde"`, *record.Value)

}

func mxRecord(client *Route53Stub) (string, error) {
	for _, change := range client.resourceRecordSetsInput.ChangeBatch.Changes {
		if *change.ResourceRecordSet.Type == "MX" {
			return *change.ResourceRecordSet.ResourceRecords[0].Value, nil
		}
	}
	return "", fmt.Errorf("no MX record was written")
}

func TestAmazonDns_UpdateDomainRecords_MxPointsAtTheDeviceWithoutRelay(t *testing.T) {
	client := &Route53Stub{}
	amazonDns := New(client, metrics.New(), 10, "syncloud.it", log.Default())

	err := amazonDns.UpdateDomainRecords(&model.Domain{Name: "alice.syncloud.it"})

	assert.Nil(t, err)
	mx, err := mxRecord(client)
	assert.Nil(t, err)
	assert.Equal(t, "1 alice.syncloud.it.", mx)
}

func TestAmazonDns_UpdateDomainRecords_MxPointsAtTheRelayWithRelay(t *testing.T) {
	client := &Route53Stub{}
	amazonDns := New(client, metrics.New(), 10, "syncloud.it", log.Default())

	err := amazonDns.UpdateDomainRecords(&model.Domain{Name: "alice.syncloud.it", Relay: true})

	assert.Nil(t, err)
	mx, err := mxRecord(client)
	assert.Nil(t, err)
	assert.Equal(t, "1 alice.mx.syncloud.it.", mx)
}

func TestAmazonDns_UpdateDomainRecords_MxForACustomDomain(t *testing.T) {
	client := &Route53Stub{}
	amazonDns := New(client, metrics.New(), 10, "syncloud.it", log.Default())

	err := amazonDns.UpdateDomainRecords(&model.Domain{Name: "example.com", Relay: true})

	assert.Nil(t, err)
	mx, err := mxRecord(client)
	assert.Nil(t, err)
	assert.Equal(t, "1 example-com.mx.syncloud.it.", mx)
}
