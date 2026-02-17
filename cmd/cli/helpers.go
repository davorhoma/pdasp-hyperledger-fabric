package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Copy MSP from ca-client to alice folder
func copyUserCerts(org, user string) error {
	src := filepath.Join("wallets", org, "ca-client", "msp")
	dst := filepath.Join("wallets", org, user, "msp")

	err := os.MkdirAll(dst, 0755)
	if err != nil {
		return err
	}

	// jednostavno kopiranje fajlova (cacerts, keystore, signcerts, user)
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		err = copyDir(srcPath, dstPath)
		if err != nil {
			return err
		}
	}
	return nil
}

func copyDir(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}

	if info.IsDir() {
		if err := os.MkdirAll(dst, info.Mode()); err != nil {
			return err
		}

		entries, err := os.ReadDir(src)
		if err != nil {
			return err
		}

		for _, entry := range entries {
			srcPath := filepath.Join(src, entry.Name())
			dstPath := filepath.Join(dst, entry.Name())
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		}
	} else {
		srcFile, err := os.Open(src)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		dstFile, err := os.Create(dst)
		if err != nil {
			return err
		}
		defer dstFile.Close()

		if _, err := io.Copy(dstFile, srcFile); err != nil {
			return err
		}
	}

	return nil
}

// copyMSPToWallet kopira sertifikate iz MSP strukture u wallet format
func copyMSPToWallet(org, username string) error {
	caConfig := getCAConfig(org)

	srcMSP := filepath.Join(caConfig.walletPath, "ca-client", "msp")
	dstWallet := filepath.Join(caConfig.walletPath, username)

	// Kreiraj wallet direktorijum
	if err := os.MkdirAll(dstWallet, 0755); err != nil {
		return fmt.Errorf("failed to create wallet dir: %w", err)
	}

	// Kopiraj signcert (sertifikat)
	signcerts, err := os.ReadDir(filepath.Join(srcMSP, "signcerts"))
	if err != nil {
		return fmt.Errorf("failed to read signcerts: %w", err)
	}
	if len(signcerts) == 0 {
		return fmt.Errorf("no signcert found in MSP")
	}
	signcertSrc := filepath.Join(srcMSP, "signcerts", signcerts[0].Name())
	signcertDst := filepath.Join(dstWallet, "certificate.pem")
	data, err := os.ReadFile(signcertSrc)
	if err != nil {
		return fmt.Errorf("failed to read signcert: %w", err)
	}
	if err := os.WriteFile(signcertDst, data, 0644); err != nil {
		return fmt.Errorf("failed to write certificate.pem: %w", err)
	}

	// Kopiraj privatni ključ
	keystore, err := os.ReadDir(filepath.Join(srcMSP, "keystore"))
	if err != nil {
		return fmt.Errorf("failed to read keystore: %w", err)
	}
	if len(keystore) == 0 {
		return fmt.Errorf("no private key found in MSP")
	}
	keySrc := filepath.Join(srcMSP, "keystore", keystore[0].Name())
	keyDst := filepath.Join(dstWallet, "private_key.pem")
	data, err = os.ReadFile(keySrc)
	if err != nil {
		return fmt.Errorf("failed to read private key: %w", err)
	}
	if err := os.WriteFile(keyDst, data, 0600); err != nil {
		return fmt.Errorf("failed to write private_key.pem: %w", err)
	}

	fmt.Printf("✅ MSP certificates copied to wallet for '%s'\n", username)
	return nil
}
