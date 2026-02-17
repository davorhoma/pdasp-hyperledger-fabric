package services

import (
	"chaincode/trading/models"
	"time"
)

func CreateProduct(
	id string,
	name string,
	expiration string,
	price float64,
	quantity int,
	merchantID string,
	merchantType string,
) (*models.Product, error) {
	if id == "" || name == "" || merchantID == "" {
		return nil, ErrInvalidInput
	}

	if price <= 0 {
		return nil, ErrInvalidAmount
	}

	if quantity <= 0 {
		return nil, ErrInvalidQuantity
	}

	if expiration == "" {
		// default expiration date: +1 year
		expiration = time.Now().AddDate(1, 0, 0).Format(time.RFC3339)
	}

	return &models.Product{
		DocType:      models.DocTypeProduct,
		ID:           id,
		Name:         name,
		Expiration:   expiration,
		Price:        price,
		Quantity:     quantity,
		MerchantID:   merchantID,
		MerchantType: merchantType,
	}, nil
}

func AddMultipleProducts(productsData []struct {
	ID           string
	Name         string
	Expiration   string
	Price        float64
	Quantity     int
	MerchantID   string
	MerchantType string
}) ([]*models.Product, error) {
	products := make([]*models.Product, 0, len(productsData))
	for _, pd := range productsData {
		p, err := CreateProduct(pd.ID, pd.Name, pd.Expiration, pd.Price, pd.Quantity, pd.MerchantID, pd.MerchantType)
		if err != nil {
			return nil, err
		}

		products = append(products, p)
	}

	return products, nil
}

func ReduceProductQuantity(p *models.Product, quantity int) error {
	if quantity <= 0 {
		return ErrInvalidQuantity
	}

	if p.Quantity < quantity {
		return ErrInsufficientStock
	}

	p.Quantity -= quantity
	return nil
}
