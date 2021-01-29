package internal

import (
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

func SignCertificateCmd(args []string, dontAskRecreate ...bool) error {
	if len(args) == 0 {
		return ArgumentRequiredError("hostname")
	}

	sslDir := filepath.Join(GetAppHomeDir(), sslBaseDir)
	certDir := filepath.Join(sslDir, certBaseDir)
	caDir := filepath.Join(sslDir, caBaseDir)

	caCertExist := CheckCaCertificateExistInDir(caDir, true)
	if !caCertExist {
		return ErrCaCertDoesNotExist
	}

	err := CreateDir(certDir)
	if err != nil {
		return err
	}

	certName := args[0]
	certificateFileName := certName + ".crt.pem"
	log.Debugln("certificate filename:", certificateFileName)
	certificateFilePath := filepath.Join(certDir, certificateFileName)

	certificateExist := false
	if len(dontAskRecreate) > 0 && dontAskRecreate[0] {
		certificateExist = CheckFileExists(certificateFilePath)
	} else {
		certificateExist = CheckFileExistsAndRecreate(certificateFilePath)
	}

	var dnsNames []string
	for _, dns := range args {
		dnsNames = append(dnsNames, dns, "*."+dns)
	}

	if !certificateExist {
		caCertFilePath, err := GetCaCertificateFilePath(caDir)
		if err != nil {
			return err
		}

		caPrivKeyFilePath, err := GetCaPrivKeyFilePath(caDir)
		if err != nil {
			return err
		}

		err = CreatePrivKeyAndCertificate(certDir, certName, dnsNames, caCertFilePath, caPrivKeyFilePath)
		if err != nil {
			return err
		}
	}

	err = SvcCmd([]string{"up", "traefik"})
	if err != nil {
		return err
	}

	err = SvcCmd([]string{"restart", "traefik"})
	if err != nil {
		return err
	}

	return nil
}
