#!/bin/bash
peer chaincode invoke -C mychannel -n trading -c '{"function":"Purchase","Args":["USER1","PROD1","MERCHANT1","INV1","2"]}' --waitForEvent
