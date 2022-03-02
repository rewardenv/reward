package core

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

var (
	SslBaseDir  = "ssl"
	CaBaseDir   = "rootca"
	CertBaseDir = "certs"
)

type certificateComponents struct {
	subject   pkix.Name
	dnsNames  []string
	privKey   *rsa.PrivateKey
	caCert    *x509.Certificate
	caPrivKey *rsa.PrivateKey
}

// GetCaCertificateFilePath returns the CA certificate path based on caDir.
func GetCaCertificateFilePath(caDir string) (string, error) {
	log.Debugln()

	if caDir == "" {
		return "", errors.New("no path provided")
	}

	caPath := filepath.Join(caDir)
	caCertDirPath := filepath.Join(caPath, "certs")
	caCertPemFilePath := filepath.Join(caCertDirPath, "ca.cert.pem")

	return caCertPemFilePath, nil
}

// GetCaPrivKeyFilePath returns the CA privkey path based on caDir.
func GetCaPrivKeyFilePath(caDir string) (string, error) {
	log.Debugln()

	if caDir == "" {
		return "", errors.New("no path provided")
	}

	caPath := filepath.Join(caDir)
	caCertDirPath := filepath.Join(caPath, "private")
	caCertPemFilePath := filepath.Join(caCertDirPath, "ca.key.pem")

	return caCertPemFilePath, nil
}

// CheckCaCertificateExistInDir checks if the CA Certificate PEM file already exists in Dir.
func CheckCaCertificateExistInDir(caDir string, dontAskRecreate ...bool) bool {
	log.Debugln()

	caCertPemFilePath, err := GetCaCertificateFilePath(caDir)
	if err != nil {
		return false
	}

	if len(dontAskRecreate) > 0 && dontAskRecreate[0] {
		return CheckFileExists(caCertPemFilePath)
	}

	return CheckFileExistsAndRecreate(caCertPemFilePath)
}

// CreateCaCertificate creates a Private Key and a Signed CA Certificate in PEM format and writes to file.
func CreateCaCertificate(caDir string) error {
	log.Debugln()

	hostname, err := os.Hostname()
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	subject := pkix.Name{
		CommonName:   fmt.Sprintf("%v Proxy Local CA (%v)", strings.Title(AppName), hostname),
		Organization: []string{AppName},
		Country:      []string{"HU"},
	}

	caFullPath := filepath.Join(caDir)
	caCertDirPath := filepath.Join(caFullPath, "certs")
	caCertPemFilePath := filepath.Join(caCertDirPath, "ca.cert.pem")
	caPrivKeyDirPath := filepath.Join(caFullPath, "private")
	caPrivKeyPemFilePath := filepath.Join(caPrivKeyDirPath, "ca.key.pem")

	log.Printf("Creating CA Certificate: %v", caCertPemFilePath)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)

	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	caCert := &x509.Certificate{
		SerialNumber:          serialNumber,
		Subject:               subject,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
	}
	caPrivKey, err := rsa.GenerateKey(rand.Reader, 2048)

	if err != nil {
		return fmt.Errorf("%w", err)
	}

	selfSignedCaCert, err := x509.CreateCertificate(rand.Reader, caCert, caCert, &caPrivKey.PublicKey, caPrivKey)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	caCertPEM := new(bytes.Buffer)
	err = pem.Encode(
		caCertPEM, &pem.Block{
			Type:  "CERTIFICATE",
			Bytes: selfSignedCaCert,
		},
	)

	if err != nil {
		return fmt.Errorf("%w", err)
	}

	caPrivKeyPEM := new(bytes.Buffer)
	err = pem.Encode(
		caPrivKeyPEM, &pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(caPrivKey),
		},
	)

	if err != nil {
		return fmt.Errorf("%w", err)
	}

	fileMode := os.FileMode(0o640)

	log.Printf("Creating CA cert file...")
	log.Debugf("path: %v, mode: %v", caCertPemFilePath, fileMode)

	err = CreateDirAndWriteBytesToFile(caCertPEM.Bytes(), caCertPemFilePath)
	if err != nil {
		return err
	}

	err = CreateDirAndWriteBytesToFile(caPrivKeyPEM.Bytes(), caPrivKeyPemFilePath)
	if err != nil {
		return err
	}

	log.Printf("CA Certificate created.")
	log.Debugf("path: %v", caCertPemFilePath)

	return nil
}

// InstallCaCertificate installs the generated CA certificate.
func InstallCaCertificate(caDir string) error {
	log.Debugln()

	caPath := filepath.Join(caDir)
	caCertDirPath := filepath.Join(caPath, "certs")
	caCertPemFilePath := filepath.Join(caCertDirPath, "ca.cert.pem")

	osDistro := GetOSDistro()
	switch osDistro {
	case "windows":
		log.Printf("Installing CA Cert for %v (requires admin privileges)...", osDistro)

		if !IsAdmin() {
			return errors.New("please run the installation in an administrative command prompt")
		}

		cmd := exec.Command("certutil", "-addstore", "-f", "Root", caCertPemFilePath)
		out, err := cmd.CombinedOutput()
		log.Debugf("output: %v", string(out))

		if err != nil {
			return err
		}

		log.Printf("CA Certificates updated.")

		return nil
	case "darwin":
		log.Printf("Installing CA Cert for %v (requires sudo privileges)...", osDistro)

		cmd := exec.Command(
			"sudo", "security", "add-trusted-cert", "-d", "-r",
			"trustRoot", "-k", "/Library/Keychains/System.keychain", caCertPemFilePath,
		)
		log.Printf("Running command: %v", cmd)

		out, err := cmd.CombinedOutput()
		log.Debugf("output: %v", string(out))

		if err != nil {
			return err
		}

		log.Printf("Updated CA Certificates %v", string(out))

		return nil
	case "ubuntu", "debian", "pop":
		destinationCaCertPemFilePath := fmt.Sprintf("/usr/local/share/ca-certificates/%v-local-ca.cert.pem", AppName)

		log.Printf("Installing CA Cert for %v (requires sudo privileges)...", osDistro)
		log.Debugf("path: %v", destinationCaCertPemFilePath)

		cmdCp := fmt.Sprintf("sudo cp -va %v %v", caCertPemFilePath, destinationCaCertPemFilePath)
		cmd := exec.Command("/bin/sh", "-c", cmdCp)
		log.Printf("Running command: %v", cmd)

		out, err := cmd.CombinedOutput()
		log.Debugf("output: %v", string(out))

		if err != nil {
			return err
		}

		cmdInstall := "sudo update-ca-certificates"
		cmd = exec.Command("/bin/sh", "-c", cmdInstall)
		log.Printf("Running command: %v", cmd)

		out, err = cmd.CombinedOutput()
		log.Debugf("output: %v", string(out))

		if err != nil {
			return err
		}

		log.Printf("CA Certificates updated.")

		return nil
	case "fedora", "centos":
		destinationCaCertPemFilePath := fmt.Sprintf("/etc/pki/ca-trust/source/anchors/%v-local-ca.cert.pem", AppName)

		log.Printf("Installing CA cert for %v (requires sudo privileges)...", osDistro)
		log.Debugf("%v", destinationCaCertPemFilePath)

		cmdCp := fmt.Sprintf("sudo cp -va %v %v", caCertPemFilePath, destinationCaCertPemFilePath)
		cmd := exec.Command("/bin/sh", "-c", cmdCp)
		log.Debugf("Running command: %v", cmd)

		out, err := cmd.CombinedOutput()
		log.Debugf("output: %v", string(out))

		if err != nil {
			return err
		}

		cmdInstall := "sudo update-ca-trust"
		cmd = exec.Command("/bin/sh", "-c", cmdInstall)
		log.Debugf("Running command: %v", cmd)

		out, err = cmd.CombinedOutput()
		log.Debugf("output: %v", string(out))

		if err != nil {
			return err
		}

		log.Printf("CA Certificates updated.")

		return nil
	default:
		return errors.New("your operating system is not supported. Yet. :(")
	}
}

// CreatePrivKeyAndCertificate creates a Private Key and a Certificate signed by caCertificate
// and writes to file in PEM format.
func CreatePrivKeyAndCertificate(
	certDir string, certName string,
	dnsNames []string, caCertFilePath, caPrivKeyFilePath string,
) error {
	log.Debugln()

	// Reading CA Cert
	r, _ := ioutil.ReadFile(caCertFilePath)
	block, _ := pem.Decode(r)
	log.Debugln(caCertFilePath, "filetype:", block.Type)

	caCert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	// Reading CA Priv Key
	r, _ = ioutil.ReadFile(caPrivKeyFilePath)
	block, _ = pem.Decode(r)
	log.Debugln(caPrivKeyFilePath, "filetype:", block.Type)

	caPrivKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	privKeyFileName := certName + ".key.pem"
	privKeyFilePath := filepath.Join(certDir, privKeyFileName)
	log.Debugln("stripped filename:", privKeyFilePath)
	privKey, err := createPrivKeyAndWriteToPemFile(2048, privKeyFilePath)

	if err != nil {
		return err
	}

	subject := pkix.Name{
		CommonName:   certName,
		Organization: []string{AppName},
		Country:      []string{"HU"},
	}

	certFileName := certName + ".crt.pem"
	certFilePath := filepath.Join(certDir, certFileName)

	components := certificateComponents{
		subject:   subject,
		dnsNames:  dnsNames,
		privKey:   privKey,
		caCert:    caCert,
		caPrivKey: caPrivKey,
	}

	signedCert, err := CreateSignedCertificate(components)

	if err != nil {
		return fmt.Errorf("%w", err)
	}

	err = certificateWriteToPemFile(signedCert, certFilePath)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}

func createPrivKeyAndWriteToPemFile(bits int, privKeyPemFilePath string) (*rsa.PrivateKey, error) {
	log.Debugln()

	privKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	privKeyPEM := new(bytes.Buffer)
	err = pem.Encode(
		privKeyPEM, &pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(privKey),
		},
	)

	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	err = CreateDirAndWriteBytesToFile(privKeyPEM.Bytes(), privKeyPemFilePath)
	if err != nil {
		return nil, err
	}

	return privKey, nil
}

func CreateSignedCertificate(c certificateComponents) ([]byte, error) {
	log.Debugln()

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)

	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	cert := &x509.Certificate{
		SerialNumber:          serialNumber,
		Subject:               c.subject,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(1, 0, 0),
		IsCA:                  false,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		BasicConstraintsValid: true,
	}
	cert.DNSNames = c.dnsNames

	// The currently supported key types are *rsa.PublicKey, *ecdsa.PublicKey and ed25519.PublicKey.
	// Pub must be a supported key type, and priv must be a crypto.Signer with a supported public key.
	signedCert, err := x509.CreateCertificate(rand.Reader, cert, c.caCert, &c.privKey.PublicKey, c.caPrivKey)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	return signedCert, nil
}

func certificateWriteToPemFile(cert []byte, certPemFilePath string) error {
	log.Debugln()

	certPem := new(bytes.Buffer)
	err := pem.Encode(
		certPem, &pem.Block{
			Type:  "CERTIFICATE",
			Bytes: cert,
		},
	)

	if err != nil {
		return fmt.Errorf("%w", err)
	}

	err = CreateDirAndWriteBytesToFile(certPem.Bytes(), certPemFilePath)
	if err != nil {
		return err
	}

	return nil
}
