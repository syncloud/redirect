package model

type RelayUsageResponse struct {
	Enabled    bool  `json:"enabled"`
	UsedBytes  int64 `json:"used_bytes"`
	LimitBytes int64 `json:"limit_bytes"`
}
