package outbound

type snsNotification struct {
	Type    string `json:"Type"`
	Message string `json:"Message"`
}
