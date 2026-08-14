package inbound

import (
	"context"
	"net"
	"strings"
	"time"
)

type Resolver interface {
	Name(address string) string
}

type PtrResolver struct {
	resolver *net.Resolver
	timeout  time.Duration
}

func NewPtrResolver(timeout time.Duration) *PtrResolver {
	return &PtrResolver{resolver: net.DefaultResolver, timeout: timeout}
}

func (r *PtrResolver) Name(address string) string {
	if address == "" {
		return ""
	}
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()
	names, err := r.resolver.LookupAddr(ctx, address)
	if err != nil || len(names) == 0 {
		return ""
	}
	return strings.TrimSuffix(names[0], ".")
}
