package models

type User struct {
	DocType   DocType  `json:"docType"`
	ID        string   `json:"id"`
	FirstName string   `json:"firstName"`
	LastName  string   `json:"lastName"`
	Email     string   `json:"email"`
	Invoices  []string `json:"invoices"`
	Balance   float64  `json:"balance"`
}
