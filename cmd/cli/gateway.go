package main

import (
	"context"
	"crypto/x509"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// const (
// 	mspID        = "Org1MSP"
// 	cryptoPath   = "../../fabric-samples/test-network/organizations/peerOrganizations/org1.example.com"
// 	certPath     = cryptoPath + "/users/User1@org1.example.com/msp/signcerts"
// 	keyPath      = cryptoPath + "/users/User1@org1.example.com/msp/keystore"
// 	tlsCertPath  = cryptoPath + "/peers/peer0.org1.example.com/tls/ca.crt"
// 	peerEndpoint = "dns:///localhost:7051"
// 	gatewayPeer  = "peer0.org1.example.com"
// )

func newGrpcConnection(config *OrgConfig) *grpc.ClientConn {
	certificatePEM, err := os.ReadFile(config.tlsCertPath)
	if err != nil {
		panic(fmt.Errorf("failed to read TLS cert: %w", err))
	}
	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(certificatePEM)
	creds := credentials.NewClientTLSFromCert(certPool, config.gatewayPeer)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(
		ctx,
		config.peerEndpoint,
		grpc.WithTransportCredentials(creds),
		grpc.WithBlock(),
	)
	if err != nil {
		panic(fmt.Errorf("failed to connect to peer: %w", err))
	}
	return conn
}

func newIdentity(config *OrgConfig) *identity.X509Identity {
	certPEM, err := readFirstFile(config.certPath)
	if err != nil {
		panic(fmt.Errorf("failed to read cert: %w", err))
	}
	cert, err := identity.CertificateFromPEM(certPEM)
	if err != nil {
		panic(err)
	}
	id, err := identity.NewX509Identity(config.mspID, cert)
	if err != nil {
		panic(err)
	}
	return id
}

func newSign(config *OrgConfig) identity.Sign {
	keyPEM, err := readFirstFile(config.keyPath)
	if err != nil {
		panic(fmt.Errorf("failed to read private key: %w", err))
	}
	key, err := identity.PrivateKeyFromPEM(keyPEM)
	if err != nil {
		panic(err)
	}
	sign, err := identity.NewPrivateKeySign(key)
	if err != nil {
		panic(err)
	}
	return sign
}

func readFirstFile(dirPath string) ([]byte, error) {
	files, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no files in %s", dirPath)
	}
	return os.ReadFile(path.Join(dirPath, files[0].Name()))
}
