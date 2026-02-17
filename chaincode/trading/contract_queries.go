package trading

import (
	"chaincode/trading/models"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hyperledger/fabric-contract-api-go/v2/contractapi"
)

type ProductFilter struct {
	ID           string   `json:"id,omitempty"`
	Name         string   `json:"name,omitempty"`
	MerchantType string   `json:"merchantType,omitempty"`
	PriceMin     *float64 `json:"priceMin,omitempty"`
	PriceMax     *float64 `json:"priceMax,omitempty"`
}

func (t *TradingContract) RichQueryProducts(ctx contractapi.TransactionContextInterface, filterJSON string) ([]*models.Product, error) {
	var filter ProductFilter
	if err := json.Unmarshal([]byte(filterJSON), &filter); err != nil {
		return nil, fmt.Errorf("cannot parse filter JSON: %v", err)
	}

	selector := make(map[string]interface{})
	selector["docType"] = "product"

	if filter.ID != "" {
		selector["id"] = filter.ID
	}
	if filter.Name != "" {
		selector["name"] = map[string]string{"$regex": "(?i)" + strings.ReplaceAll(filter.Name, " ", ".*")}
	}
	if filter.MerchantType != "" {
		selector["merchantType"] = filter.MerchantType
	}
	if filter.PriceMin != nil || filter.PriceMax != nil {
		priceRange := make(map[string]float64)
		if filter.PriceMin != nil {
			priceRange["$gte"] = *filter.PriceMin
		}
		if filter.PriceMax != nil {
			priceRange["$lte"] = *filter.PriceMax
		}
		selector["price"] = priceRange
	}

	query := map[string]interface{}{
		"selector": selector,
	}

	queryBytes, _ := json.Marshal(query)
	resultsIterator, err := ctx.GetStub().GetQueryResult(string(queryBytes))
	if err != nil {
		return nil, fmt.Errorf("query failed: %v", err)
	}
	defer resultsIterator.Close()

	var products []*models.Product
	for resultsIterator.HasNext() {
		kv, _ := resultsIterator.Next()
		var p models.Product
		_ = json.Unmarshal(kv.Value, &p)
		products = append(products, &p)
	}

	return products, nil
}

func (s *TradingContract) GetAllProducts(ctx contractapi.TransactionContextInterface) ([]*models.Product, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange("PRODUCT_", "PRODUCT_~")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var assets []*models.Product
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var asset models.Product
		err = json.Unmarshal(queryResponse.Value, &asset)
		if err != nil {
			return nil, err
		}

		assets = append(assets, &asset)
	}

	return assets, nil
}
