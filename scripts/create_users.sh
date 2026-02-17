#!/bin/bash
peer chaincode invoke -C mychannel -n trading -c '{"function":"CreateUser","Args":["USER1","Marko","Markovic","marko@example.com","1000"]}' --waitForEvent
peer chaincode invoke -C mychannel -n trading -c '{"function":"CreateUser","Args":["USER2","Ana","Anic","ana@example.com","500"]}' --waitForEvent
