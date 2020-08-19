package dns

import (
	"github.com/syncloud/redirect/model"
)

type FakeDns struct{}

func (*FakeDns) UpdateDomain(mainDomain string, domain *model.Domain) error {
	return nil
}
