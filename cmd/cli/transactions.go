package main

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-gateway/pkg/client"
)

func InitLedger(contract *client.Contract) {
	fmt.Println("--> Submit Transaction: InitLedger")
	_, err := contract.SubmitTransaction("InitLedger")
	checkErr(err)
	fmt.Println("*** Ledger initialized ✅")
}

func CreateUser(contract *client.Contract, id, firstName, lastName, email string, balance float64) {
	fmt.Printf("--> Submit Transaction: CreateUser %s\n", id)
	_, err := contract.SubmitTransaction("CreateUser", id, firstName, lastName, email, fmt.Sprintf("%f", balance))
	checkErr(err)
	fmt.Println("*** User created ✅")
}

func CreateMerchant(contract *client.Contract, id, merchantType, pib string, balance float64) {
	fmt.Printf("--> Submit Transaction: CreateMerchant %s\n", id)
	_, err := contract.SubmitTransaction("CreateMerchant", id, merchantType, pib, fmt.Sprintf("%f", balance))
	checkErr(err)
	fmt.Println("*** Merchant created ✅")
}

func CreateProduct(contract *client.Contract, merchantID, productID, name, expiry string, price float64, quantity int) {
	fmt.Printf("--> Submit Transaction: CreateProduct %s for Merchant %s\n", productID, merchantID)
	_, err := contract.SubmitTransaction("CreateProduct", merchantID, productID, name, expiry, fmt.Sprintf("%f", price), fmt.Sprintf("%d", quantity))
	checkErr(err)
	fmt.Println("*** Product created ✅")
}

func Purchase(contract *client.Contract, userID, productID, merchantID, invoiceID string, quantity int) {
	fmt.Printf("--> Submit Transaction: Purchase %s buys %d of %s from %s\n", userID, quantity, productID, merchantID)
	_, err := contract.SubmitTransaction("Purchase", userID, productID, merchantID, invoiceID, fmt.Sprintf("%d", quantity))
	if err != nil {
		fmt.Printf("Expected error (if any): %v\n", err)
		return
	}
	fmt.Println("*** Purchase executed ✅")
}

func RichQueryProducts(contract *client.Contract, filter string) {
	fmt.Printf("--> Evaluate Transaction: RichQueryProducts %s\n", filter)
	result, err := contract.EvaluateTransaction("RichQueryProducts", filter)
	checkErr(err)

	var out bytes.Buffer
	err = json.Indent(&out, result, "", "  ")
	checkErr(err)

	fmt.Println("*** Query Result:")
	fmt.Println(out.String())
}

func checkErr(err error) {
	if err != nil {
		panic(fmt.Errorf("Transaction failed: %w", err))
	}
}
