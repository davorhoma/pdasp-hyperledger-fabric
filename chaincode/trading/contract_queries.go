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

func (s *TradingContract) GetMerchantByID(ctx contractapi.TransactionContextInterface, merchantID string) (*models.Merchant, error) {
	merchantKey := "MERCHANT_" + merchantID
	merchantBytes, err := ctx.GetStub().GetState(merchantKey)
	if err != nil {
		return nil, fmt.Errorf("failed to read merchant: %v", err)
	}
	if merchantBytes == nil {
		return nil, fmt.Errorf("merchant with ID %s does not exist", merchantID)
	}

	var merchant models.Merchant
	err = json.Unmarshal(merchantBytes, &merchant)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal merchant data: %v", err)
	}

	return &merchant, nil
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

// -------------------------------------------------------------------------------
// RICH QUERY 1 – Proizvodi kojima uskoro ističe rok trajanja
//
// CouchDB prednost: selector direktno filtrira po polju "expiration" koje je
// duboko unutar JSON dokumenta, uz $lte poređenje datuma u ISO-8601 formatu.
// Vraća samo one proizvode čiji rok ističe pre navedenog datuma.
//
// LevelDB ekvivalent: ne postoji direktan ekvivalent. Moralo bi se:
//   1. GetStateByRange("PRODUCT_", "PRODUCT_~") – dohvati SVE proizvode
//   2. Deserijalizovati svaki dokument u Go strukturu
//   3. Ručno u Go kodu porediti polje Expiration sa zadatim datumom
//   To znači da se ceo skup podataka učitava u memoriju chaincode-a,
//   bez ikakve selekcije na nivou baze podataka.
// -------------------------------------------------------------------------------

// GetProductsExpiringSoon vraća sve proizvode čiji rok trajanja ističe
// pre datuma expiresBeforeDate (format: "2026-12-31T23:59:59Z").
func (t *TradingContract) GetProductsExpiringSoon(
	ctx contractapi.TransactionContextInterface,
	expiresBeforeDate string,
) ([]*models.Product, error) {

	if expiresBeforeDate == "" {
		return nil, fmt.Errorf("expiresBeforeDate ne sme biti prazan")
	}

	query := map[string]interface{}{
		"selector": map[string]interface{}{
			"docType": "product",
			"expiration": map[string]interface{}{
				"$lte": expiresBeforeDate,
			},
			"quantity": map[string]interface{}{
				"$gt": 0,
			},
		},
		"sort": []map[string]string{
			{"expiration": "asc"},
		},
	}

	queryBytes, _ := json.Marshal(query)
	resultsIterator, err := ctx.GetStub().GetQueryResult(string(queryBytes))
	if err != nil {
		return nil, fmt.Errorf("GetProductsExpiringSoon query failed: %v", err)
	}
	defer resultsIterator.Close()

	var products []*models.Product
	for resultsIterator.HasNext() {
		kv, _ := resultsIterator.Next()
		var p models.Product
		if err := json.Unmarshal(kv.Value, &p); err == nil {
			products = append(products, &p)
		}
	}

	return products, nil
}

// -------------------------------------------------------------------------------
// RICH QUERY 2 – Korisnici sa visokim stanjem na računu
//
// CouchDB prednost: $gte operator na numeričkom polju "balance" uz
// istovremeni filter po "docType". Kombinovanje filtera po tipu dokumenta
// i numeričkom opsegu u jednom prolazu kroz bazu.
//
// LevelDB ekvivalent:
//   1. GetStateByRange("USER_", "USER_~") – dohvati sve korisnike
//   2. Iterirati i ručno filtrirati u Go-u: if user.Balance >= minBalance
//   3. Nema sortiranja na nivou baze – moralo bi se sortirati u memoriji
//      sort.Slice(users, func(i,j int) bool { return users[i].Balance > users[j].Balance })
// -------------------------------------------------------------------------------

// GetUsersWithMinBalance vraća sve korisnike čije je stanje >= minBalance,
// sortirane po stanju opadajuće (najbogatiji prvi).
func (t *TradingContract) GetUsersWithMinBalance(
	ctx contractapi.TransactionContextInterface,
	minBalance float64,
) ([]*models.User, error) {

	if minBalance < 0 {
		return nil, fmt.Errorf("minBalance mora biti >= 0")
	}

	query := map[string]interface{}{
		"selector": map[string]interface{}{
			"docType": "user",
			"balance": map[string]interface{}{
				"$gte": minBalance,
			},
		},
		"sort": []map[string]string{
			{"balance": "desc"},
		},
	}

	queryBytes, _ := json.Marshal(query)
	resultsIterator, err := ctx.GetStub().GetQueryResult(string(queryBytes))
	if err != nil {
		return nil, fmt.Errorf("GetUsersWithMinBalance query failed: %v", err)
	}
	defer resultsIterator.Close()

	var users []*models.User
	for resultsIterator.HasNext() {
		kv, _ := resultsIterator.Next()
		var u models.User
		if err := json.Unmarshal(kv.Value, &u); err == nil {
			users = append(users, &u)
		}
	}

	return users, nil
}

// -------------------------------------------------------------------------------
// RICH QUERY 3 – Fakture u zadatom vremenskom periodu za određenog korisnika
//
// CouchDB prednost: kombinovanje filtera po polju "userId" (jednakost)
// I po polju "date" (opseg $gte/$lte) u jednom upitu. Ovo je tipičan
// "compound query" koji u relacionim bazama zahteva indeks na dva polja.
// CouchDB to rešava jednim Mango selektorom bez potrebe za JOIN-om.
//
// LevelDB ekvivalent:
//   1. GetStateByRange("INVOICE_", "INVOICE_~") – dohvati SVE fakture
//   2. Za svaku: unmarshaling + provera userID == zadati AND date u opsegu
//   3. O(n) gde je n ukupan broj faktura u sistemu, bez obzira na filtar
//   4. Posebno skupo ako ima hiljada faktura, a traži se samo jedna od 10.
// -------------------------------------------------------------------------------

// GetInvoicesByUserAndDateRange vraća fakture korisnika userID
// u vremenskom periodu [fromDate, toDate] (ISO-8601 format).
func (t *TradingContract) GetInvoicesByUserAndDateRange(
	ctx contractapi.TransactionContextInterface,
	userID string,
	fromDate string,
	toDate string,
) ([]*models.Invoice, error) {

	if userID == "" || fromDate == "" || toDate == "" {
		return nil, fmt.Errorf("userID, fromDate i toDate su obavezni")
	}

	query := map[string]interface{}{
		"selector": map[string]interface{}{
			"docType": "invoice",
			"userId":  userID,
			"date": map[string]interface{}{
				"$gte": fromDate,
				"$lte": toDate,
			},
		},
		"sort": []map[string]string{
			{"date": "desc"},
		},
	}

	queryBytes, _ := json.Marshal(query)
	resultsIterator, err := ctx.GetStub().GetQueryResult(string(queryBytes))
	if err != nil {
		return nil, fmt.Errorf("GetInvoicesByUserAndDateRange query failed: %v", err)
	}
	defer resultsIterator.Close()

	var invoices []*models.Invoice
	for resultsIterator.HasNext() {
		kv, _ := resultsIterator.Next()
		var inv models.Invoice
		if err := json.Unmarshal(kv.Value, &inv); err == nil {
			invoices = append(invoices, &inv)
		}
	}

	return invoices, nil
}

// -------------------------------------------------------------------------------
// RICH QUERY 4 – Merchants sa niskim stanjem zaliha (low stock alert)
//
// CouchDB prednost: $and operator koji kombinuje filter po docType, po
// merchantType i po quantity (opseg). Ovo je višestruki compound filter
// po RAZLIČITIM tipovima dokumenata koji se u LevelDB-u ne može raditi
// bez potpunog skeniranja celog world state-a.
// Posebno korisno za business logiku: "upozori me kad nešto nestaje".
//
// LevelDB ekvivalent:
//   1. GetStateByRange("PRODUCT_", "PRODUCT_~") – skeniranje svih proizvoda
//   2. Ručna filtracija: if p.MerchantType == zadati AND p.Quantity <= maxQty
//   3. Ista O(n) složenost bez obzira na broj pogodaka
// -------------------------------------------------------------------------------

// GetLowStockProducts vraća sve proizvode određenog tipa trgovca
// čija je količina na stanju <= maxQuantity.
// Korisno za upozorenja o niskim zalihama.
func (t *TradingContract) GetLowStockProducts(
	ctx contractapi.TransactionContextInterface,
	merchantType string,
	maxQuantity int,
) ([]*models.Product, error) {

	if merchantType == "" {
		return nil, fmt.Errorf("merchantType je obavezan")
	}
	if maxQuantity < 0 {
		return nil, fmt.Errorf("maxQuantity mora biti >= 0")
	}

	query := map[string]interface{}{
		"selector": map[string]interface{}{
			"docType":      "product",
			"merchantType": merchantType,
			"quantity": map[string]interface{}{
				"$lte": maxQuantity,
				"$gt":  0,
			},
		},
		"sort": []map[string]string{
			{"quantity": "asc"},
		},
	}

	queryBytes, _ := json.Marshal(query)
	resultsIterator, err := ctx.GetStub().GetQueryResult(string(queryBytes))
	if err != nil {
		return nil, fmt.Errorf("GetLowStockProducts query failed: %v", err)
	}
	defer resultsIterator.Close()

	var products []*models.Product
	for resultsIterator.HasNext() {
		kv, _ := resultsIterator.Next()
		var p models.Product
		if err := json.Unmarshal(kv.Value, &p); err == nil {
			products = append(products, &p)
		}
	}

	return products, nil
}

// -------------------------------------------------------------------------------
// RICH QUERY 5 – Sve fakture za određenog trgovca iznad zadatog iznosa
//
// CouchDB prednost: kombinovanje filtera po "merchantId" (jednakost) i
// "totalPrice" (numerički opseg $gte), uz sortiranje po totalPrice desc.
// Ovo je tipičan "top transactions" upit koji bi u LevelDB-u zahtevao:
//   - Skeniranje SVIH faktura (bez obzira na merchantId)
//   - Ručnu filtraciju i sortiranje u memoriji chaincode-a
// Posebno zanimljiv jer je totalPrice izvedena vrednost (price * qty)
// koja se čuva u dokumentu – CouchDB može da indeksira i tu vrednost.
//
// LevelDB ekvivalent:
//   1. GetStateByRange("INVOICE_", "INVOICE_~")
//   2. Ručno filtrirati: inv.MerchantID == zadati AND inv.TotalPrice >= min
//   3. sort.Slice po TotalPrice desc
//   4. Svaka nova faktura ne menja efikasnost – uvek O(n) skeniranje
// -------------------------------------------------------------------------------

// GetMerchantHighValueInvoices vraća sve fakture trgovca merchantID
// čiji je ukupan iznos >= minTotalPrice, sortirane po iznosu opadajuće.
func (t *TradingContract) GetMerchantHighValueInvoices(
	ctx contractapi.TransactionContextInterface,
	merchantID string,
	minTotalPrice float64,
) ([]*models.Invoice, error) {

	if merchantID == "" {
		return nil, fmt.Errorf("merchantID je obavezan")
	}
	if minTotalPrice < 0 {
		return nil, fmt.Errorf("minTotalPrice mora biti >= 0")
	}

	query := map[string]interface{}{
		"selector": map[string]interface{}{
			"docType":    "invoice",
			"merchantId": merchantID,
			"totalPrice": map[string]interface{}{
				"$gte": minTotalPrice,
			},
		},
		"sort": []map[string]string{
			{"totalPrice": "desc"},
		},
	}

	queryBytes, _ := json.Marshal(query)
	resultsIterator, err := ctx.GetStub().GetQueryResult(string(queryBytes))
	if err != nil {
		return nil, fmt.Errorf("GetMerchantHighValueInvoices query failed: %v", err)
	}
	defer resultsIterator.Close()

	var invoices []*models.Invoice
	for resultsIterator.HasNext() {
		kv, _ := resultsIterator.Next()
		var inv models.Invoice
		if err := json.Unmarshal(kv.Value, &inv); err == nil {
			invoices = append(invoices, &inv)
		}
	}

	return invoices, nil
}
