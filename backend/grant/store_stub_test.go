package grant

import (
	"errors"
	"time"
)

type storeStub struct {
	grants      map[string]*Grant
	redemptions map[int64]*Redemption
	nextId      int64
	claims      int
	released    int
	deleted     []int64
}

func newStoreStub() *storeStub {
	return &storeStub{
		grants:      map[string]*Grant{},
		redemptions: map[int64]*Redemption{},
		nextId:      1,
	}
}

func (s *storeStub) add(code string, months, uses int) *Grant {
	g := &Grant{Id: s.nextId, Code: code, Plan: "pro", Months: months, UsesLeft: uses, Source: "test"}
	s.nextId++
	s.grants[code] = g
	return g
}

func (s *storeStub) InsertGrant(g *Grant) (int64, error) {
	g.Id = s.nextId
	s.nextId++
	s.grants[g.Code] = g
	return g.Id, nil
}

func (s *storeStub) GetGrantByCode(code string) (*Grant, error) {
	found, ok := s.grants[code]
	if !ok {
		return nil, nil
	}
	return found, nil
}

func (s *storeStub) ClaimGrantUse(id int64) (bool, error) {
	for _, g := range s.grants {
		if g.Id != id {
			continue
		}
		if g.UsesLeft < 1 {
			return false, nil
		}
		g.UsesLeft--
		s.claims++
		return true, nil
	}
	return false, errors.New("no grant")
}

func (s *storeStub) ReleaseGrantUse(id int64) error {
	for _, g := range s.grants {
		if g.Id == id {
			g.UsesLeft++
			s.released++
			return nil
		}
	}
	return errors.New("no grant")
}

func (s *storeStub) InsertRedemption(r *Redemption) (int64, error) {
	r.Id = s.nextId
	s.nextId++
	s.redemptions[r.Id] = r
	return r.Id, nil
}

func (s *storeStub) DeleteRedemption(id int64) error {
	delete(s.redemptions, id)
	s.deleted = append(s.deleted, id)
	return nil
}

func (s *storeStub) GetRedemption(id int64) (*Redemption, error) {
	found, ok := s.redemptions[id]
	if !ok {
		return nil, nil
	}
	return found, nil
}

func (s *storeStub) addRedemption(id int64, endsAt time.Time) {
	s.redemptions[id] = &Redemption{Id: id, GrantId: 1, UserId: 7, Plan: "pro", EndsAt: endsAt}
}
