package model

type DeviceOrderRequest struct {
	Device   string `json:"device"`
	Option   string `json:"option"`
	Provider string `json:"provider"`
	Name     string `json:"name"`
	Address  string `json:"address"`
	City     string `json:"city"`
	Postcode string `json:"postcode"`
	Country  string `json:"country"`
}

type DeviceOrderResponse struct {
	Reference         string `json:"reference"`
	ProviderReference string `json:"provider_reference"`
	Total             int    `json:"total"`
}

type DeviceOrderCompleteRequest struct {
	Reference string `json:"reference"`
}
