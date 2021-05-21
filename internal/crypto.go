package internal

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

// GenerateRSAPrivateKey creates a RSA Private Key of specified byte size.
func GenerateRSAPrivateKey(bitSize int) (*rsa.PrivateKey, error) {
	log.Traceln("In function: GenerateRSAPrivateKey")

	// Private Key generation
	privateKey, err := rsa.GenerateKey(rand.Reader, bitSize)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	// Validate Private Key
	err = privateKey.Validate()
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	log.Println("Private Key generated")

	return privateKey, nil
}

// EncodeRSAPrivateKeyToPEM encodes Private Key from RSA to PEM format.
func EncodeRSAPrivateKeyToPEM(privateKey *rsa.PrivateKey) ([]byte, error) {
	log.Traceln("In function: EncodeRSAPrivateKeyToPEM")

	err := privateKey.Validate()
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	// Get ASN.1 DER format
	privDER := x509.MarshalPKCS1PrivateKey(privateKey)

	// pem.Block
	privBlock := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   privDER,
	}

	// Private key in PEM format
	privatePEM := pem.EncodeToMemory(&privBlock)

	return privatePEM, nil
}

// GenerateSSHPublicKey takes a rsa.PublicKey and return bytes suitable for writing to .pub file
// returns in the format "ssh-rsa ...".
func GenerateSSHPublicKey(publicKey *rsa.PublicKey) ([]byte, error) {
	log.Traceln("In function: GenerateSSHPublicKey")

	publicSSHKey, err := ssh.NewPublicKey(publicKey)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	sshPubKeyBytes := ssh.MarshalAuthorizedKey(publicSSHKey)

	log.Println("Public key generated")

	return sshPubKeyBytes, nil
}

// GenerateAndSaveSSHKeys generates and writes SSH keys with bitSize and saves them to pathToSave.
func GenerateAndSaveSSHKeys(bitSize int, pathToSave string) error {
	log.Traceln("In function: GenerateAndSaveSSHKeys")

	privKeyFilePath := filepath.Join(pathToSave)
	pubKeyFilePath := filepath.Join(pathToSave + ".pub")

	privateKey, err := GenerateRSAPrivateKey(bitSize)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	privateKeyPemBytes, err := EncodeRSAPrivateKeyToPEM(privateKey)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	publicKeyBytes, err := GenerateSSHPublicKey(&privateKey.PublicKey)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	err = CreateDirAndWriteBytesToFile(privateKeyPemBytes, privKeyFilePath, 0o600)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	err = CreateDirAndWriteBytesToFile(publicKeyBytes, pubKeyFilePath, 0o600)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}
