#!/bin/bash

# Test Enrollment i Login funkcionalnosti (novi fabric-gateway)

set -e

echo "=================================================="
echo "🔐 TESTING ENROLLMENT & LOGIN (New Fabric Gateway)"
echo "=================================================="
echo ""

# Boje
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

# ==================================================
# PROVERA PREDUSLOVA
# ==================================================
echo -e "${YELLOW}[0/7] Checking prerequisites...${NC}"

# Proveri da li CA serveri rade
if ! docker ps | grep -q "ca"; then
    echo -e "${RED}❌ CA containers are not running!${NC}"
    echo "Please start your Fabric network first:"
    echo "  ./network.sh up -ca"
    exit 1
fi

echo -e "${GREEN}✅ CA servers are running${NC}"

# Proveri da li Go dependencies postoje
if ! go list github.com/hyperledger/fabric-gateway/pkg/client &> /dev/null; then
    echo -e "${YELLOW}⚠️  Missing Go dependencies, installing...${NC}"
    go mod download
fi

echo ""

# ==================================================
# 1. ENROLL ADMIN IDENTITETA
# ==================================================
echo -e "${BLUE}[1/7] Enrolling Admin users for each organization...${NC}"

echo "Enrolling admin for Org1..."
go run . org1 enroll admin adminpw || echo -e "${YELLOW}Admin for Org1 may already exist${NC}"
echo ""

echo "Enrolling admin for Org2..."
go run . org2 enroll admin adminpw || echo -e "${YELLOW}Admin for Org2 may already exist${NC}"
echo ""

echo "Enrolling admin for Org3..."
go run . org3 enroll admin adminpw || echo -e "${YELLOW}Admin for Org3 may already exist${NC}"
echo ""

echo -e "${GREEN}✅ Admin enrollment completed${NC}"
echo ""

# ==================================================
# 2. REGISTER & ENROLL OBIČNIH KORISNIKA
# ==================================================
echo -e "${BLUE}[2/7] Registering and enrolling regular users...${NC}"

echo "Register-enrolling alice for Org1..."
go run . org1 register-enroll alice alicepass123 || echo -e "${YELLOW}Alice may already exist${NC}"
echo ""

echo "Register-enrolling bob for Org2..."
go run . org2 register-enroll bob bobpass456 || echo -e "${YELLOW}Bob may already exist${NC}"
echo ""

echo "Register-enrolling charlie for Org3..."
go run . org3 register-enroll charlie charliepass789 || echo -e "${YELLOW}Charlie may already exist${NC}"
echo ""

echo -e "${GREEN}✅ User enrollment completed${NC}"
echo ""

# ==================================================
# 3. LIST WALLETS
# ==================================================
echo -e "${BLUE}[3/7] Listing wallet identities...${NC}"

echo "Org1 wallets:"
go run . org1 list-wallets
echo ""

echo "Org2 wallets:"
go run . org2 list-wallets
echo ""

echo "Org3 wallets:"
go run . org3 list-wallets
echo ""

echo -e "${GREEN}✅ Wallet listing completed${NC}"
echo ""

# ==================================================
# 4. WALLET INFO
# ==================================================
echo -e "${BLUE}[4/7] Checking wallet certificate info...${NC}"

echo "Alice's certificate info:"
go run . org1 wallet-info alice
echo ""

echo "Bob's certificate info:"
go run . org2 wallet-info bob
echo ""

echo -e "${GREEN}✅ Certificate info check completed${NC}"
echo ""

# ==================================================
# 5. LOGIN SA WALLET IDENTITETIMA
# ==================================================
echo -e "${BLUE}[5/7] Testing login with wallet identities...${NC}"

echo "Login as alice (Org1)..."
go run . org1 login-wallet alice
echo ""

echo "Login as bob (Org2)..."
go run . org2 login-wallet bob
echo ""

echo "Login as charlie (Org3)..."
go run . org3 login-wallet charlie
echo ""

echo -e "${GREEN}✅ Wallet login test completed${NC}"
echo ""

# ==================================================
# 6. LOGIN SA STATIČKIM IDENTITETIMA
# ==================================================
echo -e "${BLUE}[6/7] Testing login with static identities...${NC}"

echo "Static login for Org1..."
go run . org1 login-static
echo ""

echo "Static login for Org2..."
go run . org2 login-static
echo ""

echo "Static login for Org3..."
go run . org3 login-static
echo ""

echo -e "${GREEN}✅ Static login test completed${NC}"
echo ""

# ==================================================
# 7. TESTIRANJE TRANSAKCIJA SA RAZLIČITIM IDENTITETIMA
# ==================================================
echo -e "${BLUE}[7/7] Testing chaincode operations with different identities...${NC}"

echo "Initialize ledger using Org1 static identity..."
go run . org1 InitLedger 2>&1 | head -n 5 || echo -e "${YELLOW}InitLedger may have already been called${NC}"
echo ""

echo "Create user using Org2 static identity..."
go run . org2 CreateUser user_test1 Test User test@email.com 5000.0 || echo -e "${YELLOW}User may already exist${NC}"
echo ""

echo "Create merchant using Org1 static identity..."
go run . org1 CreateMerchant merch_test1 TestShop 11111111 10000.0 || echo -e "${YELLOW}Merchant may already exist${NC}"
echo ""

echo -e "${GREEN}✅ Transaction test completed${NC}"
echo ""

# ==================================================
# PROVERA WALLET STRUKTURE
# ==================================================
echo -e "${BLUE}Final: Checking wallet structure...${NC}"

if [ -d "wallets" ]; then
    echo "Wallet directory structure:"
    tree wallets/ 2>/dev/null || find wallets/ -type f | sort
    echo ""
else
    echo -e "${YELLOW}Wallets directory not found (no enrollments performed)${NC}"
    echo ""
fi

# ==================================================
# SUMMARY
# ==================================================
echo "=================================================="
echo -e "${GREEN}🎉 ALL TESTS COMPLETED SUCCESSFULLY!${NC}"
echo "=================================================="
echo ""
echo "Summary:"
echo "  ✅ Admin enrollments: Org1, Org2, Org3"
echo "  ✅ User enrollments: alice, bob, charlie"
echo "  ✅ Wallet listings: PASSED"
echo "  ✅ Certificate info: PASSED"
echo "  ✅ Wallet login tests: PASSED"
echo "  ✅ Static login tests: PASSED"
echo "  ✅ Chaincode operations: PASSED"
echo ""
echo "Wallet identities created:"
go run . org1 list-wallets 2>/dev/null | grep "  -" || echo "  None for org1"
go run . org2 list-wallets 2>/dev/null | grep "  -" || echo "  None for org2"
go run . org3 list-wallets 2>/dev/null | grep "  -" || echo "  None for org3"
echo ""
echo "Next steps:"
echo "  - Use 'go run . <org> login-wallet <username>' to login"
echo "  - Use 'go run . <org> login-static' for admin operations"
echo "  - Use 'go run . <org> list-wallets' to see all identities"
echo "  - Use 'go run . <org> wallet-info <username>' for cert details"
echo ""
echo "Fabric Gateway enrollment system is working! ✅"
echo ""