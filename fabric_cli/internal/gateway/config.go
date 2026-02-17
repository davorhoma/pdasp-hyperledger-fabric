package gateway

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config is the top-level structure of config.yaml.
type Config struct {
	Channel   string             `yaml:"channel"`
	Chaincode string             `yaml:"chaincode"`
	Profiles  map[string]Profile `yaml:"profiles"`
}

// LoadConfig reads config.yaml from the given path.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	if len(cfg.Profiles) == 0 {
		return nil, fmt.Errorf("config must define at least one profile")
	}
	return &cfg, nil
}

func AddProfileToConfig(configPath string, profile *ProfileInfo) error {
	cfg := &Config{
		Profiles: make(map[string]Profile),
	}

	if _, err := os.Stat(configPath); err == nil {
		data, err := os.ReadFile(configPath)
		if err != nil {
			return fmt.Errorf("cannot read config file: %w", err)
		}
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return fmt.Errorf("cannot parse YAML: %w", err)
		}
	}

	cfg.Profiles[profile.ProfileKey] = Profile{
		Name:         profile.Name,
		OrgMSP:       profile.OrgMSP,
		PeerEndpoint: profile.PeerEndpoint,
		GatewayPeer:  profile.GatewayPeer,
		PeerTLSCert:  profile.PeerTLSCert,
		CertPath:     profile.CertPath,
		KeyPath:      profile.KeyPath,
	}

	out, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("cannot marshal YAML: %w", err)
	}

	if err := os.WriteFile(configPath, out, 0644); err != nil {
		return fmt.Errorf("cannot write config file: %w", err)
	}

	fmt.Printf("Profile '%s' added to config.yaml\n", profile.ProfileKey)
	return nil
}

func WalletPaths(orgMSP, username string) (certPath, keyPath string, err error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", "", fmt.Errorf("cannot get cwd: %w", err)
	}
	mspDir := filepath.Join(cwd, "wallet", orgMSP, username, "msp")
	return filepath.Join(mspDir, "signcerts", "cert.pem"), filepath.Join(mspDir, "keystore"), nil
}
