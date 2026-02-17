package commands

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric-gateway/pkg/client"
)

// Purchase invokes Purchase on the chaincode.
func Purchase(contract *client.Contract, userID, productID, invoiceID string, quantity int) error {
	fmt.Printf("→ Invoking Purchase (user=%s, product=%s, qty=%d, invoiceID=%s)\n",
		userID, productID, quantity, invoiceID)
	qty := strconv.Itoa(quantity)
	_, err := contract.SubmitTransaction("Purchase", userID, productID, invoiceID, qty)
	if err != nil {
		return fmt.Errorf("Purchase failed: %w", err)
	}
	fmt.Println("✓ Purchase completed successfully")
	return nil
}

// GetAllProducts queries all products via range query.
func GetAllProducts(contract *client.Contract) ([]byte, error) {
	fmt.Println("→ Querying GetAllProducts")
	result, err := contract.EvaluateTransaction("GetAllProducts")
	if err != nil {
		return nil, fmt.Errorf("GetAllProducts failed: %w", err)
	}
	return prettyJSON(result), nil
}

// RichQueryProducts sends a CouchDB selector-based rich query.
// filterJSON example: {"name":"Mleko","priceMin":10,"priceMax":100}
func RichQueryProducts(contract *client.Contract, filterJSON string) ([]byte, error) {
	fmt.Printf("→ Querying RichQueryProducts (filter=%s)\n", filterJSON)
	// Validate JSON
	var raw json.RawMessage
	if err := json.Unmarshal([]byte(filterJSON), &raw); err != nil {
		return nil, fmt.Errorf("invalid filter JSON: %w", err)
	}
	result, err := contract.EvaluateTransaction("RichQueryProducts", filterJSON)
	if err != nil {
		return nil, fmt.Errorf("RichQueryProducts failed: %w", err)
	}
	return prettyJSON(result), nil
}

// InitLedger invokes InitLedger.
func InitLedger(contract *client.Contract) error {
	fmt.Println("→ Invoking InitLedger")
	_, err := contract.SubmitTransaction("InitLedger")
	if err != nil {
		return fmt.Errorf("InitLedger failed: %w", err)
	}
	fmt.Println("✓ Ledger initialized")
	return nil
}

// prettyJSON re-formats raw JSON bytes with indentation.
// Falls back to raw bytes if parsing fails.
func prettyJSON(raw []byte) []byte {
	var v interface{}
	if err := json.Unmarshal(raw, &v); err != nil {
		return raw
	}
	pretty, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return raw
	}
	return pretty
}
