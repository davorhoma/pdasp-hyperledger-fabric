#!/bin/bash
# Kupovina nepostojećeg proizvoda
peer chaincode invoke -C mychannel -n trading -c '{"function":"Purchase","Args":["USER1","INVALID_PROD","MERCHANT1","INV2","1"]}' --waitForEvent

# Kupovina sa nedovoljno sredstava
peer chaincode invoke -C mychannel -n trading -c '{"function":"Purchase","Args":["USER2","PROD1","MERCHANT1","INV3","1000"]}' --waitForEvent
