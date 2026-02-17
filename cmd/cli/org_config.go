package main

type OrgConfig struct {
	mspID        string
	cryptoPath   string
	certPath     string
	keyPath      string
	tlsCertPath  string
	peerEndpoint string
	gatewayPeer  string
}

func getOrgConfig(org string) OrgConfig {
	basePath := "../../fabric-samples/test-network/organizations/peerOrganizations"

	switch org {
	case "org1":
		cryptoPath := basePath + "/org1.example.com"
		return OrgConfig{
			mspID:        "Org1MSP",
			cryptoPath:   cryptoPath,
			certPath:     cryptoPath + "/users/User1@org1.example.com/msp/signcerts",
			keyPath:      cryptoPath + "/users/User1@org1.example.com/msp/keystore",
			tlsCertPath:  cryptoPath + "/peers/peer0.org1.example.com/tls/ca.crt",
			peerEndpoint: "dns:///localhost:7051",
			gatewayPeer:  "peer0.org1.example.com",
		}
	case "org2":
		cryptoPath := basePath + "/org2.example.com"
		return OrgConfig{
			mspID:        "Org2MSP",
			cryptoPath:   cryptoPath,
			certPath:     cryptoPath + "/users/User1@org2.example.com/msp/signcerts",
			keyPath:      cryptoPath + "/users/User1@org2.example.com/msp/keystore",
			tlsCertPath:  cryptoPath + "/peers/peer0.org2.example.com/tls/ca.crt",
			peerEndpoint: "dns:///localhost:9051",
			gatewayPeer:  "peer0.org2.example.com",
		}
	case "org3":
		cryptoPath := basePath + "/org3.example.com"
		return OrgConfig{
			mspID:        "Org3MSP",
			cryptoPath:   cryptoPath,
			certPath:     cryptoPath + "/users/User1@org3.example.com/msp/signcerts",
			keyPath:      cryptoPath + "/users/User1@org3.example.com/msp/keystore",
			tlsCertPath:  cryptoPath + "/peers/peer0.org3.example.com/tls/ca.crt",
			peerEndpoint: "dns:///localhost:11051",
			gatewayPeer:  "peer0.org3.example.com",
		}
	default:
		panic("Unknown organization: " + org)
	}
}
