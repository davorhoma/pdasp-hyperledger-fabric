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
MERCHANT_ID="${MERCHANT_ID:-MERCHANT1}"

# Raw JSON niz proizvoda
RAW_PRODUCTS_JSON='[
  {"ID":"PROD_NEW1","Name":"Test Proizvod 1","Expiration":"2027-12-31T23:59:59Z","Price":100,"Quantity":5},
  {"ID":"PROD_NEW2","Name":"Test Proizvod 2","Expiration":"2027-11-30T23:59:59Z","Price":200,"Quantity":3}
]'

# Escape u jedan string za CLI
PRODUCTS_JSON=$(echo "$RAW_PRODUCTS_JSON" | jq -c . | jq -R .)



echo "══════════════════════════════════════════"
echo "  TEST Add 2 Products for merchant $MERCHANT_ID"
echo "══════════════════════════════════════════"
echo "Merchant ID: $MERCHANT_ID"
echo "Products: $PRODUCTS_JSON"
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
  -c "{\"function\":\"AddProducts\",\"Args\":[\"$MERCHANT_ID\",$PRODUCTS_JSON]}" \
  --waitForEvent


echo
echo "✔ Products added to merchant $MERCHANT_ID successfully."

# ─────────────────────────────────────────────────────────────
# Test podaci 2
# ─────────────────────────────────────────────────────────────
MERCHANT_ID="${MERCHANT_ID:-MERCHANT1}"

# Raw JSON niz proizvoda
RAW_PRODUCTS_JSON='[
  {"ID":"PROD_NEW1","Name":"Test Proizvod 3","Expiration":"2027-12-31T23:59:59Z","Price":150,"Quantity":8}
]'

# Escape u jedan string za CLI
PRODUCTS_JSON=$(echo "$RAW_PRODUCTS_JSON" | jq -c . | jq -R .)

echo "══════════════════════════════════════════"
echo "  TEST Add 1 Product for merchant $MERCHANT_ID"
echo "══════════════════════════════════════════"
echo "Merchant ID: $MERCHANT_ID"
echo "Products: $PRODUCTS_JSON"
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
  -c "{\"function\":\"AddProducts\",\"Args\":[\"$MERCHANT_ID\",$PRODUCTS_JSON]}" \
  --waitForEvent


echo
echo "✔ Products added to merchant $MERCHANT_ID successfully."