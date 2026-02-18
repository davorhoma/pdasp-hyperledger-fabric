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
ENTITY_TYPE="${ENTITY_TYPE:-user}"
ID="${ID:-u1}"
AMOUNT="${AMOUNT:-1000}"

echo "══════════════════════════════════════════"
echo "  TEST Deposit to entity ${ENTITY_TYPE} with ID ${ID}"
echo "══════════════════════════════════════════"
echo "ENTITY TYPE: $ENTITY_TYPE"
echo "ID: $ID"
echo "AMOUNT: $AMOUNT"
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
  -c "{\"function\":\"Deposit\",\"Args\":[\"$ENTITY_TYPE\",\"$ID\",\"$AMOUNT\"]}" \
  --waitForEvent