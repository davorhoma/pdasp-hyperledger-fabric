package main

import (
	"crypto/x509"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hyperledger/fabric-ca/api"
	"github.com/hyperledger/fabric-ca/lib"
	"github.com/hyperledger/fabric-ca/lib/tls"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
)

// EnrollUser registruje novog korisnika sa Fabric CA i čuva kredencijale u wallet
// EnrollUser registruje i enroll-uje korisnika, čuva sertifikat i ključ u wallet
func EnrollUser(org, username, password string) error {
	fmt.Printf("==> Enrolling user '%s' for organization '%s'\n", username, org)

	// Učitaj CA konfiguraciju
	caConfig := getCAConfig(org)

	// Proveri da li identitet već postoji
	walletDir := filepath.Join(caConfig.walletPath, username)
	certPath := filepath.Join(walletDir, "certificate.pem")
	if _, err := os.Stat(certPath); err == nil {
		return fmt.Errorf("identity '%s' already exists in wallet", username)
	}

	// Kreiraj CA klijent
	client, err := createCAClient(caConfig)
	if err != nil {
		return fmt.Errorf("failed to create CA client: %w", err)
	}

	// Enroll korisnika preko CA
	enrollmentResponse, err := client.Enroll(&api.EnrollmentRequest{
		Name:   username,
		Secret: password,
	})
	if err != nil {
		return fmt.Errorf("failed to enroll user: %w", err)
	}

	fmt.Println("✅ Successfully enrolled user")

	// Kreiraj wallet direktorijum ako ne postoji
	err = os.MkdirAll(walletDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create wallet directory: %w", err)
	}

	// Sačuvaj sertifikat
	certPEM := enrollmentResponse.Identity.GetECert().Cert()
	err = os.WriteFile(certPath, []byte(certPEM), 0644)
	if err != nil {
		return fmt.Errorf("failed to save certificate: %w", err)
	}

	// Sačuvaj privatni ključ
	// Sačuvaj privatni ključ
	keyPEM, err := identity.PrivateKeyToPEM(enrollmentResponse.Identity.GetECert().Key())
	if err != nil {
		return fmt.Errorf("failed to convert private key to PEM: %w", err)
	}

	keyPath := filepath.Join(walletDir, "private_key.pem")
	err = os.WriteFile(keyPath, keyPEM, 0600)
	if err != nil {
		return fmt.Errorf("failed to save private key: %w", err)
	}

	fmt.Printf("✅ Identity '%s' saved to wallet\n", username)
	return nil
}

// RegisterAndEnrollUser registruje novog korisnika preko admin-a, zatim ga enrolluje
// RegisterAndEnrollUser registruje novog korisnika preko admin-a, zatim ga enroll-uje
func RegisterAndEnrollUser(org, adminUsername, adminPassword, newUsername, newPassword string) error {
	fmt.Printf("==> Registering and enrolling user '%s' for organization '%s'\n", newUsername, org)

	caConfig := getCAConfig(org)

	// Kreiraj CA klijent
	client, err := createCAClient(caConfig)
	if err != nil {
		return fmt.Errorf("failed to create CA client: %w", err)
	}

	// Enroll admin da bi mogao da registruje novog korisnika
	adminEnrollment, err := client.Enroll(&api.EnrollmentRequest{
		Name:   adminUsername,
		Secret: adminPassword,
	})
	if err != nil {
		return fmt.Errorf("failed to enroll admin: %w", err)
	}

	// Registruj novog korisnika
	regReq := &api.RegistrationRequest{
		Name:           newUsername,
		Secret:         newPassword,
		Type:           "client",
		MaxEnrollments: -1,
		Affiliation:    "",
	}

	_, err = adminEnrollment.Identity.Register(regReq)
	if err != nil {
		fmt.Printf("⚠️  User may already be registered, trying to enroll anyway...\n")
	}

	// Enroll novog korisnika (kreira cert i private key u wallet)
	err = EnrollUser(org, newUsername, newPassword)
	if err != nil {
		return fmt.Errorf("failed to enroll user '%s': %w", newUsername, err)
	}

	return nil
}

func createCAClient(caConfig CAConfig) (*lib.Client, error) {
	client := &lib.Client{
		HomeDir: filepath.Join(caConfig.walletPath, "ca-client"),
		Config: &lib.ClientConfig{
			URL:    caConfig.caURL,
			CAName: caConfig.caName,
			TLS: tls.ClientTLSConfig{
				Enabled:   true,
				CertFiles: []string{caConfig.tlsCertPath}, // <-- ovde ide putanja
			},
		},
	}

	if err := client.Init(); err != nil {
		return nil, fmt.Errorf("failed to init CA client: %w", err)
	}

	return client, nil
}

// saveIdentityToWallet skladišti sertifikat i privatni ključ u wallet direktorijum
func saveIdentityToWallet(walletPath string, cert, key []byte) error {
	// Kreiraj wallet direktorijum
	err := os.MkdirAll(walletPath, 0755)
	if err != nil {
		return fmt.Errorf("failed to create wallet directory: %w", err)
	}

	// Sačuvaj sertifikat
	certPath := filepath.Join(walletPath, "certificate.pem")
	err = os.WriteFile(certPath, cert, 0644)
	if err != nil {
		return fmt.Errorf("failed to save certificate: %w", err)
	}

	// Sačuvaj privatni ključ
	keyPath := filepath.Join(walletPath, "private_key.pem")
	err = os.WriteFile(keyPath, key, 0600)
	if err != nil {
		return fmt.Errorf("failed to save private key: %w", err)
	}

	return nil
}

// CAConfig sadrži konfiguraciju za Fabric CA
type CAConfig struct {
	caURL       string
	caName      string
	tlsCertPath string
	walletPath  string
	mspID       string
	adminUser   string
	adminPass   string
}

// getCAConfig vraća CA konfiguraciju za datu organizaciju
func getCAConfig(org string) CAConfig {
	// NAPOMENA: Prilagodi putanje prema tvojoj Fabric mreži!
	basePath := "../../../../../fabric-samples/test-network/organizations/peerOrganizations"

	switch org {
	case "org1":
		return CAConfig{
			caURL:       "https://localhost:7054",
			caName:      "ca-org1",
			tlsCertPath: filepath.Join(basePath, "org1.example.com/ca/ca.org1.example.com-cert.pem"),
			walletPath:  "wallets/org1",
			mspID:       "Org1MSP",
			adminUser:   "admin",
			adminPass:   "adminpw",
		}
	case "org2":
		return CAConfig{
			caURL:       "https://localhost:8054",
			caName:      "ca-org2",
			tlsCertPath: filepath.Join(basePath, "org2.example.com/ca/ca.org2.example.com-cert.pem"),
			walletPath:  "wallets/org2",
			mspID:       "Org2MSP",
			adminUser:   "admin",
			adminPass:   "adminpw",
		}
	case "org3":
		return CAConfig{
			caURL:       "https://localhost:11054",
			caName:      "ca-org3",
			tlsCertPath: filepath.Join(basePath, "org3.example.com/ca/ca.org3.example.com-cert.pem"),
			walletPath:  "wallets/org3",
			mspID:       "Org3MSP",
			adminUser:   "admin",
			adminPass:   "adminpw",
		}
	default:
		panic("Unknown organization: " + org)
	}
}

// UseStaticIdentity učitava statički identitet direktno sa diska (bez CA)
// Ovo se koristi za testiranje kada nemaš CA ili za admin identitete
func UseStaticIdentity(org string) OrgConfig {
	fmt.Printf("==> Using STATIC identity for org '%s'\n", org)
	config := getOrgConfig(org)
	fmt.Printf("✅ Loaded static identity: %s\n", config.mspID)
	return config
}

// LoadIdentityFromWallet učitava već postojeći enrollovan identitet iz walleta
func LoadIdentityFromWallet(org, username string) (*identity.X509Identity, identity.Sign, error) {
	fmt.Printf("==> Loading identity '%s' from wallet for org '%s'\n", username, org)

	caConfig := getCAConfig(org)
	walletPath := filepath.Join(caConfig.walletPath, username)

	// Proveri da li wallet postoji
	if _, err := os.Stat(walletPath); os.IsNotExist(err) {
		return nil, nil, fmt.Errorf("identity '%s' does not exist in wallet. Run 'enroll' first", username)
	}

	// Učitaj sertifikat
	certPath := filepath.Join(walletPath, "certificate.pem")
	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read certificate: %w", err)
	}

	cert, err := identity.CertificateFromPEM(certPEM)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	// Kreiraj X509 identity
	id, err := identity.NewX509Identity(caConfig.mspID, cert)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create identity: %w", err)
	}

	// Učitaj privatni ključ
	keyPath := filepath.Join(walletPath, "private_key.pem")
	keyPEM, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read private key: %w", err)
	}

	privateKey, err := identity.PrivateKeyFromPEM(keyPEM)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	// Kreiraj sign funkciju
	sign, err := identity.NewPrivateKeySign(privateKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create sign function: %w", err)
	}

	fmt.Printf("✅ Identity '%s' loaded from wallet\n", username)
	return id, sign, nil
}

// ListWalletIdentities izlistava sve identitete u wallet-u za datu organizaciju
func ListWalletIdentities(org string) ([]string, error) {
	caConfig := getCAConfig(org)

	entries, err := os.ReadDir(caConfig.walletPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to read wallet directory: %w", err)
	}

	var identities []string
	for _, entry := range entries {
		if entry.IsDir() {
			identities = append(identities, entry.Name())
		}
	}

	return identities, nil
}

// DeleteIdentityFromWallet briše identitet iz wallet-a
func DeleteIdentityFromWallet(org, username string) error {
	caConfig := getCAConfig(org)
	walletPath := filepath.Join(caConfig.walletPath, username)

	err := os.RemoveAll(walletPath)
	if err != nil {
		return fmt.Errorf("failed to delete identity: %w", err)
	}

	fmt.Printf("✅ Identity '%s' deleted from wallet\n", username)
	return nil
}

// GetIdentityInfo vraća informacije o identitetu iz walleta
func GetIdentityInfo(org, username string) (*x509.Certificate, error) {
	caConfig := getCAConfig(org)
	walletPath := filepath.Join(caConfig.walletPath, username)
	certPath := filepath.Join(walletPath, "certificate.pem")

	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate: %w", err)
	}

	cert, err := identity.CertificateFromPEM(certPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	return cert, nil
}
