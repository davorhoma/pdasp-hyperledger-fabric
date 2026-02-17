package gateway

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// RegisterAndEnrollUser registers and enrolls a new user via Fabric CA.
// The enrolled MSP credentials are saved to fabric_cli/wallet/<org>/<username>/msp/
func RegisterAndEnrollUser(username, password, orgMSP string) (*ProfileInfo, error) {
	// 1. Resolve paths
	fabricSamplesDir, err := findFabricSamplesDir()
	if err != nil {
		return nil, fmt.Errorf("cannot find fabric-samples: %w", err)
	}

	testNetworkDir := filepath.Join(fabricSamplesDir, "test-network")
	caFolder := filepath.Join(testNetworkDir, "organizations", "fabric-ca", caFolderNameForOrg(orgMSP))
	tlsCertPath := filepath.Join(caFolder, "ca-cert.pem")
	caName := caNameForOrg(orgMSP)
	caPort := caPortForOrg(orgMSP)

	if _, err := os.Stat(tlsCertPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("CA TLS cert not found: %s (is the Fabric network running?)", tlsCertPath)
	}

	// 2. Determine wallet path (local to CLI binary)
	walletDir, err := getWalletDir(orgMSP, username)
	if err != nil {
		return nil, err
	}

	// 2a. Check if user is already enrolled locally
	certPath := filepath.Join(walletDir, "signcerts", "cert.pem")
	if _, err := os.Stat(certPath); err == nil {
		return nil, fmt.Errorf("user '%s' is already enrolled locally at %s\n"+
			"To re-enroll, delete the wallet directory first: rm -rf %s",
			username, walletDir, walletDir)
	}

	// 3. Enroll CA admin (if not already enrolled)
	adminMSPDir := filepath.Join(caFolder, "admin-msp")
	if _, err := os.Stat(adminMSPDir); os.IsNotExist(err) {
		fmt.Println("→ CA admin not enrolled yet, enrolling now...")
		if err := enrollCAAdmin(caFolder, tlsCertPath, caName, caPort); err != nil {
			return nil, fmt.Errorf("failed to enroll CA admin: %w", err)
		}
		fmt.Println("✓ CA admin enrolled")
	}

	// 4. Register user with CA
	fmt.Printf("→ Registering user '%s' with CA...\n", username)
	alreadyRegistered := false
	if err := registerUser(caFolder, tlsCertPath, caName, caPort, username, password); err != nil {
		errStr := err.Error()
		// Check if user already registered on CA server
		if strings.Contains(errStr, "already registered") || strings.Contains(errStr, "is already registered") {
			alreadyRegistered = true
			fmt.Println("  ⚠️  User already registered on CA (possibly with different password)")
			fmt.Println("  Attempting to enroll with provided password...")
		} else {
			return nil, fmt.Errorf("failed to register user: %w", err)
		}
	} else {
		fmt.Println("✓ User registered successfully")
	}

	// 5. Enroll user
	fmt.Printf("→ Enrolling user '%s'...\n", username)
	if err := enrollUser(caFolder, tlsCertPath, caName, caPort, username, password, walletDir); err != nil {
		if alreadyRegistered {
			return nil, fmt.Errorf("enrollment failed: user '%s' exists on CA but password is incorrect\n"+
				"Possible solutions:\n"+
				"  1. Use the correct password that was used during initial registration\n"+
				"  2. Contact CA admin to revoke and re-register this user\n"+
				"Original error: %w", username, err)
		}
		return nil, fmt.Errorf("failed to enroll user: %w", err)
	}
	fmt.Printf("✓ User '%s' enrolled successfully\n", username)
	fmt.Printf("  Credentials saved to: %s\n", walletDir)

	profileKey := strings.ToLower(username)
	name := fmt.Sprintf("%s (%s)", strings.Title(username), orgMSP)

	peerEndpoint, gatewayPeer, peerTLSCert := peerInfoForOrg(orgMSP)
	keyPath := filepath.Join(walletDir, "keystore")
	return &ProfileInfo{
		ProfileKey:   profileKey,
		Name:         name,
		OrgMSP:       orgMSP,
		PeerEndpoint: peerEndpoint,
		GatewayPeer:  gatewayPeer,
		PeerTLSCert:  peerTLSCert,
		CertPath:     certPath,
		KeyPath:      keyPath,
	}, nil
}

func enrollCAAdmin(caFolder, tlsCertPath, caName, caPort string) error {
	adminMSPDir := filepath.Join(caFolder, "admin-msp")
	os.MkdirAll(adminMSPDir, 0755)

	cmd := exec.Command(
		"fabric-ca-client",
		"enroll",
		"-u", fmt.Sprintf("https://admin:adminpw@localhost:%s", caPort),
		"--caname", caName,
		"--tls.certfiles", tlsCertPath,
		"-M", adminMSPDir,
	)

	cmd.Env = append(os.Environ(), "FABRIC_CA_CLIENT_HOME="+caFolder)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w\nOutput: %s", err, string(output))
	}
	return nil
}

func registerUser(caFolder, tlsCertPath, caName, caPort, username, password string) error {
	cmd := exec.Command(
		"fabric-ca-client",
		"register",
		"--id.name", username,
		"--id.secret", password,
		"--id.type", "client",
		"-u", fmt.Sprintf("https://localhost:%s", caPort),
		"--caname", caName,
		"--tls.certfiles", tlsCertPath,
		"-M", filepath.Join(caFolder, "admin-msp"), // Use admin MSP for registration
	)

	cmd.Env = append(os.Environ(), "FABRIC_CA_CLIENT_HOME="+caFolder)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w\nOutput: %s", err, string(output))
	}
	return nil
}

func enrollUser(caFolder, tlsCertPath, caName, caPort, username, password, walletDir string) error {
	os.MkdirAll(walletDir, 0755)

	cmd := exec.Command(
		"fabric-ca-client",
		"enroll",
		"-u", fmt.Sprintf("https://%s:%s@localhost:%s", username, password, caPort),
		"--caname", caName,
		"--tls.certfiles", tlsCertPath,
		"-M", walletDir,
	)

	cmd.Env = append(os.Environ(), "FABRIC_CA_CLIENT_HOME="+caFolder)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w\nOutput: %s", err, string(output))
	}
	return nil
}

func getWalletDir(orgMSP, username string) (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("cannot get current directory: %w", err)
	}
	return filepath.Join(cwd, "wallet", orgMSP, username, "msp"), nil
}

func findFabricSamplesDir() (string, error) {
	if dir := os.Getenv("FABRIC_SAMPLES_DIR"); dir != "" {
		if _, err := os.Stat(dir); err == nil {
			abs, _ := filepath.Abs(dir)
			return abs, nil
		}
	}

	cwd, _ := os.Getwd()
	candidates := []string{
		filepath.Join(cwd, "..", "fabric-samples"),
		filepath.Join(cwd, "..", "..", "fabric-samples"),
		filepath.Join(cwd, "fabric-samples"),
	}

	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			abs, _ := filepath.Abs(c)
			return abs, nil
		}
	}

	return "", fmt.Errorf("fabric-samples not found; set FABRIC_SAMPLES_DIR env var")
}

func caFolderNameForOrg(orgMSP string) string {
	switch orgMSP {
	case "Org1MSP":
		return "org1"
	case "Org2MSP":
		return "org2"
	case "Org3MSP":
		return "org3"
	case "OrdererMSP":
		return "ordererOrg"
	default:
		return ""
	}
}

func caPortForOrg(orgMSP string) string {
	switch orgMSP {
	case "Org1MSP":
		return "7054"
	case "Org2MSP":
		return "8054"
	case "Org3MSP":
		return "11054"
	case "OrdererMSP":
		return "9054"
	default:
		return ""
	}
}

func caNameForOrg(orgMSP string) string {
	switch orgMSP {
	case "Org1MSP":
		return "ca-org1"
	case "Org2MSP":
		return "ca-org2"
	case "Org3MSP":
		return "ca-org3"
	case "OrdererMSP":
		return "ca-orderer"
	default:
		return ""
	}
}

func peerInfoForOrg(orgMSP string) (peerEndpoint, gatewayPeer, peerTLSCert string) {
	switch orgMSP {
	case "Org1MSP":
		return "localhost:7051", "peer0.org1.example.com", "../fabric-samples/test-network/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt"
	case "Org2MSP":
		return "localhost:9051", "peer0.org2.example.com", "../fabric-samples/test-network/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt"
	case "Org3MSP":
		return "localhost:11051", "peer0.org3.example.com", "../fabric-samples/test-network/organizations/peerOrganizations/org3.example.com/peers/peer0.org3.example.com/tls/ca.crt"
	default:
		return "", "", ""
	}
}
