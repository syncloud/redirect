package service

import (
	"fmt"
	"sort"
	"strings"

	"github.com/syncloud/redirect/model"
)

type NsCheckerDb interface {
	GetDomainByName(name string) (*model.Domain, error)
}

type NsCheckerDns interface {
	GetHostedZoneNameServers(id string) ([]*string, error)
}

type NsCheckerResolver interface {
	LookupNameServers(domain string) ([]string, error)
}

type NsChecker struct {
	db               NsCheckerDb
	amazonDns        NsCheckerDns
	resolver         NsCheckerResolver
	freeHostedZoneId string
}

func NewNsChecker(db NsCheckerDb, amazonDns NsCheckerDns, resolver NsCheckerResolver, freeHostedZoneId string) *NsChecker {
	return &NsChecker{
		db:               db,
		amazonDns:        amazonDns,
		resolver:         resolver,
		freeHostedZoneId: freeHostedZoneId,
	}
}

func (c *NsChecker) Check(userId int64, domainName string) (*model.NameServerCheckResult, error) {
	domain, err := c.db.GetDomainByName(domainName)
	if err != nil {
		return nil, err
	}
	if domain == nil || domain.UserId != userId {
		return nil, fmt.Errorf("not found")
	}
	if c.freeHostedZoneId == domain.HostedZoneId {
		return &model.NameServerCheckResult{Matched: true}, nil
	}

	expectedPtrs, err := c.amazonDns.GetHostedZoneNameServers(domain.HostedZoneId)
	if err != nil {
		return nil, err
	}
	expected := make([]string, 0, len(expectedPtrs))
	for _, e := range expectedPtrs {
		if e != nil {
			expected = append(expected, normalizeNs(*e))
		}
	}
	sort.Strings(expected)

	actual, err := c.resolver.LookupNameServers(domainName)
	if err != nil {
		return &model.NameServerCheckResult{
			Matched:  false,
			Expected: expected,
			Error:    err.Error(),
		}, nil
	}
	actualNorm := make([]string, 0, len(actual))
	for _, a := range actual {
		actualNorm = append(actualNorm, normalizeNs(a))
	}
	sort.Strings(actualNorm)

	return &model.NameServerCheckResult{
		Matched:  stringSlicesEqual(expected, actualNorm),
		Expected: expected,
		Actual:   actualNorm,
	}, nil
}

func normalizeNs(s string) string {
	return strings.ToLower(strings.TrimSuffix(s, "."))
}

func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
