#!/bin/bash
peer chaincode query -C mychannel -n trading -c '{"function":"RichQueryProducts","Args":["{\"name\":\"Mleko\",\"merchantType\":\"supermarket\",\"priceMin\":10,\"priceMax\":60}"]}'
