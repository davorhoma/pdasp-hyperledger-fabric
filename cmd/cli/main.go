package main

import (
	"fmt"
	"os"

	"github.com/hyperledger/fabric-gateway/pkg/client"
)

func main() {
	if len(os.Args) < 3 {
		printUsage()
		return
	}

	org := os.Args[1]
	command := os.Args[2]
	args := os.Args[3:]

	// =========================================
	// OPCIJA 1: ENROLL (direktno sa CA)
	// =========================================
	if command == "enroll" {
		if len(args) < 2 {
			fmt.Println("Usage: go run . <org> enroll <username> <password>")
			return
		}
		err := EnrollUser(org, args[0], args[1])
		checkErr(err)
		return
	}

	// =========================================
	// OPCIJA 2: REGISTER & ENROLL (preko admin-a)
	// =========================================
	if command == "register-enroll" {
		if len(args) < 2 {
			fmt.Println("Usage: go run . <org> register-enroll <username> <password>")
			fmt.Println("This will use admin credentials to register the user first, then enroll")
			return
		}
		caConfig := getCAConfig(org)
		err := RegisterAndEnrollUser(org, caConfig.adminUser, caConfig.adminPass, args[0], args[1])
		checkErr(err)
		return
	}

	// =========================================
	// OPCIJA 3: LOGIN sa CA Wallet identitetom
	// =========================================
	if command == "login-wallet" {
		if len(args) < 1 {
			fmt.Println("Usage: go run . <org> login-wallet <username>")
			return
		}
		id, sign, err := LoadIdentityFromWallet(org, args[0])
		checkErr(err)
		fmt.Printf("✅ Successfully logged in as '%s'\n", args[0])
		fmt.Printf("   MSP ID: %s\n", id.MspID())
		_ = sign // Sign funkcija se koristi za potpisivanje transakcija
		return
	}

	// =========================================
	// OPCIJA 4: LOGIN sa STATIČKIM identitetom (bez CA)
	// =========================================
	if command == "login-static" {
		config := UseStaticIdentity(org)
		fmt.Println("✅ Logged in with STATIC identity:", config.mspID)
		return
	}

	// =========================================
	// WALLET MANAGEMENT
	// =========================================
	if command == "list-wallets" {
		identities, err := ListWalletIdentities(org)
		checkErr(err)
		fmt.Printf("Identities in wallet for %s:\n", org)
		if len(identities) == 0 {
			fmt.Println("  (none)")
		} else {
			for _, id := range identities {
				fmt.Printf("  - %s\n", id)
			}
		}
		return
	}

	if command == "delete-wallet" {
		if len(args) < 1 {
			fmt.Println("Usage: go run . <org> delete-wallet <username>")
			return
		}
		err := DeleteIdentityFromWallet(org, args[0])
		checkErr(err)
		return
	}

	if command == "wallet-info" {
		if len(args) < 1 {
			fmt.Println("Usage: go run . <org> wallet-info <username>")
			return
		}
		cert, err := GetIdentityInfo(org, args[0])
		checkErr(err)
		fmt.Printf("Identity: %s\n", args[0])
		fmt.Printf("  Subject: %s\n", cert.Subject)
		fmt.Printf("  Issuer: %s\n", cert.Issuer)
		fmt.Printf("  Valid from: %s\n", cert.NotBefore)
		fmt.Printf("  Valid until: %s\n", cert.NotAfter)
		return
	}

	config := UseStaticIdentity(org)

	conn := newGrpcConnection(&config)
	defer conn.Close()

	id := newIdentity(&config)
	sign := newSign(&config)

	gw, err := client.Connect(id, client.WithSign(sign), client.WithClientConnection(conn))
	checkErr(err)
	defer gw.Close()

	network := gw.GetNetwork("channel1")
	contract := network.GetContract("trading")

	switch command {
	case "InitLedger":
		InitLedger(contract)

	case "CreateUser":
		if len(args) < 5 {
			fmt.Println("Usage: CreateUser <id> <firstName> <lastName> <email> <balance>")
			return
		}
		CreateUser(contract, args[0], args[1], args[2], args[3], parseFloat(args[4]))

	case "CreateMerchant":
		if len(args) < 4 {
			fmt.Println("Usage: CreateMerchant <id> <type> <pib> <balance>")
			return
		}
		CreateMerchant(contract, args[0], args[1], args[2], parseFloat(args[3]))

	case "CreateProduct":
		if len(args) < 6 {
			fmt.Println("Usage: CreateProduct <merchantID> <productID> <name> <expiry> <price> <quantity>")
			return
		}
		CreateProduct(contract,
			args[0],
			args[1],
			args[2],
			args[3],
			parseFloat(args[4]),
			parseInt(args[5]),
		)

	case "Purchase":
		if len(args) < 5 {
			fmt.Println("Usage: Purchase <userID> <productID> <merchantID> <invoiceID> <quantity>")
			return
		}
		Purchase(contract,
			args[0],
			args[1],
			args[2],
			args[3],
			parseInt(args[4]),
		)

	case "RichQueryProducts":
		if len(args) < 1 {
			fmt.Println("Usage: RichQueryProducts <jsonFilter>")
			return
		}
		RichQueryProducts(contract, args[0])

	default:
		fmt.Println("Unknown command:", command)
		printUsage()
	}
}

func printUsage() {
	fmt.Println("==============================================")
	fmt.Println("Usage:")
	fmt.Println("go run . <org> <command> [args...]")
	fmt.Println("")
	fmt.Println("Organizations:")
	fmt.Println("  org1, org2, org3")
	fmt.Println("")
	fmt.Println("==========================================")
	fmt.Println("IDENTITY MANAGEMENT:")
	fmt.Println("==========================================")
	fmt.Println("  enroll <username> <password>")
	fmt.Println("      - Enroll new user with Fabric CA")
	fmt.Println("      - Example: go run . org1 enroll alice alicepass123")
	fmt.Println("")
	fmt.Println("  login-wallet <username>")
	fmt.Println("      - Login using enrolled identity from wallet")
	fmt.Println("      - Example: go run . org1 login-wallet alice")
	fmt.Println("")
	fmt.Println("  login-static")
	fmt.Println("      - Login using static identity (User1)")
	fmt.Println("      - Example: go run . org1 login-static")
	fmt.Println("")
	fmt.Println("==========================================")
	fmt.Println("CHAINCODE OPERATIONS (use static identity):")
	fmt.Println("==========================================")
	fmt.Println("  InitLedger")
	fmt.Println("  CreateUser <id> <firstName> <lastName> <email> <balance>")
	fmt.Println("  CreateMerchant <id> <type> <pib> <balance>")
	fmt.Println("  CreateProduct <merchantID> <productID> <name> <expiry> <price> <quantity>")
	fmt.Println("  Purchase <userID> <productID> <merchantID> <invoiceID> <quantity>")
	fmt.Println("  RichQueryProducts <jsonFilter>")
	fmt.Println("==============================================")
}

func parseFloat(s string) float64 {
	var f float64
	fmt.Sscan(s, &f)
	return f
}

func parseInt(s string) int {
	var i int
	fmt.Sscan(s, &i)
	return i
}
