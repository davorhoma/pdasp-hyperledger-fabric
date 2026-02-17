package services

import (
	"chaincode/trading/models"
)

func CreateMerchant(id, merchantType, pib string) (*models.Merchant, error) {
	if id == "" || merchantType == "" || pib == "" {
		return nil, ErrInvalidInput
	}

	merchant := &models.Merchant{
		DocType:         models.DocTypeMerchant,
		ID:              id,
		Type:            merchantType,
		PIB:             pib,
		ProductsForSale: []string{},
		Invoices:        []string{},
		Balance:         0,
	}

	return merchant, nil
}

func AddProductsToMerchant(merchant *models.Merchant, products ...*models.Product) error {
	if merchant == nil {
		return ErrNotFound
	}

	for _, p := range products {
		if p.MerchantID != merchant.ID {
			return ErrInvalidInput
		}

		merchant.ProductsForSale = append(merchant.ProductsForSale, p.ID)
	}

	return nil
}

func DepositToMerchant(m *models.Merchant, amount float64) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}

	m.Balance += amount
	return nil
}
