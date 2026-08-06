package mailin

import "github.com/syncloud/redirect/model"

type DomainStore interface {
	GetDomainByName(name string) (*model.Domain, error)
}
