package logic

import (
	"path/filepath"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/rewardenv/reward/internal/config"
	cryptopkg "github.com/rewardenv/reward/internal/crypto"
	"github.com/rewardenv/reward/pkg/util"
)

// RunCmdSignCertificate represents the sign-certificate command.
func (c *Client) RunCmdSignCertificate(args []string, force ...bool) error {
	if len(args) == 0 {
		return config.ErrHostnameRequired
	}

	crypto := cryptopkg.New(c.Config)

	caCertExist := crypto.CheckCACertificateExistInDir(c.SSLCADir(), true)
	if !caCertExist {
		return config.ErrCaCertDoesNotExist
	}

	err := util.CreateDir(c.SSLCertDir(), nil)
	if err != nil {
		return errors.Wrap(err, "creating ssl cert directory")
	}

	certName := args[0]
	certificateFileName := certName + ".crt.pem"
	log.Debugln("certificate filename:", certificateFileName)
	certificateFilePath := filepath.Join(c.SSLCertDir(), certificateFileName)

	certificateExist := false
	if len(force) > 0 && force[0] {
		certificateExist = util.FileExists(certificateFilePath)
	} else {
		certificateExist = util.CheckFileExistsAndRecreate(certificateFilePath)
	}

	dnsNames := make([]string, 0, len(args))
	for _, dns := range args {
		dnsNames = append(dnsNames, dns, "*."+dns)
	}

	if !certificateExist {
		caCertFilePath, err := crypto.CACertificateFilePath(c.SSLCADir())
		if err != nil {
			return errors.Wrap(err, "getting ca cert file path")
		}

		caPrivKeyFilePath, err := crypto.CAPrivKeyFilePath(c.SSLCADir())
		if err != nil {
			return errors.Wrap(err, "getting ca priv key file path")
		}

		err = crypto.CreatePrivateKeyAndCertificate(c.SSLCertDir(),
			certName,
			dnsNames,
			caCertFilePath,
			caPrivKeyFilePath)
		if err != nil {
			return errors.Wrap(err, "creating private key and certificate")
		}
	}

	err = c.RunCmdSvc([]string{"up", "traefik"})
	if err != nil {
		return errors.Wrap(err, "bringing up traefik")
	}

	err = c.RunCmdSvc([]string{"restart", "traefik"})
	if err != nil {
		return errors.Wrap(err, "restarting traefik")
	}

	return nil
}
