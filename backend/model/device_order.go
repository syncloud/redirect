package model

type DeviceCatalogResponse struct {
	Devices        interface{} `json:"devices"`
	Shipping       int         `json:"shipping"`
	Currency       string      `json:"currency"`
	PayPalClientId string      `json:"paypal_client_id"`
	PayPalSdkUrl   string      `json:"paypal_sdk_url"`
}

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
	Url               string `json:"url"`
	Total             int    `json:"total"`
}

type DeviceOrderCompleteRequest struct {
	Reference string `json:"reference"`
}

type DeviceOrderStatusRequest struct {
	Reference string `json:"reference"`
	Status    string `json:"status"`
}

type DeviceOrderView struct {
	Reference string `json:"reference"`
	Device    string `json:"device"`
	Option    string `json:"option"`
	Total     string `json:"total"`
	Status    string `json:"status"`
	Ordered   string `json:"ordered"`
	Email     string `json:"email,omitempty"`
	Name      string `json:"name,omitempty"`
	Address   string `json:"address,omitempty"`
	City      string `json:"city,omitempty"`
	Postcode  string `json:"postcode,omitempty"`
	Country   string `json:"country,omitempty"`
}
