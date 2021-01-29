package internal_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"math/big"
	"path/filepath"
	"reflect"
	"testing"

	. "reward/internal"

	"github.com/spf13/afero"
	"golang.org/x/crypto/ssh"
)

var (
	existingKeyFile                     = "/testdata/an-existing-key-file"
	testPrivKey, testInvalidPrivKey     *rsa.PrivateKey
	testPrivatePEM, testSSHPubKeyPem    []byte
	testRSAPubKey, testInvalidRSAPubKey *rsa.PublicKey
	testSSHKeyFilePath                  = filepath.Join(existingDir, "tunnel", "ssh_key")
)

func initCryptoTests() {
	testPrivKey, _ = rsa.GenerateKey(rand.Reader, 2048)
	testRSAPubKey = &testPrivKey.PublicKey
	testInvalidRSAPubKey = &rsa.PublicKey{
		N: big.NewInt(123),
		E: 0,
	}
	testInvalidPrivKey = &rsa.PrivateKey{
		PublicKey: *testInvalidRSAPubKey,
		D:         nil,
		Primes:    nil,
		Precomputed: rsa.PrecomputedValues{
			Dp:        nil,
			Dq:        nil,
			Qinv:      nil,
			CRTValues: nil,
		},
	}
	testPrivatePEM = pem.EncodeToMemory(&pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   x509.MarshalPKCS1PrivateKey(testPrivKey),
	})

	testSSHPubKey, _ := ssh.NewPublicKey(testRSAPubKey)
	testSSHPubKeyPem = ssh.MarshalAuthorizedKey(testSSHPubKey)

	_ = afero.WriteFile(AFS, existingKeyFile, testPrivatePEM, 0o644)
}

func TestGenerateRSAPrivateKey(t *testing.T) {
	t.Parallel()

	type args struct {
		bitSize int
	}

	tests := []struct {
		name    string
		args    args
		want    *rsa.PrivateKey
		wantErr bool
	}{
		{
			"Return a 2048 bit RSA Private Key",
			args{2048},
			testPrivKey,
			false,
		},
		{
			"Return an error for 1 bit RSA Private Key",
			args{1},
			nil,
			true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := GenerateRSAPrivateKey(tt.args.bitSize)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateRSAPrivateKey() error = %v, wantErr %v", err, tt.wantErr)

				return
			}
			if reflect.TypeOf(got) != reflect.TypeOf(tt.want) {
				t.Errorf("GenerateRSAPrivateKey() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEncodeRSAPrivateKeyToPEM(t *testing.T) {
	t.Parallel()

	type args struct {
		privateKey *rsa.PrivateKey
	}

	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			"Return PEM bytes for key",
			args{testPrivKey},
			testPrivatePEM,
			false,
		},
		{
			"Return error if the key is invalid",
			args{testInvalidPrivKey},
			testPrivatePEM,
			true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := EncodeRSAPrivateKeyToPEM(tt.args.privateKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("EncodeRSAPrivateKeyToPEM() error = %v, wantErr %v", err, tt.wantErr)

				return
			}
			if reflect.TypeOf(got) == reflect.TypeOf(&tt.want) {
				t.Errorf("EncodeRSAPrivateKeyToPEM() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerateSshPublicKey(t *testing.T) {
	t.Parallel()

	type args struct {
		publicKey *rsa.PublicKey
	}

	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			"Return PEM bytes for key",
			args{testRSAPubKey},
			testSSHPubKeyPem,
			false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := GenerateSSHPublicKey(tt.args.publicKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateSSHPublicKey() error = %v, wantErr %v", err, tt.wantErr)

				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GenerateSSHPublicKey() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerateAndSaveSshKeys(t *testing.T) {
	t.Parallel()

	type args struct {
		bitSize    int
		pathToSave string
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"Write files to non existing path",
			args{2048, testSSHKeyFilePath},
			false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := GenerateAndSaveSSHKeys(tt.args.bitSize, tt.args.pathToSave); (err != nil) != tt.wantErr {
				t.Errorf("GenerateAndSaveSSHKeys() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
