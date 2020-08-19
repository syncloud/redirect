package dns

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/smira/go-statsd"
	"github.com/syncloud/redirect/model"
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
	UpdateDomain(mainDomain string, domain *model.Domain) error
}

type AmazonDns struct {
	client          *route53.Route53
	statsdClient    *statsd.Client
	accessKeyId     string
	secretAccessKey string
	hostedZoneId    string
}

func New(statsdClient *statsd.Client, accessKeyId string, secretAccessKey string, hostedZoneId string) *AmazonDns {
	mySession := session.Must(session.NewSession(&aws.Config{Credentials: credentials.NewStaticCredentials(accessKeyId, secretAccessKey, "")}))
	client := route53.New(mySession)
	return &AmazonDns{
		client,
		statsdClient,
		accessKeyId,
		secretAccessKey,
		hostedZoneId,
	}
}

func (a *AmazonDns) UpdateDomain(mainDomain string, domain *model.Domain) error {
	fullDomain := domain.DnsName(mainDomain)
	ipv4 := domain.DnsIpv4()
	ipv6 := domain.DnsIpv6()
	spf := "\"v=spf1 a mx -all\""
	mx := fmt.Sprintf("1 %s", fullDomain)
	dkim := domain.DkimKey
	err := a.deleteDomain(mainDomain, domain)
	if err != nil {
		return err
	}
	err = a.actionDomain(fullDomain, ipv4, ipv6, dkim, spf, mx, "CREATE")
	if err != nil {
		return err
	}
	return nil
}

func (a *AmazonDns) deleteDomain(mainDomain string, domain *model.Domain) error {
	fullDomain := domain.DnsName(mainDomain)
	err := a.actionDomain(fullDomain, &defaultIpv4, &defaultIpv6, &defaultDkim, defaultSpf, defaultMx, "UPSERT")
	if err != nil {
		return err
	}
	err = a.actionDomain(fullDomain, &defaultIpv4, &defaultIpv6, &defaultDkim, defaultSpf, defaultMx, "DELETE")
	if err != nil {
		return err
	}
	return nil
}

func (a *AmazonDns) change(action string, name string, value string, changeType string) *route53.Change {
  fmt.Printf("dns change: %s, %s, %s, %s\n", action, name, value, changeType)
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

func (a *AmazonDns) actionDomain(domain string, ipv4 *string, ipv6 *string, dkim *string, spf string, mx string, action string) error {

	a.statsdClient.Incr("dns.ip.connect", 1)

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

	input := &route53.ChangeResourceRecordSetsInput{
		ChangeBatch:  &route53.ChangeBatch{Changes: changes},
		HostedZoneId: aws.String(a.hostedZoneId),
	}
	result, err := a.client.ChangeResourceRecordSets(input)
	if err != nil {
		a.statsdClient.Incr("dns.ip.error", 1)
		return err
	}

	fmt.Println(result)

	a.statsdClient.Incr("dns.ip.commit", 1)
	return nil
}
