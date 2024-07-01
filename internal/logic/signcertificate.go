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

	if err := util.CreateDir(c.SSLCertDir(), nil); err != nil {
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

		if err := crypto.CreatePrivateKeyAndCertificate(c.SSLCertDir(),
			certName,
			dnsNames,
			caCertFilePath,
			caPrivKeyFilePath); err != nil {
			return errors.Wrap(err, "creating private key and certificate")
		}
	}

	if err := c.RunCmdSvc([]string{"up", "traefik"}); err != nil {
		return errors.Wrap(err, "bringing up traefik")
	}

	if err := c.RunCmdSvc([]string{"restart", "traefik"}); err != nil {
		return errors.Wrap(err, "restarting traefik")
	}

	return nil
}
