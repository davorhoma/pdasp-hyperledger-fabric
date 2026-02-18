#!/usr/bin/env bash
# ──────────────────────────────────────────────────────────────────────────────
# test_all.sh  –  End-to-end test script for the Fabric Trading chaincode
#
# Usage:
#   ./test_all.sh                 # uses default config.yaml + org1admin
#   PROFILE=org2admin ./test_all.sh
#
# Prerequisites:
#   - Fabric network is running
#   - CLI binary is built: go build -o fabric-cli ./cmd/
#   - Config paths in config.yaml point to real crypto-material
# ──────────────────────────────────────────────────────────────────────────────

set -euo pipefail

CLI="./fabric-cli"
CONFIG="${CONFIG:-config.yaml}"
PROFILE="${PROFILE:-org1admin}"
PROFILE2="${PROFILE2:-org2admin}"    # second identity for the login-switch test

RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; NC='\033[0m'

pass() { echo -e "${GREEN}✓ PASS${NC}: $1"; }
fail() { echo -e "${RED}✗ FAIL${NC}: $1"; exit 1; }
section() { echo -e "\n${YELLOW}══ $1 ══${NC}"; }

# ─── Helper: run a single CLI command non-interactively ──────────────────────
# Usage: cli_invoke <profile> <args...>
# We pipe a newline so the scanner doesn't block, then send "0" to quit.
cli_invoke() {
    local profile="$1"; shift
    echo -e "0" | "$CLI" -config "$CONFIG" -profile "$profile" "$@" 2>&1
}

# ─── Helper: run an interactive sequence of menu choices ─────────────────────
# Usage: cli_menu <profile> <input_string>
# input_string: newline-separated menu choices + answers, ending with "0"
cli_menu() {
    local profile="$1"; shift
    echo -e "$1" | "$CLI" -config "$CONFIG" -profile "$profile"
}

# ─────────────────────────────────────────────────────────────────────────────
section "0. Build CLI"
# ─────────────────────────────────────────────────────────────────────────────
go build -o "$CLI" ./cmd/ || fail "Build failed"
pass "CLI binary built"

# ─────────────────────────────────────────────────────────────────────────────
section "1. Init Ledger  [Org1Admin]"
# ─────────────────────────────────────────────────────────────────────────────
output=$(cli_menu "$PROFILE" "1\n0")
echo "$output"
echo "$output" | grep -q "Ledger initialized" || fail "InitLedger"
pass "InitLedger"

# ─────────────────────────────────────────────────────────────────────────────
section "2. Create Merchant  [Org1Admin]"
# ─────────────────────────────────────────────────────────────────────────────
output=$(cli_menu "$PROFILE" "2\nMERCHANT3\npharmacy\n111222333\n0")
echo "$output"
echo "$output" | grep -q "Merchant created successfully" || fail "CreateMerchant"
pass "CreateMerchant"

# ─────────────────────────────────────────────────────────────────────────────
section "3. Add Products to Merchant  [Org1Admin]"
# ─────────────────────────────────────────────────────────────────────────────
PRODUCTS_JSON='[{"id":"PROD5","name":"Aspirin","expiration":"2027-06-01T00:00:00Z","price":250,"quantity":100},{"id":"PROD6","name":"Paracetamol","expiration":"2027-01-01T00:00:00Z","price":180,"quantity":50}]'
output=$(cli_menu "$PROFILE" "3\nMERCHANT3\n${PRODUCTS_JSON}\n0")
echo "$output"
echo "$output" | grep -q "Products added successfully" || fail "AddProducts"
pass "AddProducts"

# ─────────────────────────────────────────────────────────────────────────────
section "4. Create User  [Org2Admin]"  # ← different identity!
# ─────────────────────────────────────────────────────────────────────────────
output=$(cli_menu "$PROFILE2" "4\nUSER3\nMilan\nMilanovic\nmilan@example.com\n0")
echo "$output"
echo "$output" | grep -q "User created successfully" || fail "CreateUser"
pass "CreateUser (Org2)"

# ─────────────────────────────────────────────────────────────────────────────
section "5. Deposit to User  [Org2Admin]"
# ─────────────────────────────────────────────────────────────────────────────
output=$(cli_menu "$PROFILE2" "5\nuser\nUSER3\n2000\n0")
echo "$output"
echo "$output" | grep -q "Deposited" || fail "Deposit to User"
pass "Deposit to User"

# ─────────────────────────────────────────────────────────────────────────────
section "6. Deposit to Merchant  [Org1Admin]"
# ─────────────────────────────────────────────────────────────────────────────
output=$(cli_menu "$PROFILE" "5\nmerchant\nMERCHANT3\n5000\n0")
echo "$output"
echo "$output" | grep -q "Deposited" || fail "Deposit to Merchant"
pass "Deposit to Merchant"

# ─────────────────────────────────────────────────────────────────────────────
section "7. Purchase Product  [Org2Admin]"
# ─────────────────────────────────────────────────────────────────────────────
output=$(cli_menu "$PROFILE2" "6\nUSER3\nPROD5\nINV_TEST_001\n2\n0")
echo "$output"
echo "$output" | grep -q "Purchase completed successfully" || fail "Purchase"
pass "Purchase"

# ─────────────────────────────────────────────────────────────────────────────
section "8. Purchase – Insufficient Funds (should fail gracefully)"
# ─────────────────────────────────────────────────────────────────────────────
# USER1 has only 500 deposited initially; we try to buy 200 * 150 = 30000
output=$(cli_menu "$PROFILE" "6\nUSER1\nPROD3\nINV_FAIL\n200\n0")
echo "$output"
echo "$output" | grep -qi "error\|insufficient\|failed" || fail "Purchase-insufficient-funds should have errored"
pass "Purchase – insufficient funds returns an error message"

# ─────────────────────────────────────────────────────────────────────────────
section "9. Purchase – Insufficient Stock (should fail gracefully)"
# ─────────────────────────────────────────────────────────────────────────────
# PROD2 (Hleb) has quantity 15; request 9999
output=$(cli_menu "$PROFILE" "6\nUSER1\nPROD2\nINV_FAIL2\n9999\n0")
echo "$output"
echo "$output" | grep -qi "error\|insufficient\|failed" || fail "Purchase-insufficient-stock should have errored"
pass "Purchase – insufficient stock returns an error message"

# ─────────────────────────────────────────────────────────────────────────────
section "10. Get All Products (range query)  [Org1Admin]"
# ─────────────────────────────────────────────────────────────────────────────
output=$(cli_menu "$PROFILE" "7\n0")
echo "$output"
echo "$output" | grep -q "PROD" || fail "GetAllProducts returned no products"
pass "GetAllProducts"

# ─────────────────────────────────────────────────────────────────────────────
section "11. Rich Query – filter by name  [Org1Admin]"
# ─────────────────────────────────────────────────────────────────────────────
# Menu 8: id=blank, name=Mleko, merchantType=blank, priceMin=blank, priceMax=blank
output=$(cli_menu "$PROFILE" "8\n\nMleko\n\n\n\n0")
echo "$output"
echo "$output" | grep -qi "mleko\|PROD1" || fail "RichQuery by name"
pass "RichQuery – by name"

# ─────────────────────────────────────────────────────────────────────────────
section "12. Rich Query – filter by merchant type  [Org1Admin]"
# ─────────────────────────────────────────────────────────────────────────────
output=$(cli_menu "$PROFILE" "8\n\n\nauto_parts\n\n\n0")
echo "$output"
echo "$output" | grep -qi "auto_parts\|PROD3\|PROD4" || fail "RichQuery by merchantType"
pass "RichQuery – by merchantType"

# ─────────────────────────────────────────────────────────────────────────────
section "13. Rich Query – filter by price range  [Org2Admin]"
# ─────────────────────────────────────────────────────────────────────────────
output=$(cli_menu "$PROFILE2" "8\n\n\n\n10\n60\n0")
echo "$output"
echo "$output" | grep -qi "price\|PROD\|\"price\"" || fail "RichQuery by price range"
pass "RichQuery – by price range"

# ─────────────────────────────────────────────────────────────────────────────
section "14. Rich Query – combined filter  [Org2Admin]"
# ─────────────────────────────────────────────────────────────────────────────
output=$(cli_menu "$PROFILE2" "8\n\nFilter\nsupermarket\n\n100\n0")
echo "$output"
echo "$output" | { grep -qi "PROD\|\[\]" || true; }  # empty result is also valid JSON
echo "$output" | grep -qi "error" && fail "RichQuery combined returned an error"
pass "RichQuery – combined filters"

# ─────────────────────────────────────────────────────────────────────────────
section "15. Login with second identity (Org2Admin) independently"
# ─────────────────────────────────────────────────────────────────────────────
output=$(cli_menu "$PROFILE2" "7\n0")
echo "$output"
echo "$output" | grep -q "PROD" || fail "Org2Admin – GetAllProducts"
pass "Org2Admin can query independently"

# ─────────────────────────────────────────────────────────────────────────────
section "16. Error handling – entity not found"
# ─────────────────────────────────────────────────────────────────────────────
output=$(cli_menu "$PROFILE" "6\nNONEXISTENT_USER\nPROD1\nINV_NOUSER\n1\n0")
echo "$output"
echo "$output" | grep -qi "error\|not found\|failed" || fail "Purchase with nonexistent user should error"
pass "Error – nonexistent user"

# ─────────────────────────────────────────────────────────────────────────────
echo ""
echo -e "${GREEN}══════════════════════════════════════════${NC}"
echo -e "${GREEN}  All tests passed! ✓${NC}"
echo -e "${GREEN}══════════════════════════════════════════${NC}"

# ─────────────────────────────────────────────────────────────────────────────
section "0.5. Register & Enroll User  [Org1Admin]"
# ─────────────────────────────────────────────────────────────────────────────
BASE_DIR=$(pwd)

ADMIN_MSP="$BASE_DIR/../fabric-samples/test-network/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp"
CA_CERT="$BASE_DIR/../fabric-samples/test-network/organizations/fabric-ca/org1/ca-cert.pem"

USERNAME="davor"
PASSWORD="davor123"

ENROLL_DIR="$BASE_DIR/wallet/Org1MSP/${USERNAME}/msp"

echo "→ Registering user ${USERNAME} via Fabric CA"

FABRIC_CA_CLIENT_HOME="$ADMIN_MSP" \
fabric-ca-client register \
  --id.name "$USERNAME" \
  --id.secret "$PASSWORD" \
  --id.type client \
  -u https://localhost:7054 \
  --caname ca-org1 \
  --tls.certfiles "$CA_CERT"

echo "→ Enrolling user ${USERNAME}..."
FABRIC_CA_CLIENT_HOME="$ADMIN_MSP" \
fabric-ca-client enroll \
  -u https://${USERNAME}:${PASSWORD}@localhost:7054 \
  --caname ca-org1 \
  --tls.certfiles "$CA_CERT" \
  -M "$ENROLL_DIR"

echo "✓ User '${USERNAME}' enrolled successfully in $ENROLL_DIR"
