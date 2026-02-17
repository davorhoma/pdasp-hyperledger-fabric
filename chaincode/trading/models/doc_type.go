package models

type DocType string

const (
	DocTypeMerchant DocType = "merchant"
	DocTypeProduct  DocType = "product"
	DocTypeUser     DocType = "user"
	DocTypeInvoice  DocType = "invoice"
)
