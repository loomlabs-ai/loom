package models

type ConnectProviderReq struct {
	ProviderName string `json:"name"`
	BaseURL      string `json:"base_url"`
	APIKey       string `json:"api_key"`
}
