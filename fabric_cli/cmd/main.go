package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"fabric-cli/internal/commands"
	gw "fabric-cli/internal/gateway"
)

var (
	configPath = flag.String("config", "config.yaml", "path to config.yaml")
	profileKey = flag.String("profile", "", "profile name to use (skips interactive login)")
	channel    = flag.String("channel", "", "override channel name from config")
	chaincode  = flag.String("chaincode", "", "override chaincode name from config")
)

func main() {
	flag.Parse()

	cfg, err := gw.LoadConfig(*configPath)
	if err != nil {
		fatalf("Cannot load config: %v\n", err)
	}

	if *channel != "" {
		cfg.Channel = *channel
	}
	if *chaincode != "" {
		cfg.Chaincode = *chaincode
	}

	selectedProfile := selectProfile(cfg)

	fmt.Printf("\n🔐 Logging in as [%s] ...\n", selectedProfile.Name)
	conn, err := gw.Connect(selectedProfile, cfg.Channel, cfg.Chaincode)
	if err != nil {
		fatalf("Connection failed: %v\n", err)
	}
	defer conn.Close()

	fmt.Printf("✅ Connected | Channel: %s | Chaincode: %s\n\n", cfg.Channel, cfg.Chaincode)

	scanner := bufio.NewScanner(os.Stdin)
	for {
		printMenu()
		choice := prompt(scanner, "Enter choice")

		switch strings.TrimSpace(choice) {
		case "1":
			handleInitLedger(scanner, conn)
		case "2":
			handleCreateMerchant(scanner, conn)
		case "3":
			handleAddProducts(scanner, conn)
		case "4":
			handleCreateUser(scanner, conn)
		case "5":
			handleDeposit(scanner, conn)
		case "6":
			handlePurchase(scanner, conn)
		case "7":
			handleGetAllProducts(conn)
		case "8":
			handleRichQuery(scanner, conn)
		case "9":
			handleSwitchProfile(cfg, conn)
			return
		case "10":
			handleRegisterAndEnroll(scanner)
		case "0":
			fmt.Println("Goodbye!")
			return
		default:
			fmt.Println("⚠️  Unknown option, try again.")
		}
	}
}

func printMenu() {
	fmt.Println()
	fmt.Println("════════════════════════════════════")
	fmt.Println("  Fabric Trading CLI")
	fmt.Println("════════════════════════════════════")
	fmt.Println("  INVOKE")
	fmt.Println("  1) Init Ledger")
	fmt.Println("  2) Create Merchant")
	fmt.Println("  3) Add Products to Merchant")
	fmt.Println("  4) Create User")
	fmt.Println("  5) Deposit Funds")
	fmt.Println("  6) Purchase Product")
	fmt.Println("  QUERY")
	fmt.Println("  7) Get All Products")
	fmt.Println("  8) Rich Query Products")
	fmt.Println("  OTHER")
	fmt.Println("  9) Switch Identity / Re-login")
	fmt.Println("  10) Enroll / Register user")
	fmt.Println("  0) Exit")
	fmt.Println("════════════════════════════════════")
}

func handleInitLedger(scanner *bufio.Scanner, conn *gw.Connection) {
	if err := commands.InitLedger(conn.Contract); err != nil {
		printErr(err)
	}
}

func handleCreateMerchant(scanner *bufio.Scanner, conn *gw.Connection) {
	id := prompt(scanner, "Merchant ID")
	mtype := prompt(scanner, "Type (e.g. supermarket, auto_parts)")
	pib := prompt(scanner, "PIB")
	if err := commands.CreateMerchant(conn.Contract, id, mtype, pib); err != nil {
		printErr(err)
	}
}

func handleAddProducts(scanner *bufio.Scanner, conn *gw.Connection) {
	merchantID := prompt(scanner, "Merchant ID")
	fmt.Println("Enter products as JSON array, e.g.:")
	fmt.Println(`  [{"id":"P5","name":"Cola","expiration":"2026-12-31T00:00:00Z","price":120,"quantity":30}]`)
	productsJSON := prompt(scanner, "Products JSON")
	if err := commands.AddProducts(conn.Contract, merchantID, productsJSON); err != nil {
		printErr(err)
	}
}

func handleCreateUser(scanner *bufio.Scanner, conn *gw.Connection) {
	id := prompt(scanner, "User ID")
	first := prompt(scanner, "First name")
	last := prompt(scanner, "Last name")
	email := prompt(scanner, "Email")
	if err := commands.CreateUser(conn.Contract, id, first, last, email); err != nil {
		printErr(err)
	}
}

func handleDeposit(scanner *bufio.Scanner, conn *gw.Connection) {
	entityType := promptChoice(scanner, "Entity type", "user", "merchant")
	id := prompt(scanner, "ID")
	amtStr := prompt(scanner, "Amount")
	amt, err := strconv.ParseFloat(strings.TrimSpace(amtStr), 64)
	if err != nil || amt <= 0 {
		fmt.Println("⚠️  Invalid amount")
		return
	}
	if err := commands.Deposit(conn.Contract, entityType, id, amt); err != nil {
		printErr(err)
	}
}

func handlePurchase(scanner *bufio.Scanner, conn *gw.Connection) {
	userID := prompt(scanner, "User ID")
	productID := prompt(scanner, "Product ID")
	invoiceID := prompt(scanner, "Invoice ID (e.g. INV001)")
	qtyStr := prompt(scanner, "Quantity")
	qty, err := strconv.Atoi(strings.TrimSpace(qtyStr))
	if err != nil || qty <= 0 {
		fmt.Println("⚠️  Invalid quantity")
		return
	}
	if err := commands.Purchase(conn.Contract, userID, productID, invoiceID, qty); err != nil {
		printErr(err)
	}
}

func handleGetAllProducts(conn *gw.Connection) {
	result, err := commands.GetAllProducts(conn.Contract)
	if err != nil {
		printErr(err)
		return
	}
	printResult(result)
}

func handleRichQuery(scanner *bufio.Scanner, conn *gw.Connection) {
	fmt.Println("Build your filter (leave blank to skip a field):")

	filter := map[string]interface{}{}

	if v := prompt(scanner, "  Product ID"); v != "" {
		filter["id"] = v
	}
	if v := prompt(scanner, "  Name (partial match supported)"); v != "" {
		filter["name"] = v
	}
	if v := prompt(scanner, "  Merchant type (e.g. supermarket)"); v != "" {
		filter["merchantType"] = v
	}
	if v := prompt(scanner, "  Min price (leave blank to skip)"); v != "" {
		if f, err := strconv.ParseFloat(strings.TrimSpace(v), 64); err == nil {
			filter["priceMin"] = f
		}
	}
	if v := prompt(scanner, "  Max price (leave blank to skip)"); v != "" {
		if f, err := strconv.ParseFloat(strings.TrimSpace(v), 64); err == nil {
			filter["priceMax"] = f
		}
	}

	filterBytes, _ := json.Marshal(filter)
	result, err := commands.RichQueryProducts(conn.Contract, string(filterBytes))
	if err != nil {
		printErr(err)
		return
	}
	printResult(result)
}

func handleSwitchProfile(cfg *gw.Config, oldConn *gw.Connection) {
	oldConn.Close()
	fmt.Println("Switching identity — please restart the program or select a new profile below.")
	p := selectProfile(cfg)
	conn, err := gw.Connect(p, cfg.Channel, cfg.Chaincode)
	if err != nil {
		fatalf("Reconnection failed: %v\n", err)
	}
	defer conn.Close()
	fmt.Printf("✅ Now logged in as [%s]\n", p.Name)

	scanner := bufio.NewScanner(os.Stdin)
	for {
		printMenu()
		choice := prompt(scanner, "Enter choice")
		if choice == "0" {
			return
		}
		dispatchMenu(choice, scanner, conn, cfg)
	}
}

func dispatchMenu(choice string, scanner *bufio.Scanner, conn *gw.Connection, cfg *gw.Config) {
	switch strings.TrimSpace(choice) {
	case "1":
		handleInitLedger(scanner, conn)
	case "2":
		handleCreateMerchant(scanner, conn)
	case "3":
		handleAddProducts(scanner, conn)
	case "4":
		handleCreateUser(scanner, conn)
	case "5":
		handleDeposit(scanner, conn)
	case "6":
		handlePurchase(scanner, conn)
	case "7":
		handleGetAllProducts(conn)
	case "8":
		handleRichQuery(scanner, conn)
	case "9":
		handleSwitchProfile(cfg, conn)
	case "10":
		handleRegisterAndEnroll(scanner)
	}
}

func handleRegisterAndEnroll(scanner *bufio.Scanner) {
	fmt.Println("── Enroll / Register new user ──")
	username := prompt(scanner, "Username")
	password := prompt(scanner, "Password for user")
	orgMSP := prompt(scanner, "Org MSP (e.g. Org1MSP)")

	profile, err := gw.RegisterAndEnrollUser(username, password, orgMSP)
	if err != nil {
		printErr(err)
		return
	}
	fmt.Printf("\nUser %s enrolled successfully!\n", username)

	err = gw.AddProfileToConfig("config.yaml", profile)
	if err != nil {
		printErr(err)
		return
	}
}

func selectProfile(cfg *gw.Config) gw.Profile {
	if *profileKey != "" {
		p, ok := cfg.Profiles[*profileKey]
		if !ok {
			fatalf("Profile %q not found in config\n", *profileKey)
		}
		return p
	}

	keys := make([]string, 0, len(cfg.Profiles))
	for k := range cfg.Profiles {
		keys = append(keys, k)
	}

	fmt.Println("════════════════════════════════════")
	fmt.Println("  Available identities / profiles")
	fmt.Println("════════════════════════════════════")
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	for i, k := range keys {
		p := cfg.Profiles[k]
		fmt.Fprintf(tw, "  %d)\t%s\t(%s)\n", i+1, p.Name, p.OrgMSP)
	}
	tw.Flush()
	fmt.Println("════════════════════════════════════")

	scanner := bufio.NewScanner(os.Stdin)
	for {
		choice := prompt(scanner, "Select profile number")
		n, err := strconv.Atoi(strings.TrimSpace(choice))
		if err != nil || n < 1 || n > len(keys) {
			fmt.Printf("⚠️  Enter a number between 1 and %d\n", len(keys))
			continue
		}
		return cfg.Profiles[keys[n-1]]
	}
}

func prompt(scanner *bufio.Scanner, label string) string {
	fmt.Printf("  %s: ", label)
	scanner.Scan()
	return strings.TrimSpace(scanner.Text())
}

func promptChoice(scanner *bufio.Scanner, label string, options ...string) string {
	for {
		v := prompt(scanner, fmt.Sprintf("%s (%s)", label, strings.Join(options, "/")))
		for _, o := range options {
			if v == o {
				return v
			}
		}
		fmt.Printf("⚠️  Choose one of: %s\n", strings.Join(options, ", "))
	}
}

func printErr(err error) {
	fmt.Printf("❌ Error: %v\n", err)
}

func printResult(data []byte) {
	fmt.Println("─── Result ────────────────────────────")
	fmt.Println(string(data))
	fmt.Println("───────────────────────────────────────")
}

func fatalf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "FATAL: "+format, args...)
	os.Exit(1)
}
