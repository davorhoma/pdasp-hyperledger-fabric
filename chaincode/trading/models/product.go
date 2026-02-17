package models

type Product struct {
	DocType      DocType `json:"docType"`
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Expiration   string  `json:"expiration,omitempty"`
	Price        float64 `json:"price"`
	Quantity     int     `json:"quantity"`
	MerchantID   string  `json:"merchantId"`
	MerchantType string  `json:"merchantType"`
}
