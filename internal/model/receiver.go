package model

// ReceiverResponse represents the response from the dummy receiver service.
type ReceiverResponse struct {
	UserAgent     string `json:"user_agent"`
	IPAddress     string `json:"ip_address"`
	ForwardedHost string `json:"forwarded_host"`
}
