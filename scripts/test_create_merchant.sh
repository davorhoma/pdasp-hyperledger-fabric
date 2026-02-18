#!/usr/bin/env bash
set -euo pipefail

# ─────────────────────────────────────────────────────────────
# Putanje nezavisne od lokacije pokretanja
# ─────────────────────────────────────────────────────────────
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="${PROJECT_ROOT:-$(cd "$SCRIPT_DIR/.." && pwd)}"

FABRIC_SAMPLES_DIR="${FABRIC_SAMPLES_DIR:-$PROJECT_ROOT/fabric-samples}"
TEST_NETWORK_DIR="${TEST_NETWORK_DIR:-$FABRIC_SAMPLES_DIR/test-network}"
BIN_DIR="${BIN_DIR:-$FABRIC_SAMPLES_DIR/bin}"
CONFIG_DIR="${CONFIG_DIR:-$FABRIC_SAMPLES_DIR/config}"

export PATH="$BIN_DIR:$PATH"
export FABRIC_CFG_PATH="$CONFIG_DIR"

# ─────────────────────────────────────────────────────────────
# Parametri (možeš override kroz env)
# ─────────────────────────────────────────────────────────────
CHANNEL_NAME="${CHANNEL_NAME:-channel1}"
CHAINCODE_NAME="${CHAINCODE_NAME:-trading}"
ORDERER_CA="${ORDERER_CA:-$TEST_NETWORK_DIR/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem}"

ORG1_PEER_TLS_ROOTCERT_FILE="${ORG1_PEER_TLS_ROOTCERT_FILE:-$TEST_NETWORK_DIR/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt}"
ORG1_MSP_PATH="${ORG1_MSP_PATH:-$TEST_NETWORK_DIR/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp}"

ORG2_PEER_TLS_ROOTCERT_FILE="${ORG2_PEER_TLS_ROOTCERT_FILE:-$TEST_NETWORK_DIR/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt}"
ORG2_PEER_ADDRESS="${ORG2_PEER_ADDRESS:-localhost:9051}"

export CORE_PEER_TLS_ENABLED=true
export CORE_PEER_LOCALMSPID="Org1MSP"
export CORE_PEER_TLS_ROOTCERT_FILE="$ORG1_PEER_TLS_ROOTCERT_FILE"
export CORE_PEER_MSPCONFIGPATH="$ORG1_MSP_PATH"
export CORE_PEER_ADDRESS="localhost:7051"

# ─────────────────────────────────────────────────────────────
# Test podaci
# ─────────────────────────────────────────────────────────────
MERCHANT_ID="${MERCHANT_ID:-m1}"
MERCHANT_TYPE="${MERCHANT_TYPE:-retail}"
MERCHANT_PIB="${MERCHANT_PIB:-123456789}"

echo "══════════════════════════════════════════"
echo "  TEST CreateMerchant invoke"
echo "══════════════════════════════════════════"
echo "ID:   $MERCHANT_ID"
echo "TYPE: $MERCHANT_TYPE"
echo "PIB:  $MERCHANT_PIB"
echo

peer chaincode invoke \
  -o localhost:7050 \
  --ordererTLSHostnameOverride orderer.example.com \
  --tls \
  --cafile "$ORDERER_CA" \
  -C "$CHANNEL_NAME" \
  -n "$CHAINCODE_NAME" \
  --peerAddresses localhost:7051 \
  --tlsRootCertFiles "$ORG1_PEER_TLS_ROOTCERT_FILE" \
  --peerAddresses "$ORG2_PEER_ADDRESS" \
  --tlsRootCertFiles "$ORG2_PEER_TLS_ROOTCERT_FILE" \
  -c "{\"function\":\"CreateMerchant\",\"Args\":[\"$MERCHANT_ID\",\"$MERCHANT_TYPE\",\"$MERCHANT_PIB\"]}" \
  --waitForEvent

echo
echo "⏳ Čekam da se transakcija potvrdi..."
sleep 2

echo
echo "══════════════════════════════════════════"
echo "  PROVERA – Query merchant iz ledger-a"
echo "══════════════════════════════════════════"

QUERY_RESULT=$(peer chaincode query \
  -C "$CHANNEL_NAME" \
  -n "$CHAINCODE_NAME" \
  -c "{\"function\":\"GetMerchantByID\",\"Args\":[\"$MERCHANT_ID\"]}" 2>/dev/null || true)

if [[ -z "$QUERY_RESULT" ]]; then
  echo "❌ Query nije vratio podatke."
  exit 1
fi

echo "Ledger odgovor:"
echo "$QUERY_RESULT"
echo

if echo "$QUERY_RESULT" | grep -q "$MERCHANT_ID"; then
  echo "✔ SUCCESS — Merchant je uspešno upisan u ledger."
else
  echo "❌ Merchant nije pronađen nakon invoke-a."
  exit 1
fi
