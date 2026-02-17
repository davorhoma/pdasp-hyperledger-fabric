#!/bin/bash
peer chaincode invoke -C mychannel -n trading -c '{"function":"CreateMerchant","Args":["MERCHANT1","supermarket","123456789","1000"]}' --waitForEvent
peer chaincode invoke -C mychannel -n trading -c '{"function":"CreateProduct","Args":["MERCHANT1","PROD1","Mleko","2026-12-31","50","100"]}' --waitForEvent
peer chaincode invoke -C mychannel -n trading -c '{"function":"CreateProduct","Args":["MERCHANT1","PROD2","Hleb","","30","50"]}' --waitForEvent
