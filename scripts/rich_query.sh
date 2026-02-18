#!/bin/bash
peer chaincode query -C channel1 -n trading -c '{"function":"RichQueryProducts","Args":["{\"name\":\"Mleko\",\"merchantType\":\"supermarket\",\"priceMin\":10,\"priceMax\":60}"]}'
