package commands

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-gateway/pkg/client"
)

// CreateMerchant invokes CreateMerchant on the chaincode.
func CreateMerchant(contract *client.Contract, id, merchantType, pib string) error {
	fmt.Printf("→ Invoking CreateMerchant (id=%s, type=%s, pib=%s)\n", id, merchantType, pib)
	_, err := contract.SubmitTransaction("CreateMerchant", id, merchantType, pib)
	if err != nil {
		return fmt.Errorf("CreateMerchant failed: %w", err)
	}
	fmt.Println("✓ Merchant created successfully")
	return nil
}

// AddProducts invokes AddProducts – products is a JSON array of product objects.
func AddProducts(contract *client.Contract, merchantID, productsJSON string) error {
	fmt.Printf("→ Invoking AddProducts (merchantID=%s)\n", merchantID)
	// Validate JSON before sending
	var raw json.RawMessage
	if err := json.Unmarshal([]byte(productsJSON), &raw); err != nil {
		return fmt.Errorf("invalid products JSON: %w", err)
	}
	_, err := contract.SubmitTransaction("AddProducts", merchantID, productsJSON)
	if err != nil {
		return fmt.Errorf("AddProducts failed: %w", err)
	}
	fmt.Println("✓ Products added successfully")
	return nil
}
