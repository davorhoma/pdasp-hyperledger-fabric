package models

type Merchant struct {
	DocType         DocType  `json:"docType"`
	ID              string   `json:"id"`
	Type            string   `json:"type"`
	PIB             string   `json:"pib"`
	ProductsForSale []string `json:"products"`
	Invoices        []string `json:"invoices"`
	Balance         float64  `json:"balance"`
}
