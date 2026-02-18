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
USER_ID="${USER_ID:-u1}"
USER_FIRSTNAME="${USER_FIRSTNAME:-Davor}"
USER_LASTNAME="${USER_LASTNAME:-Homa}"
USER_EMAIL="${USER_EMAIL:-davor.homa@example.com}"

echo "══════════════════════════════════════════"
echo "  TEST CreateUser invoke"
echo "══════════════════════════════════════════"
echo "USER:   $USER_ID"
echo "FIRSTNAME: $USER_FIRSTNAME"
echo "LASTNAME:  $USER_LASTNAME"
echo "EMAIL:  $USER_EMAIL"
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
  -c "{\"function\":\"CreateUser\",\"Args\":[\"$USER_ID\",\"$USER_FIRSTNAME\",\"$USER_LASTNAME\",\"$USER_EMAIL\"]}" \
  --waitForEvent

# echo
# echo "⏳ Čekam da se transakcija potvrdi..."
# sleep 2

# echo
# echo "══════════════════════════════════════════"
# echo "  PROVERA – Query user iz ledger-a"
# echo "══════════════════════════════════════════"

# QUERY_RESULT=$(peer chaincode query \
#   -C "$CHANNEL_NAME" \
#   -n "$CHAINCODE_NAME" \
#   -c "{\"function\":\"GetUserByID\",\"Args\":[\"$USER_ID\"]}" 2>/dev/null || true)

# if [[ -z "$QUERY_RESULT" ]]; then
#   echo "❌ Query nije vratio podatke."
#   exit 1
# fi

# echo "Ledger odgovor:"
# echo "$QUERY_RESULT"
# echo

# if echo "$QUERY_RESULT" | grep -q "$USER_ID"; then
#   echo "✔ SUCCESS — User je uspešno upisan u ledger."
# else
#   echo "❌ User nije pronađen nakon invoke-a."
#   exit 1
# fi
