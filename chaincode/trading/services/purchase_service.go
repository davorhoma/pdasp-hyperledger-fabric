package services

import (
	"chaincode/trading/models"
	"time"
)

func Purchase(user *models.User, product *models.Product, merchant *models.Merchant, quantity int, invoiceID string) (*models.Invoice, error) {
	if quantity <= 0 {
		return nil, ErrInvalidQuantity
	}

	if err := ReduceProductQuantity(product, quantity); err != nil {
		return nil, err
	}

	total := product.Price * float64(quantity)

	if err := WithdrawFromUser(user, total); err != nil {
		return nil, err
	}

	if err := DepositToMerchant(merchant, total); err != nil {
		return nil, err
	}

	invoice := &models.Invoice{
		DocType:    models.DocTypeInvoice,
		ID:         invoiceID,
		UserID:     user.ID,
		MerchantID: merchant.ID,
		ProductID:  product.ID,
		Quantity:   quantity,
		TotalPrice: total,
		Date:       time.Now().Format(time.RFC3339),
	}

	user.Invoices = append(user.Invoices, invoice.ID)
	merchant.Invoices = append(merchant.Invoices, invoice.ID)

	return invoice, nil
}
