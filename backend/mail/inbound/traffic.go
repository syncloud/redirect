package inbound

type Traffic interface {
	OverLimit(domain string) bool
	Record(domain string, bytes int64)
}
