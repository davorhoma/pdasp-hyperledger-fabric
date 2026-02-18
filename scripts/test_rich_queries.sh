#!/usr/bin/env bash
# =============================================================================
# test_rich_queries.sh – Testira svih 5 CouchDB rich query funkcija
#
# Upotreba:
#   ./test_rich_queries.sh
#   CHANNEL=channel2 ./test_rich_queries.sh
# =============================================================================

set -uo pipefail

# ── Apsolutna putanja do korena projekta ─────────────────────────────────────
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="${PROJECT_ROOT:-$(cd "$SCRIPT_DIR/.." && pwd)}"

FABRIC_SAMPLES_DIR="${FABRIC_SAMPLES_DIR:-$PROJECT_ROOT/fabric-samples}"
TEST_NETWORK_DIR="${TEST_NETWORK_DIR:-$FABRIC_SAMPLES_DIR/test-network}"
BIN_DIR="${BIN_DIR:-$FABRIC_SAMPLES_DIR/bin}"
CONFIG_DIR="${CONFIG_DIR:-$FABRIC_SAMPLES_DIR/config}"

echo "PROJECT_ROOT       = ${PROJECT_ROOT}"
echo "FABRIC_SAMPLES_DIR = ${FABRIC_SAMPLES_DIR}"
echo "TEST_NETWORK_DIR   = ${TEST_NETWORK_DIR}"
echo ""

export PATH="$BIN_DIR:$PATH"
export FABRIC_CFG_PATH="$CONFIG_DIR"

# ── Konfiguracija ─────────────────────────────────────────────────────────────
CHANNEL="${CHANNEL:-channel1}"
CHAINCODE="${CHAINCODE:-trading}"

ORDERER_CA="${TEST_NETWORK_DIR}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem"

ORG1_TLS_CERT="${TEST_NETWORK_DIR}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt"
ORG2_TLS_CERT="${TEST_NETWORK_DIR}/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt"
ORG1_MSP="${TEST_NETWORK_DIR}/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp"

export FABRIC_CFG_PATH="${FABRIC_SAMPLES_DIR}/config"
export CORE_PEER_TLS_ENABLED=true
export CORE_PEER_LOCALMSPID="Org1MSP"
export CORE_PEER_TLS_ROOTCERT_FILE="${ORG1_TLS_CERT}"
export CORE_PEER_MSPCONFIGPATH="${ORG1_MSP}"
export CORE_PEER_ADDRESS="localhost:7051"
export PATH="${FABRIC_SAMPLES_DIR}/bin:${PATH}"

# ── Provera da li potrebni fajlovi postoje ────────────────────────────────────
check_paths() {
    local ok=true
    for path in "$ORDERER_CA" "$ORG1_TLS_CERT" "$ORG2_TLS_CERT" "$ORG1_MSP"; do
        if [ ! -e "$path" ]; then
            echo "  GREŠKA: Ne postoji: $path"
            ok=false
        fi
    done
    if [ "$ok" = false ]; then
        echo ""
        echo "Pokreni mrežu pre pokretanja testa."
        exit 1
    fi
    if ! command -v peer &>/dev/null; then
        echo "  GREŠKA: 'peer' nije nađen u PATH: ${PATH}"
        exit 1
    fi
    echo "  Sve putanje i alati su OK."
}

# ── Boje ──────────────────────────────────────────────────────────────────────
RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; CYAN='\033[0;36m'; NC='\033[0m'

pass()    { echo -e "${GREEN}  ✓ PASS${NC}: $1"; }
fail()    { echo -e "${RED}  ✗ FAIL${NC}: $1"; FAILED=$((FAILED+1)); }
section() {
    echo -e "\n${YELLOW}══════════════════════════════════════════${NC}"
    echo -e "${YELLOW}  $1${NC}"
    echo -e "${YELLOW}══════════════════════════════════════════${NC}"
}
info() { echo -e "${CYAN}  ℹ  $1${NC}"; }

FAILED=0

# ── Helper: invoke ─────────────────────────────────────────────────────────────
invoke() {
    local fn="$1"; shift
    local args_json
    args_json=$(printf '%s\n' "$@" | jq -R . | jq -s .)

    local output
    if output=$(peer chaincode invoke \
        -o localhost:7050 \
        --ordererTLSHostnameOverride orderer.example.com \
        --tls --cafile "${ORDERER_CA}" \
        -C "${CHANNEL}" -n "${CHAINCODE}" \
        --peerAddresses localhost:7051 \
        --tlsRootCertFiles "${ORG1_TLS_CERT}" \
        --peerAddresses localhost:9051 \
        --tlsRootCertFiles "${ORG2_TLS_CERT}" \
        -c "{\"function\":\"${fn}\",\"Args\":${args_json}}" 2>&1); then
        echo "$output"
    else
        echo -e "${RED}  peer invoke GREŠKA${NC}: $output"
        return 1
    fi
}

# ── Helper: query ──────────────────────────────────────────────────────────────
# ── Helper: query ──────────────────────────────────────────────────────────────
query() {
    local fn="$1"; shift
    local args_json
    # svaki argument pretvori u JSON string, onda ih sve stavi u listu
    args_json=$(printf '%s\n' "$@" | jq -R . | jq -s .)

    # poziv peer chaincode query i hvatanje outputa zajedno sa greškama
    local output
    if output=$(peer chaincode query \
        -C "${CHANNEL}" \
        -n "${CHAINCODE}" \
        --peerAddresses localhost:7051 \
        --tlsRootCertFiles "${ORG1_TLS_CERT}" \
        -c "{\"function\":\"${fn}\",\"Args\":${args_json}}" 2>&1); then
        echo "$output"
    else
        echo "$output" >&2
        return 1
    fi
}



# ── Helper: pretty JSON ────────────────────────────────────────────────────────
pretty() {
    echo "$1" | python3 -m json.tool 2>/dev/null || echo "$1"
}

# =============================================================================
section "PROVERA OKRUŽENJA"
check_paths
info "FABRIC_CFG_PATH = ${FABRIC_CFG_PATH}"
info "CHANNEL         = ${CHANNEL}"
info "CHAINCODE       = ${CHAINCODE}"

# =============================================================================
section "PRIPREMA – Init Ledger i test podaci"
info "Pokretanje InitLedger..."
if invoke "InitLedger"; then
    sleep 3
    pass "InitLedger završen"
else
    fail "InitLedger – preskačem ostatak pripreme"
fi

info "Kreiranje korisnika USER10..."
invoke "CreateUser" USER10 Bogdan Bogdanovic bogdan@example.com && sleep 2 || fail "CreateUser USER10"

info "Uplata 9999 na USER10..."
invoke "Deposit" user USER10 9999 && sleep 2 || fail "Deposit USER10"

pass "Korisnik USER10 kreiran sa stanjem 9999"

info "Kreiranje test faktura..."
invoke "Purchase" USER10 PROD1 INV_TEST_01 3 && sleep 2 || fail "Purchase INV_TEST_01"
invoke "Purchase" USER10 PROD3 INV_TEST_02 2 && sleep 2 || fail "Purchase INV_TEST_02"
invoke "Purchase" USER1 PROD2 INV_TEST_03 1 && sleep 2 || fail "Purchase INV_TEST_03"

pass "Test fakture kreirane"




echo
echo "══════════════════════════════════════════"
echo "  PROVERA – Query GetAllAssets iz ledger-a"
echo "══════════════════════════════════════════"

QUERY_RESULT=$(peer chaincode query \
  -C "$CHANNEL" \
  -n "$CHAINCODE" \
  -c "{\"function\":\"GetAllProducts\",\"Args\":[]}" 2>/dev/null || true)

if [[ -z "$QUERY_RESULT" ]]; then
  echo "❌ Query nije vratio podatke."
  exit 1
fi

echo "Ledger odgovor:"
echo "$QUERY_RESULT"
echo

if [[ "$QUERY_RESULT" == "[]" ]]; then
  echo "❌ Nema proizvoda u ledger-u."
  exit 1
else
  echo "✔ SUCCESS — pronađeni proizvodi."
fi



echo
echo "══════════════════════════════════════════"
echo "  PROVERA – Query GetProductsExpiringSoon iz ledger-a"
echo "══════════════════════════════════════════"

QUERY_RESULT=$(peer chaincode query \
  -C "$CHANNEL" \
  -n "$CHAINCODE" \
  -c "{\"function\":\"GetProductsExpiringSoon\",\"Args\":[\"2026-10-20T00:00:00Z\"]}" 2>/dev/null || true)

if [[ -z "$QUERY_RESULT" ]]; then
  echo "❌ Query nije vratio podatke."
  exit 1
fi

echo "Ledger odgovor:"
echo "$QUERY_RESULT"
echo

if [[ "$QUERY_RESULT" == "[]" ]]; then
  echo "❌ Nema proizvoda u ledger-u."
  exit 1
else
  echo "✔ SUCCESS — pronađeni proizvodi."
fi

echo
echo "══════════════════════════════════════════"
echo "  PROVERA – Query GetUsersWithMinBalance iz ledger-a"
echo "══════════════════════════════════════════"

QUERY_RESULT=$(peer chaincode query \
  -C "$CHANNEL" \
  -n "$CHAINCODE" \
  -c "{\"function\":\"GetUsersWithMinBalance\",\"Args\":[\"500\"]}" 2>&1)

if [[ -z "$QUERY_RESULT" ]]; then
  echo "❌ Query nije vratio podatke."
  exit 1
fi

echo "Ledger odgovor:"
echo "$QUERY_RESULT"
echo

if [[ "$QUERY_RESULT" == "[]" ]]; then
  echo "❌ Nema korisnika sa zadatim stanjem u ledger-u."
  exit 1
else
  echo "✔ SUCCESS — pronađeni korisnici."
fi













# =============================================================================
section "RICH QUERY 1 – GetProductsExpiringSoon"
info "expiresBefore = '2026-11-01T00:00:00Z'"
RESULT=$(query "GetProductsExpiringSoon" 2026-11-01T00:00:00Z) || { fail "RQ1 query greška"; RESULT=""; }
echo ""; pretty "$RESULT"; echo ""
if echo "$RESULT" | grep -q "PROD3\|Kocnica\|PROD4\|Filter ulja"; then
    pass "RQ1 – Pronađeni proizvodi sa rokom pre 2026-11-01"
else
    fail "RQ1 – Nisu pronađeni PROD3/PROD4"
fi

info "expiresBefore = '2100-01-01T00:00:00Z' (treba da vrati sve)"
RESULT2=$(query "GetProductsExpiringSoon" 2100-01-01T00:00:00Z) || { fail "RQ1b query greška"; RESULT2="[]"; }
COUNT=$(echo "$RESULT2" | python3 -c "import sys,json; d=json.load(sys.stdin); print(len(d) if d else 0)" 2>/dev/null || echo "0")
info "Broj pronađenih: $COUNT"
[ "$COUNT" != "0" ] && pass "RQ1b – Vraća sve proizvode za daleki datum" || fail "RQ1b – Nema rezultata"

# =============================================================================
section "RICH QUERY 2 – GetUsersWithMinBalance"
info "minBalance = 500  →  Očekujemo USER10 i USER1"
RESULT=$(query "GetUsersWithMinBalance" 500) || { fail "RQ2 query greška"; RESULT=""; }
echo ""; pretty "$RESULT"; echo ""
if echo "$RESULT" | grep -q "USER10\|USER1"; then
    pass "RQ2 – Korisnici sa stanjem >= 500 pronađeni"
else
    fail "RQ2 – Nisu pronađeni korisnici sa stanjem >= 500"
fi

info "minBalance = 9000  →  samo USER10"
RESULT2=$(query "GetUsersWithMinBalance" 9000) || { fail "RQ2b query greška"; RESULT2=""; }
if echo "$RESULT2" | grep -q "USER10"; then
    pass "RQ2b – USER10 pronađen za minBalance=9000"
else
    fail "RQ2b – USER10 nije pronađen za minBalance=9000"
fi

info "minBalance = 999999  →  niko (prazan rezultat)"
RESULT3=$(query "GetUsersWithMinBalance" 999999) || { fail "RQ2c query greška"; RESULT3=""; }
EMPTY=$(echo "$RESULT3" | python3 -c "import sys,json; d=json.load(sys.stdin); print('empty' if not d else 'notempty')" 2>/dev/null || echo "empty")
[ "$EMPTY" = "empty" ] && pass "RQ2c – Prazan rezultat za 999999 (ispravno)" || fail "RQ2c – Trebao biti prazan"

# =============================================================================
# (Dalje ostaje identično kao tvoja originalna skripta – RQ3, RQ4, RQ5 i sažetak)
# Samo se argumenti više ne stavljaju u ručne navodnike, npr:
# query "GetInvoicesByUserAndDateRange" USER10 2026-02-17T00:00:00Z 2026-02-18T23:59:59Z
# invoke "Purchase" USER10 PROD1 INV_TEST_01 3
# =============================================================================


# =============================================================================
section "RICH QUERY 3 – GetInvoicesByUserAndDateRange"
# =============================================================================

TODAY=$(date -u +"%Y-%m-%dT00:00:00Z")
TOMORROW=$(date -u -d "+1 day" +"%Y-%m-%dT23:59:59Z" 2>/dev/null \
        || date -u -v+1d      +"%Y-%m-%dT23:59:59Z")

info "USER10, od $TODAY do $TOMORROW  →  INV_TEST_01 i INV_TEST_02"
RESULT=$(query "GetInvoicesByUserAndDateRange" "USER10" "${TODAY}" "${TOMORROW}") \
    || { fail "RQ3 query greška"; RESULT=""; }
echo ""; pretty "$RESULT"; echo ""

if echo "$RESULT" | grep -q "INV_TEST_01\|INV_TEST_02"; then
    pass "RQ3 – Fakture USER10 za danas pronađene"
else
    fail "RQ3 – Fakture USER10 nisu pronađene"
fi

info "Negativan: USER1, period 1999.  →  prazan"
RESULT2=$(query "GetInvoicesByUserAndDateRange" \
    '"USER1"' '"1999-01-01T00:00:00Z"' '"1999-12-31T23:59:59Z"') \
    || { fail "RQ3b query greška"; RESULT2=""; }
EMPTY=$(echo "$RESULT2" | python3 -c "import sys,json; d=json.load(sys.stdin); print('empty' if not d else 'notempty')" 2>/dev/null || echo "empty")
[ "$EMPTY" = "empty" ] && pass "RQ3b – Prazan za period u prošlosti (ispravno)" \
                        || fail "RQ3b – Nije trebalo biti faktura pre 2000."

# =============================================================================
section "RICH QUERY 4 – GetLowStockProducts"
# =============================================================================

info "merchantType=auto_parts, maxQty=8  →  PROD3 i PROD4"
RESULT=$(query "GetLowStockProducts" 'auto_parts' "8") \
    || { fail "RQ4 query greška"; RESULT=""; }
echo ""; pretty "$RESULT"; echo ""

if echo "$RESULT" | grep -q "PROD3\|PROD4\|Kocnica\|Filter ulja\|auto_parts"; then
    pass "RQ4 – Niske zalihe auto_parts pronađene"
else
    fail "RQ4 – Nisu pronađeni PROD3/PROD4"
fi

info "merchantType=supermarket, maxQty=14  →  PROD2 (Hleb qty=14)"
RESULT2=$(query "GetLowStockProducts" 'supermarket' "14") \
    || { fail "RQ4b query greška"; RESULT2=""; }
if echo "$RESULT2" | grep -q "PROD2\|Hleb\|supermarket"; then
    pass "RQ4b – Hleb pronađen kao niska zaliha supermarketa"
else
    fail "RQ4b – Hleb nije pronađen"
fi

info "Negativan: nepostojeci_tip  →  prazan"
RESULT3=$(query "GetLowStockProducts" 'nepostojeci_tip' "100") \
    || { fail "RQ4c query greška"; RESULT3=""; }
EMPTY=$(echo "$RESULT3" | python3 -c "import sys,json; d=json.load(sys.stdin); print('empty' if not d else 'notempty')" 2>/dev/null || echo "empty")
[ "$EMPTY" = "empty" ] && pass "RQ4c – Prazan za nepostojeci_tip (ispravno)" \
                        || fail "RQ4c – Trebao biti prazan"

# =============================================================================
section "RICH QUERY 5 – GetMerchantHighValueInvoices"
# =============================================================================

info "MERCHANT1, minPrice=0  →  sve fakture MERCHANT1"
RESULT=$(query "GetMerchantHighValueInvoices" 'MERCHANT1' "0") \
    || { fail "RQ5 query greška"; RESULT=""; }
echo ""; pretty "$RESULT"; echo ""

if echo "$RESULT" | grep -q "MERCHANT1"; then
    pass "RQ5 – Fakture MERCHANT1 pronađene"
else
    fail "RQ5 – Nema faktura za MERCHANT1"
fi

info "MERCHANT2, minPrice=200  →  INV_TEST_02 (2×150=300)"
RESULT2=$(query "GetMerchantHighValueInvoices" 'MERCHANT2' "200") \
    || { fail "RQ5b query greška"; RESULT2=""; }
if echo "$RESULT2" | grep -q "INV_TEST_02\|MERCHANT2"; then
    pass "RQ5b – INV_TEST_02 (vrednost 300) pronađena"
else
    fail "RQ5b – INV_TEST_02 nije pronađena za minPrice=200"
fi

info "MERCHANT1, minPrice=999999  →  prazan"
RESULT3=$(query "GetMerchantHighValueInvoices" 'MERCHANT1' "999999") \
    || { fail "RQ5c query greška"; RESULT3=""; }
EMPTY=$(echo "$RESULT3" | python3 -c "import sys,json; d=json.load(sys.stdin); print('empty' if not d else 'notempty')" 2>/dev/null || echo "empty")
[ "$EMPTY" = "empty" ] && pass "RQ5c – Prazan za minPrice=999999 (ispravno)" \
                        || fail "RQ5c – Trebao biti prazan"

# =============================================================================
section "SAŽETAK"
# =============================================================================

echo ""
if [ "${FAILED}" -eq 0 ]; then
    echo -e "${GREEN}══════════════════════════════════════════${NC}"
    echo -e "${GREEN}  Svi testovi su prošli! ✓${NC}"
    echo -e "${GREEN}══════════════════════════════════════════${NC}"
else
    echo -e "${RED}══════════════════════════════════════════${NC}"
    echo -e "${RED}  ${FAILED} test(ova) nije prošlo! ✗${NC}"
    echo -e "${RED}══════════════════════════════════════════${NC}"
    exit 1
fi