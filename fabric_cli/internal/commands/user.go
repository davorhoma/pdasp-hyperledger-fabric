package commands

import (
	"fmt"

	"github.com/hyperledger/fabric-gateway/pkg/client"
)

// CreateUser invokes CreateUser on the chaincode.
func CreateUser(contract *client.Contract, id, firstName, lastName, email string) error {
	fmt.Printf("→ Invoking CreateUser (id=%s, name=%s %s, email=%s)\n", id, firstName, lastName, email)
	_, err := contract.SubmitTransaction("CreateUser", id, firstName, lastName, email)
	if err != nil {
		return fmt.Errorf("CreateUser failed: %w", err)
	}
	fmt.Println("✓ User created successfully")
	return nil
}

// Deposit invokes Deposit on the chaincode.
// entityType: "user" | "merchant"
func Deposit(contract *client.Contract, entityType, id string, amount float64) error {
	fmt.Printf("→ Invoking Deposit (type=%s, id=%s, amount=%.2f)\n", entityType, id, amount)
	amountStr := fmt.Sprintf("%.2f", amount)
	_, err := contract.SubmitTransaction("Deposit", entityType, id, amountStr)
	if err != nil {
		return fmt.Errorf("Deposit failed: %w", err)
	}
	fmt.Printf("✓ Deposited %.2f to %s %s\n", amount, entityType, id)
	return nil
}
