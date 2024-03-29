package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"

	"github.com/rewardenv/reward/internal/config"
	"github.com/rewardenv/reward/pkg/util"
)

type Client struct {
	config *config.Config
}

func New(c *config.Config) *Client {
	return &Client{
		config: c,
	}
}

// generateRSAPrivateKey creates an RSA Private Key of specified byte size.
func (c *Client) generateRSAPrivateKey(bitSize int) (*rsa.PrivateKey, error) {
	log.Debugln("Generating RSA private key...")

	// Private Key generation
	privateKey, err := rsa.GenerateKey(rand.Reader, bitSize)
	if err != nil {
		return nil, fmt.Errorf("cannot generate private key: %w", err)
	}

	// Validate Private Key
	err = privateKey.Validate()
	if err != nil {
		return nil, fmt.Errorf("cannot validate private key: %w", err)
	}

	log.Debugln("...private key generated.")

	return privateKey, nil
}

// EncodeRSAPrivateKeyToPEM encodes Private Key from RSA to PEM format.
func (c *Client) EncodeRSAPrivateKeyToPEM(privateKey *rsa.PrivateKey) ([]byte, error) {
	log.Debugln("Encoding RSA private key to PEM...")

	err := privateKey.Validate()
	if err != nil {
		return nil, fmt.Errorf("cannot validate private key: %w", err)
	}

	return pem.EncodeToMemory(&pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   x509.MarshalPKCS1PrivateKey(privateKey),
	}), nil
}

// GenerateSSHPublicKey takes a rsa.PublicKey and return bytes suitable for writing to .pub file
// returns in the format "ssh-rsa ...".
func (c *Client) GenerateSSHPublicKey(publicKey *rsa.PublicKey) ([]byte, error) {
	log.Debugln("Generating SSH public key...")

	publicSSHKey, err := ssh.NewPublicKey(publicKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create public key: %w", err)
	}

	key := ssh.MarshalAuthorizedKey(publicSSHKey)

	log.Debugln("...SSH public key generated.")

	return key, nil
}

// GenerateSSHKeys generates and writes SSH keys with bitSize and saves them to path.
func (c *Client) GenerateSSHKeys(bitSize int, path string) error {
	log.Debugln("Generating SSH keys...")

	//nolint:gocritic
	privateKeyPath := filepath.Join(path)
	//nolint:gocritic
	publicKeyPath := filepath.Join(path + ".pub")

	privateKey, err := c.generateRSAPrivateKey(bitSize)
	if err != nil {
		return fmt.Errorf("cannot generate private key: %w", err)
	}

	privateKeyBytes, err := c.EncodeRSAPrivateKeyToPEM(privateKey)
	if err != nil {
		return fmt.Errorf("cannot encode private key to PEM: %w", err)
	}

	publicKeyBytes, err := c.GenerateSSHPublicKey(&privateKey.PublicKey)
	if err != nil {
		return fmt.Errorf("cannot generate ssh public key: %w", err)
	}

	err = util.CreateDirAndWriteToFile(privateKeyBytes, privateKeyPath, 0o600)
	if err != nil {
		return fmt.Errorf("cannot write private key to file: %w", err)
	}

	err = util.CreateDirAndWriteToFile(publicKeyBytes, publicKeyPath, 0o600)
	if err != nil {
		return fmt.Errorf("cannot write public key to file: %w", err)
	}

	log.Debugln("...SSH keys generated.")

	return nil
}
