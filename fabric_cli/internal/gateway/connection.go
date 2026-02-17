package gateway

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Profile holds all connection parameters for a single org/user identity.
type Profile struct {
	Name         string `yaml:"name"`
	OrgMSP       string `yaml:"orgMSP"`
	PeerEndpoint string `yaml:"peerEndpoint"`
	PeerTLSCert  string `yaml:"peerTLSCert"`
	CertPath     string `yaml:"certPath"`
	KeyPath      string `yaml:"keyPath"`
	GatewayPeer  string `yaml:"gatewayPeer"`
}

type ProfileInfo struct {
	ProfileKey   string
	Name         string
	OrgMSP       string
	PeerEndpoint string
	GatewayPeer  string
	PeerTLSCert  string
	CertPath     string
	KeyPath      string
}

// Connection wraps a live Fabric Gateway connection.
type Connection struct {
	Profile  Profile
	grpcConn *grpc.ClientConn
	Gateway  *client.Gateway
	Network  *client.Network
	Contract *client.Contract
}

// Connect opens a gRPC + gateway connection and returns a ready Connection.
func Connect(p Profile, channelName, chaincodeName string) (*Connection, error) {
	tlsCert, err := loadCertificate(p.PeerTLSCert)
	if err != nil {
		return nil, fmt.Errorf("load TLS cert: %w", err)
	}

	creds := credentials.NewClientTLSFromCert(newCertPool(tlsCert), p.GatewayPeer)
	grpcConn, err := grpc.NewClient(p.PeerEndpoint, grpc.WithTransportCredentials(creds))
	if err != nil {
		return nil, fmt.Errorf("gRPC dial: %w", err)
	}

	id, err := newIdentity(p.OrgMSP, p.CertPath)
	if err != nil {
		grpcConn.Close()
		return nil, fmt.Errorf("new identity: %w", err)
	}

	sign, err := newSigner(p.KeyPath)
	if err != nil {
		grpcConn.Close()
		return nil, fmt.Errorf("new signer: %w", err)
	}

	gw, err := client.Connect(
		id,
		client.WithSign(sign),
		client.WithClientConnection(grpcConn),
		client.WithEvaluateTimeout(5*time.Second),
		client.WithEndorseTimeout(15*time.Second),
		client.WithSubmitTimeout(5*time.Second),
		client.WithCommitStatusTimeout(60*time.Second),
	)
	if err != nil {
		grpcConn.Close()
		return nil, fmt.Errorf("gateway connect: %w", err)
	}

	network := gw.GetNetwork(channelName)
	contract := network.GetContract(chaincodeName)

	return &Connection{
		Profile:  p,
		grpcConn: grpcConn,
		Gateway:  gw,
		Network:  network,
		Contract: contract,
	}, nil
}

// Close releases all resources.
func (c *Connection) Close() {
	c.Gateway.Close()
	c.grpcConn.Close()
}

// ─── helpers ────────────────────────────────────────────────────────────────

func loadCertificate(path string) (*x509.Certificate, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("no PEM block in %s", path)
	}
	return x509.ParseCertificate(block.Bytes)
}

func newCertPool(cert *x509.Certificate) *x509.CertPool {
	pool := x509.NewCertPool()
	pool.AddCert(cert)
	return pool
}

func newIdentity(mspID, certPath string) (*identity.X509Identity, error) {
	cert, err := loadCertificate(certPath)
	if err != nil {
		return nil, err
	}
	return identity.NewX509Identity(mspID, cert)
}

// newSigner loads the first private key PEM it finds in keyPath (file or dir).
func newSigner(keyPath string) (identity.Sign, error) {
	// Support both a direct PEM file and a Fabric CA "keystore" directory.
	stat, err := os.Stat(keyPath)
	if err != nil {
		return nil, err
	}

	var pemData []byte
	if stat.IsDir() {
		entries, err := os.ReadDir(keyPath)
		if err != nil {
			return nil, err
		}
		if len(entries) == 0 {
			return nil, fmt.Errorf("no key files in directory %s", keyPath)
		}
		pemData, err = os.ReadFile(filepath.Join(keyPath, entries[0].Name()))
		if err != nil {
			return nil, err
		}
	} else {
		pemData, err = os.ReadFile(keyPath)
		if err != nil {
			return nil, err
		}
	}

	privateKey, err := identity.PrivateKeyFromPEM(pemData)
	if err != nil {
		return nil, err
	}
	return identity.NewPrivateKeySign(privateKey)
}
