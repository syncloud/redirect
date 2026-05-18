package dns

import (
	"context"
	"net"
	"sort"
	"strings"
	"time"
)

type PublicResolver struct {
	resolver *net.Resolver
}

func NewPublicResolver() *PublicResolver {
	r := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, _ string) (net.Conn, error) {
			d := net.Dialer{Timeout: 5 * time.Second}
			return d.DialContext(ctx, network, "8.8.8.8:53")
		},
	}
	return &PublicResolver{resolver: r}
}

func (p *PublicResolver) LookupNameServers(domain string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	nss, err := p.resolver.LookupNS(ctx, domain)
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(nss))
	for _, ns := range nss {
		out = append(out, strings.ToLower(strings.TrimSuffix(ns.Host, ".")))
	}
	sort.Strings(out)
	return out, nil
}
