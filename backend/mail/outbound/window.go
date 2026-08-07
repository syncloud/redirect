package outbound

import "time"

type window struct {
	start time.Time
	count int64
}
