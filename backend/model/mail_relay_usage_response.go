package model

type MailRelayUsageResponse struct {
	Enabled       bool  `json:"enabled"`
	UsedMessages  int64 `json:"used_messages"`
	LimitMessages int64 `json:"limit_messages"`
}
