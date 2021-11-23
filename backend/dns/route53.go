package dns

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/syncloud/redirect/metrics"
	"github.com/syncloud/redirect/model"
	"github.com/syncloud/redirect/utils"
	"strings"
  "log"
)

var defaultIpv4 string
var defaultIpv6 string
var defaultDkim string

const (
	defaultMx  = "1 mx"
	defaultSpf = "\"v=spf1 -all\""
)

func init() {
	defaultIpv4 = "127.0.0.1"
	defaultIpv6 = "fe80::"
	defaultDkim = "none"
}

type Dns interface {
	CreateHostedZone(domain string) (*string, error)
	DeleteHostedZone(hostedZoneId string) error
	CreateCertbotRecord(hostedZoneId string, name string, value string) error
	DeleteCertbotRecord(hostedZoneId string, name string, value string) error
	UpdateDomainRecords(domain *model.Domain) error
	DeleteDomainRecords(domain *model.Domain) error
	GetHostedZoneNameServers(id string) ([]*string, error)
}

type Route53 interface {
	ChangeResourceRecordSets(input *route53.ChangeResourceRecordSetsInput) (*route53.ChangeResourceRecordSetsOutput, error)
	CreateHostedZone(input *route53.CreateHostedZoneInput) (*route53.CreateHostedZoneOutput, error)
	DeleteHostedZone(input *route53.DeleteHostedZoneInput) (*route53.DeleteHostedZoneOutput, error)
	GetHostedZone(input *route53.GetHostedZoneInput) (*route53.GetHostedZoneOutput, error)
}

type AmazonDns struct {
	client       Route53
	statsdClient metrics.StatsdClient
}

func New(statsdClient metrics.StatsdClient, client Route53) *AmazonDns {
	return &AmazonDns{
		client,
		statsdClient,
	}
}

func (a *AmazonDns) CreateHostedZone(domain string) (*string, error) {
	fmt.Printf("create hosted zone for: %v\n", domain)
	zone, err := a.client.CreateHostedZone(&route53.CreateHostedZoneInput{
		CallerReference:  aws.String(utils.Uuid()),
		HostedZoneConfig: &route53.HostedZoneConfig{PrivateZone: aws.Bool(false)},
		Name:             aws.String(domain),
	})
	if err != nil {
		return nil, err
	}
	id := strings.ReplaceAll(*zone.HostedZone.Id, "/hostedzone/", "")
	fmt.Printf("created hosted zone id: %s\n", id)
	return &id, nil
}

func (a *AmazonDns) GetHostedZoneNameServers(id string) ([]*string, error) {
	fmt.Printf("get hosted zone for: %v\n", id)
	zone, err := a.client.GetHostedZone(&route53.GetHostedZoneInput{
		Id: aws.String(id),
	})
	if err != nil {
		return nil, err
	}
	fmt.Printf("created hosted zone id: %s\n", id)
	return zone.DelegationSet.NameServers, nil
}

func (a *AmazonDns) DeleteHostedZone(hostedZoneId string) error {
	fmt.Printf("delete hosted zone: %s\n", hostedZoneId)
	output, err := a.client.DeleteHostedZone(&route53.DeleteHostedZoneInput{
		Id: aws.String(hostedZoneId),
	})
	if err != nil {
		return err
	}
	fmt.Printf("deleted hosted zone output: %v\n", output)
	return nil
}

func (a *AmazonDns) UpdateDomainRecords(domain *model.Domain) error {
	err := a.DeleteDomainRecords(domain)
	if err != nil {
		return err
	}
	err = a.actionDomain(domain.FQDN(), domain.DnsIpv4(), domain.DnsIpv6(), domain.DkimKey, "\"v=spf1 a mx -all\"", fmt.Sprintf("1 %s", domain.FQDN()), "CREATE", domain.HostedZoneId)
	if err != nil {
		return err
	}
	return nil
}

func (a *AmazonDns) DeleteDomainRecords(domain *model.Domain) error {
	err := a.actionDomain(domain.FQDN(), &defaultIpv4, &defaultIpv6, &defaultDkim, defaultSpf, defaultMx, "UPSERT", domain.HostedZoneId)
	if err != nil {
		return err
	}
	err = a.actionDomain(domain.FQDN(), &defaultIpv4, &defaultIpv6, &defaultDkim, defaultSpf, defaultMx, "DELETE", domain.HostedZoneId)
	if err != nil {
		return err
	}
	return nil
}

func (a *AmazonDns) CreateCertbotRecord(hostedZoneId string, name string, value string) error {
  log.Println("certbot txt name ", name)
	return a.commit([]*route53.Change{
		a.change("UPSERT", name, `"`+value+`"`, "TXT"),
	}, hostedZoneId)
}

func (a *AmazonDns) DeleteCertbotRecord(hostedZoneId string, name string, value string) error {
	return a.commit([]*route53.Change{
		a.change("DELETE", name, `"`+value+`"`, "TXT"),
	}, hostedZoneId)
}

func (a *AmazonDns) change(action string, name string, value string, changeType string) *route53.Change {
	return &route53.Change{
		Action: aws.String(action),
		ResourceRecordSet: &route53.ResourceRecordSet{
			Name: aws.String(name),
			ResourceRecords: []*route53.ResourceRecord{
				{
					Value: aws.String(value),
				},
			},
			TTL:  aws.Int64(600),
			Type: aws.String(changeType),
		},
	}
}

func (a *AmazonDns) changeA(ip string, domain string, action string) *route53.Change {
	return a.change(action, domain, ip, "A")
}

func (a *AmazonDns) changeAAAA(ip string, domain string, action string) *route53.Change {
	return a.change(action, domain, ip, "AAAA")
}

func (a *AmazonDns) changeDKIM(domain string, dkim string, action string) *route53.Change {
	name := fmt.Sprintf("mail._domainkey.%s", domain)
	dkimValue := fmt.Sprintf("\"v=DKIM1; k=rsa; p=%s\"", dkim)
	return a.change(action, name, dkimValue, "TXT")
}

func (a *AmazonDns) actionDomain(domain string, ipv4 *string, ipv6 *string, dkim *string, spf string, mx string, action string, hostedZoneId string) error {

	var changes []*route53.Change

	if ipv6 != nil {
		changes = append(changes, a.changeAAAA(*ipv6, domain, action))
		changes = append(changes, a.changeAAAA(*ipv6, fmt.Sprintf("*.%s", domain), action))
	}
	if ipv4 != nil {
		changes = append(changes, a.changeA(*ipv4, domain, action))
		changes = append(changes, a.changeA(*ipv4, fmt.Sprintf("*.%s", domain), action))
	}
	if dkim != nil {
		changes = append(changes, a.changeDKIM(domain, *dkim, action))
	}
	changes = append(changes, a.change(action, domain, mx, "MX"))
	changes = append(changes, a.change(action, domain, spf, "SPF"))
	changes = append(changes, a.change(action, domain, spf, "TXT"))

	err := a.commit(changes, hostedZoneId)
	return err
}

func (a *AmazonDns) commit(changes []*route53.Change, hostedZoneId string) error {
	a.statsdClient.Incr("dns.client.connect", 1)
	input := &route53.ChangeResourceRecordSetsInput{
		ChangeBatch:  &route53.ChangeBatch{Changes: changes},
		HostedZoneId: aws.String(hostedZoneId),
	}
	_, err := a.client.ChangeResourceRecordSets(input)
	if err != nil {
		a.statsdClient.Incr("dns.client.error", 1)
		return err
	}

	a.statsdClient.Incr("dns.client.commit", 1)
	return nil
}
