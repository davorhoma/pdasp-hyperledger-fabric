package models

type Invoice struct {
	DocType    DocType `json:"docType"`
	ID         string  `json:"id"`
	MerchantID string  `json:"merchantId"`
	UserID     string  `json:"userId"`
	ProductID  string  `json:"productId"`
	Quantity   int     `json:"quantity"`
	TotalPrice float64 `json:"totalPrice"`
	Date       string  `json:"date"`
}
