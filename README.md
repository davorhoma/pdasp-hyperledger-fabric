# PDASP Projekat – Hyperledger Fabric

Ovaj projekat koristi **Hyperledger Fabric** za implementaciju blockchain mreže i **chaincode-a** u Go-u. Projekat uključuje i konzolnu aplikaciju za interakciju sa mrežom.

---

## Pokretanje Fabric mreže

1. Pređite u direktorijum test mreže:

```bash
cd ./fabric-samples/test-network
```

2. Pokrenite mrežu i kreirajte prvi kanal sa CouchDB:
```bash
./network.sh up createChannel -ca -s couchdb
```

3. Kreirajte dodatni kanal channel2:
```bash
./network.sh createChannel -c channel2
```

4. Deploy-ujte chaincode trading:
```bash
./network.sh deployCC \
  -ccn trading \
  -ccp ../../chaincode \
  -ccl go \
  -ccep "OutOf(2,'Org1MSP.peer','Org2MSP.peer','Org3MSP.peer')" \
  -cci InitLedger
```

# Pokretanje testova za chaincode

1. Pređite u direktorijum sa skriptama:
```bash
cd ./scripts
```

2. Testove možete pokrenuti pojedinačno. Struktura skripti:
scripts/
├── create_merchants.sh
├── init_ledger.sh
├── test_create_merchant.sh
├── test_create_products.sh
├── test_create_users.sh
├── test_deposit.sh
├── test_errors.sh
├── test_get_all_products.sh
├── test_get_user.sh
├── test_purchase.sh
└── test_rich_queries.sh


3. Svaku skriptu pokrećete posebno, npr.:
```bash
./test_create_merchant.sh
```

### Pokretanje testova za konzolnu aplikaciju

1. Pređite u direktorijum konzolne aplikacije:
```bash
cd ./fabric_cli
```

2. Pokrenite sve testove:
```bash
./test_all.sh
```

### Pokretanje konzolne aplikacije

1. Pređite u direktorijum konzolne aplikacije:
```bash
cd ./fabric_cli
```

2. Pokrenite aplikaciju:
```bash
go run ./cmd/main.go
```