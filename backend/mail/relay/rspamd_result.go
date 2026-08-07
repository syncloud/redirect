package relay

type rspamdResult struct {
	Action string  `json:"action"`
	Score  float64 `json:"score"`
}
