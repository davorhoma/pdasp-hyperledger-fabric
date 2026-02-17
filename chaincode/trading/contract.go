package trading

import (
	"chaincode/trading/models"
	"chaincode/trading/services"
	"encoding/json"

	"github.com/hyperledger/fabric-contract-api-go/v2/contractapi"
)

type TradingContract struct {
	contractapi.Contract
}

func (t *TradingContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	merchant1, _ := services.CreateMerchant("MERCHANT1", "supermarket", "123456789")
	merchant2, _ := services.CreateMerchant("MERCHANT2", "auto_parts", "987654321")

	product1, _ := services.CreateProduct("PROD1", "Mleko", "2026-12-31T23:59:59Z", 50, 10, merchant1.ID, merchant1.Type)
	product2, _ := services.CreateProduct("PROD2", "Hleb", "2026-11-15T23:59:59Z", 20, 15, merchant1.ID, merchant1.Type)
	product3, _ := services.CreateProduct("PROD3", "Kocnica", "2026-10-03T23:59:59Z", 150, 5, merchant2.ID, merchant2.Type)
	product4, _ := services.CreateProduct("PROD4", "Filter ulja", "2026-10-10T23:59:59Z", 80, 8, merchant2.ID, merchant2.Type)

	_ = services.AddProductsToMerchant(merchant1, product1, product2)
	_ = services.AddProductsToMerchant(merchant2, product3, product4)

	user1, _ := services.CreateUser("USER1", "Marko", "Markovic", "marko@example.com")
	user2, _ := services.CreateUser("USER2", "Jelena", "Jovanovic", "jelena@example.com")

	_ = services.DepositToEntity(user1, 500)
	_ = services.DepositToEntity(user2, 300)
	_ = services.DepositToEntity(merchant1, 1000)
	_ = services.DepositToEntity(merchant2, 1000)

	entities := []struct {
		key  string
		data interface{}
	}{
		{"MERCHANT_" + merchant1.ID, merchant1},
		{"MERCHANT_" + merchant2.ID, merchant2},
		{"PRODUCT_" + product1.ID, product1},
		{"PRODUCT_" + product2.ID, product2},
		{"PRODUCT_" + product3.ID, product3},
		{"PRODUCT_" + product4.ID, product4},
		{"USER_" + user1.ID, user1},
		{"USER_" + user2.ID, user2},
	}

	for _, e := range entities {
		bytes, _ := json.Marshal(e.data)
		if err := ctx.GetStub().PutState(e.key, bytes); err != nil {
			return err
		}
	}

	return nil
}

func (t *TradingContract) CreateMerchant(ctx contractapi.TransactionContextInterface, id, merchantType, pib string) error {
	merchant, err := services.CreateMerchant(id, merchantType, pib)
	if err != nil {
		return err
	}

	key := "MERCHANT_" + merchant.ID
	bytes, _ := json.Marshal(merchant)
	return ctx.GetStub().PutState(key, bytes)
}

func (t *TradingContract) AddProducts(ctx contractapi.TransactionContextInterface, merchantID string, productsData []models.Product) error {
	merchantBytes, err := ctx.GetStub().GetState("MERCHANT_" + merchantID)
	if err != nil || merchantBytes == nil {
		return services.ErrNotFound
	}

	var merchant models.Merchant
	_ = json.Unmarshal(merchantBytes, &merchant)

	var products []*models.Product
	for _, pd := range productsData {
		p, err := services.CreateProduct(pd.ID, pd.Name, pd.Expiration, pd.Price, pd.Quantity, merchantID, merchant.Type)
		if err != nil {
			return err
		}

		products = append(products, p)

		key := "PRODUCT_" + p.ID
		bytes, _ := json.Marshal(p)
		if err := ctx.GetStub().PutState(key, bytes); err != nil {
			return err
		}
	}

	if err := services.AddProductsToMerchant(&merchant, products...); err != nil {
		return err
	}

	updatedMerchantBytes, _ := json.Marshal(merchant)
	return ctx.GetStub().PutState("MERCHANT_"+merchant.ID, updatedMerchantBytes)
}

func (t *TradingContract) CreateUser(ctx contractapi.TransactionContextInterface, id, firstName, lastName, email string) error {
	user, err := services.CreateUser(id, firstName, lastName, email)
	if err != nil {
		return err
	}

	key := "USER_" + user.ID
	bytes, _ := json.Marshal(user)
	return ctx.GetStub().PutState(key, bytes)
}

func (t *TradingContract) Purchase(ctx contractapi.TransactionContextInterface,
	userID, productID, invoiceID string, quantity int) error {

	userBytes, err := ctx.GetStub().GetState("USER_" + userID)
	if err != nil || userBytes == nil {
		return services.ErrNotFound
	}

	var user models.User
	_ = json.Unmarshal(userBytes, &user)

	productBytes, err := ctx.GetStub().GetState("PRODUCT_" + productID)
	if err != nil || productBytes == nil {
		return services.ErrNotFound
	}

	var product models.Product
	_ = json.Unmarshal(productBytes, &product)

	merchantBytes, err := ctx.GetStub().GetState("MERCHANT_" + product.MerchantID)
	if err != nil || merchantBytes == nil {
		return services.ErrNotFound
	}

	var merchant models.Merchant
	_ = json.Unmarshal(merchantBytes, &merchant)

	invoice, err := services.Purchase(&user, &product, &merchant, quantity, invoiceID)
	if err != nil {
		return err
	}

	if err := ctx.GetStub().PutState("USER_"+user.ID, mustMarshal(user)); err != nil {
		return err
	}
	if err := ctx.GetStub().PutState("PRODUCT_"+product.ID, mustMarshal(product)); err != nil {
		return err
	}
	if err := ctx.GetStub().PutState("MERCHANT_"+merchant.ID, mustMarshal(merchant)); err != nil {
		return err
	}
	if err := ctx.GetStub().PutState("INVOICE_"+invoice.ID, mustMarshal(invoice)); err != nil {
		return err
	}

	return nil
}

func mustMarshal(v interface{}) []byte {
	b, _ := json.Marshal(v)
	return b
}

func (t *TradingContract) Deposit(ctx contractapi.TransactionContextInterface,
	entityType, id string, amount float64) error {

	switch entityType {
	case "user":
		userBytes, err := ctx.GetStub().GetState("USER_" + id)
		if err != nil || userBytes == nil {
			return services.ErrNotFound
		}

		var user models.User
		_ = json.Unmarshal(userBytes, &user)
		if err := services.DepositToEntity(&user, amount); err != nil {
			return err
		}

		return ctx.GetStub().PutState("USER_"+user.ID, mustMarshal(user))

	case "merchant":
		merchantBytes, err := ctx.GetStub().GetState("MERCHANT_" + id)
		if err != nil || merchantBytes == nil {
			return services.ErrNotFound
		}

		var merchant models.Merchant
		_ = json.Unmarshal(merchantBytes, &merchant)
		if err := services.DepositToEntity(&merchant, amount); err != nil {
			return err
		}

		return ctx.GetStub().PutState("MERCHANT_"+merchant.ID, mustMarshal(merchant))

	default:
		return services.ErrInvalidInput
	}
}
