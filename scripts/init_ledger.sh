#!/bin/bash
peer chaincode invoke -C mychannel -n trading -c '{"function":"InitLedger","Args":[]}' --waitForEvent
