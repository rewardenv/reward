package commands

import (
	"path/filepath"

	"github.com/rewardenv/reward/internal/core"

	log "github.com/sirupsen/logrus"
)

// SignCertificateCmd represents the sign-certificate command.
func SignCertificateCmd(args []string, dontAskRecreate ...bool) error {
	if len(args) == 0 {
		return core.ArgumentRequiredError("hostname")
	}

	sslDir := filepath.Join(core.GetAppHomeDir(), core.SslBaseDir)
	certDir := filepath.Join(sslDir, core.CertBaseDir)
	caDir := filepath.Join(sslDir, core.CaBaseDir)

	caCertExist := core.CheckCaCertificateExistInDir(caDir, true)
	if !caCertExist {
		return core.ErrCaCertDoesNotExist
	}

	err := core.CreateDir(certDir)
	if err != nil {
		return err
	}

	certName := args[0]
	certificateFileName := certName + ".crt.pem"
	log.Debugln("certificate filename:", certificateFileName)
	certificateFilePath := filepath.Join(certDir, certificateFileName)

	certificateExist := false
	if len(dontAskRecreate) > 0 && dontAskRecreate[0] {
		certificateExist = core.CheckFileExists(certificateFilePath)
	} else {
		certificateExist = core.CheckFileExistsAndRecreate(certificateFilePath)
	}

	var dnsNames []string
	for _, dns := range args {
		dnsNames = append(dnsNames, dns, "*."+dns)
	}

	if !certificateExist {
		caCertFilePath, err := core.GetCaCertificateFilePath(caDir)
		if err != nil {
			return err
		}

		caPrivKeyFilePath, err := core.GetCaPrivKeyFilePath(caDir)
		if err != nil {
			return err
		}

		err = core.CreatePrivKeyAndCertificate(certDir, certName, dnsNames, caCertFilePath, caPrivKeyFilePath)
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
