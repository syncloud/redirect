package mailrelay

type sesEvent struct {
	NotificationType string `json:"notificationType"`
	EventType        string `json:"eventType"`
	Bounce           *struct {
		BounceType        string `json:"bounceType"`
		BouncedRecipients []struct {
			EmailAddress string `json:"emailAddress"`
		} `json:"bouncedRecipients"`
	} `json:"bounce"`
	Mail struct {
		Source string `json:"source"`
	} `json:"mail"`
}
