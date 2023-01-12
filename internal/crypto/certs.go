package crypto

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"reward/internal/util"
)

type certificateComponents struct {
	subject   pkix.Name
	dnsNames  []string
	privKey   *rsa.PrivateKey
	caCert    *x509.Certificate
	caPrivKey *rsa.PrivateKey
}

var (
	ErrNoCAPath = fmt.Errorf("no path provided")
)

// CACertificateFilePath returns the CA certificate path based on caDir.
func (c *Client) CACertificateFilePath(caDir string) (string, error) {
	if caDir == "" {
		return "", ErrNoCAPath
	}

	return filepath.Join(caDir, "certs", "ca.cert.pem"), nil
}

// CAPrivKeyFilePath returns the CA privkey path based on caDir.
func (c *Client) CAPrivKeyFilePath(caDir string) (string, error) {
	if caDir == "" {
		return "", ErrNoCAPath
	}

	return filepath.Join(caDir, "private", "ca.key.pem"), nil
}

// CheckCACertificateExistInDir checks if the CA Certificate PEM file already exists in Dir.
func (c *Client) CheckCACertificateExistInDir(caDir string, dontAskRecreate ...bool) bool {
	caCertPemFilePath, err := c.CACertificateFilePath(caDir)
	if err != nil {
		return false
	}

	if len(dontAskRecreate) > 0 && dontAskRecreate[0] {
		return util.FileExists(caCertPemFilePath)
	}

	return util.CheckFileExistsAndRecreate(caCertPemFilePath)
}

// CreateCACertificate creates a Private Key and a Signed CA Certificate in PEM format and writes to file.
func (c *Client) CreateCACertificate(caDir string) error {
	var (
		caCertificatePEMFilePath = filepath.Join(caDir, "certs", "ca.cert.pem")
		caPrivateKeyPEMFilePath  = filepath.Join(caDir, "private", "ca.key.pem")
		caCertificatePEM         = new(bytes.Buffer)
		caPrivateKeyPEM          = new(bytes.Buffer)
	)

	log.Printf("Creating CA Certificate: %s...", caCertificatePEMFilePath)

	hostname, err := os.Hostname()
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	caPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	serialNumber, _ := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	caCertificate := &x509.Certificate{
		SerialNumber:          serialNumber,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		Subject: pkix.Name{
			CommonName: fmt.Sprintf("%s Proxy Local CA (%s)",
				cases.Title(language.English).String(c.config.AppName()),
				hostname),
			Organization: []string{c.config.AppName()},
			Country:      []string{"HU"},
		},
	}

	selfSignedCACertificate, err := x509.CreateCertificate(rand.Reader,
		caCertificate,
		caCertificate,
		&caPrivateKey.PublicKey,
		caPrivateKey)
	if err != nil {
		return fmt.Errorf("cannot create self-signed certificate: %w", err)
	}

	err = pem.Encode(caCertificatePEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: selfSignedCACertificate,
	})
	if err != nil {
		return fmt.Errorf("cannot encode certificate to pem: %w", err)
	}

	err = pem.Encode(caPrivateKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(caPrivateKey),
	})
	if err != nil {
		return fmt.Errorf("cannot encode private key to pem: %w", err)
	}

	err = util.CreateDirAndWriteToFile(caCertificatePEM.Bytes(), caCertificatePEMFilePath)
	if err != nil {
		return fmt.Errorf("cannot write certificate to file: %w", err)
	}

	err = util.CreateDirAndWriteToFile(caPrivateKeyPEM.Bytes(), caPrivateKeyPEMFilePath)
	if err != nil {
		return fmt.Errorf("cannot write private key to file: %w", err)
	}

	log.Printf("...CA Certificate created.")

	return nil
}

// InstallCACertificate installs the generated CA certificate.
func (c *Client) InstallCACertificate(caDir string) error {
	caCertificatePEMFilePath := filepath.Join(caDir, "certs", "ca.cert.pem")
	osDistro := util.OSDistro()

	switch osDistro {
	case "windows":
		return c.windowsInstallCACertificate(caCertificatePEMFilePath)

	case "darwin":
		return c.darwinInstallCACertificate(caCertificatePEMFilePath)

	case "ubuntu", "debian", "pop", "elementary", "linuxmint":
		return c.debianInstallCACertificate(caCertificatePEMFilePath)

	case "fedora", "centos":
		return c.rhelInstallCACertificate(caCertificatePEMFilePath)

	case "arch", "manjaro":
		return c.archInstallCACertificate(caCertificatePEMFilePath)

	default:
		return fmt.Errorf("your operating system is not supported. yet. :(")
	}
}

func (c *Client) archInstallCACertificate(caCertificatePEMFilePath string) error {
	log.Println("Installing CA certificate for Arch based Linux distribution (requires sudo privileges)...")

	// nolint: gosec
	cmd := exec.Command("/bin/sh",
		"-c",
		fmt.Sprintf("sudo cp -va %s /etc/ca-certificates/trust-source/anchors/%s-local-ca.cert.pem",
			caCertificatePEMFilePath,
			c.config.AppName()))

	log.Debugf("Running command: %s", cmd)

	out, err := cmd.CombinedOutput()

	log.Tracef("Command output: %s", out)

	if err != nil {
		return fmt.Errorf("error running command: %w", err)
	}

	cmd = exec.Command("/bin/sh", "-c", "sudo update-ca-trust")

	log.Debugf("Running command: %s", cmd)

	out, err = cmd.CombinedOutput()

	log.Tracef("Command output: %s", out)

	if err != nil {
		return fmt.Errorf("error updating system ca certificate trust: %w", err)
	}

	log.Println("...CA Certificate installed.")

	return nil
}

func (c *Client) rhelInstallCACertificate(caCertificatePEMFilePath string) error {
	log.Println("Installing CA certificate for RHEL based Linux distribution (requires sudo privileges)...")

	// nolint: gosec
	cmd := exec.Command("/bin/sh",
		"-c",
		fmt.Sprintf("sudo cp -va %s /etc/pki/ca-trust/source/anchors/%s-local-ca.cert.pem",
			caCertificatePEMFilePath,
			c.config.AppName()),
	)

	log.Debugf("Running command: %s", cmd)

	out, err := cmd.CombinedOutput()

	log.Tracef("Command output: %s", out)

	if err != nil {
		return fmt.Errorf("error copying ca certificate: %w", err)
	}

	cmd = exec.Command("/bin/sh", "-c", "sudo update-ca-trust")

	log.Debugf("Running command: %s", cmd)

	out, err = cmd.CombinedOutput()

	log.Tracef("Command output: %s", out)

	if err != nil {
		return fmt.Errorf("error updating system ca certificate trust: %w", err)
	}

	log.Println("...CA Certificate installed.")

	return nil
}

func (c *Client) debianInstallCACertificate(caCertificatePEMFilePath string) error {
	log.Println("Installing CA Certificate for Debian based Linux distribution (requires sudo privileges)...")

	// nolint: gosec
	cmd := exec.Command("/bin/sh",
		"-c",
		fmt.Sprintf("sudo cp -va %s /usr/local/share/ca-certificates/%s-local-ca.cert.pem",
			caCertificatePEMFilePath,
			c.config.AppName()),
	)

	log.Debugf("Running command: %s", cmd)

	out, err := cmd.CombinedOutput()

	log.Tracef("Command output: %s", out)

	if err != nil {
		return fmt.Errorf("error copying ca certificate: %w", err)
	}

	cmd = exec.Command("/bin/sh", "-c", "sudo update-ca-certificates")

	log.Debugf("Running command: %s", cmd)

	out, err = cmd.CombinedOutput()

	log.Tracef("Command output: %s", out)

	if err != nil {
		return fmt.Errorf("error updating system ca certificate trust: %w", err)
	}

	log.Println("...CA Certificate installed.")

	return nil
}

func (c *Client) darwinInstallCACertificate(caCertificatePEMFilePath string) error {
	log.Println("Installing CA Certificate for macOS (requires sudo privileges)...")

	cmd := exec.Command(
		"sudo", "security", "add-trusted-cert", "-d", "-r",
		"trustRoot", "-k", "/Library/Keychains/System.keychain", caCertificatePEMFilePath,
	)

	log.Debugf("Running command: %s", cmd)

	out, err := cmd.CombinedOutput()

	log.Tracef("Command output: %s", out)

	if err != nil {
		return fmt.Errorf("error installing ca certificate: %w", err)
	}

	log.Println("...CA Certificate installed.")

	return nil
}

func (c *Client) windowsInstallCACertificate(caCertificatePEMFilePath string) error {
	log.Println("Installing CA Certificate for Windows (requires admin privileges)...")

	if !util.IsAdmin() {
		return fmt.Errorf("not enough privileges to install CA Certificate")
	}

	cmd := exec.Command("certutil", "-addstore", "-f", "Root", caCertificatePEMFilePath)

	log.Debugf("Running command: %s", cmd)

	out, err := cmd.CombinedOutput()

	log.Tracef("Command output: %s", out)

	if err != nil {
		return fmt.Errorf("error installing ca certificate: %w", err)
	}

	log.Println("...CA Certificate installed.")

	return nil
}

// CreatePrivateKeyAndCertificate creates a Private Key and a Certificate signed by caCertificate
// and writes to file in PEM format.
func (c *Client) CreatePrivateKeyAndCertificate(
	certificateDir string, certificateName string,
	dnsNames []string, caCertificateFilePath, caPrivateKeyFilePath string,
) error {
	log.Printf("Creating private key and certificate for %s...", dnsNames)

	r, err := os.ReadFile(caCertificateFilePath)
	if err != nil {
		return fmt.Errorf("cannot read ca certificate: %w", err)
	}

	block, _ := pem.Decode(r)

	caCertificate, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return fmt.Errorf("cannot parse ca certificate: %w", err)
	}

	r, err = os.ReadFile(caPrivateKeyFilePath)
	if err != nil {
		return fmt.Errorf("cannot read private key: %w", err)
	}

	block, _ = pem.Decode(r)

	caPrivateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return fmt.Errorf("cannot parse private key: %w", err)
	}

	privateKey, err := createPrivateKeyAndWriteToPemFile(2048,
		filepath.Join(certificateDir, certificateName+".key.pem"))
	if err != nil {
		return fmt.Errorf("cannot create private key: %w", err)
	}

	signedCert, err := c.createSignedCertificate(certificateComponents{
		subject: pkix.Name{
			CommonName:   certificateName,
			Organization: []string{c.config.AppName()},
			Country:      []string{"HU"},
		},
		dnsNames:  dnsNames,
		privKey:   privateKey,
		caCert:    caCertificate,
		caPrivKey: caPrivateKey,
	})
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	err = c.certificateWriteToPemFile(signedCert, filepath.Join(certificateDir, certificateName+".crt.pem"))
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	log.Println("...private key and certificate created.")

	return nil
}

func createPrivateKeyAndWriteToPemFile(bits int, privateKeyPEMFilePath string) (*rsa.PrivateKey, error) {
	log.Debugf("Creating Private Key and writing to file %s...", privateKeyPEMFilePath)

	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, fmt.Errorf("cannot generate rsa key: %w", err)
	}

	privateKeyPEM := new(bytes.Buffer)

	err = pem.Encode(
		privateKeyPEM, &pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	err = util.CreateDirAndWriteToFile(privateKeyPEM.Bytes(), privateKeyPEMFilePath)
	if err != nil {
		return nil, err
	}

	log.Debugln("...private key created.")

	return privateKey, nil
}

func (c *Client) createSignedCertificate(comps certificateComponents) ([]byte, error) {
	log.Println("Signing certificate...")

	serialNumber, _ := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	cert := &x509.Certificate{
		SerialNumber:          serialNumber,
		Subject:               comps.subject,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(1, 0, 0),
		IsCA:                  false,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		BasicConstraintsValid: true,
		DNSNames:              comps.dnsNames,
	}

	signedCert, err := x509.CreateCertificate(rand.Reader,
		cert,
		comps.caCert,
		&comps.privKey.PublicKey,
		comps.caPrivKey)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	log.Println("...signed certificate created.")

	return signedCert, nil
}

func (c *Client) certificateWriteToPemFile(cert []byte, certPEMFilePath string) error {
	log.Debugf("Writing certificate to file %s...", certPEMFilePath)

	certPem := new(bytes.Buffer)
	err := pem.Encode(
		certPem, &pem.Block{
			Type:  "CERTIFICATE",
			Bytes: cert,
		},
	)

	if err != nil {
		return fmt.Errorf("cannot encode certificate to pem: %w", err)
	}

	err = util.CreateDirAndWriteToFile(certPem.Bytes(), certPEMFilePath)
	if err != nil {
		return fmt.Errorf("cannot write certificate to file: %w", err)
	}

	log.Debugln("...certificate created.")

	return nil
}
