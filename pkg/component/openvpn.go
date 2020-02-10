package component

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"io/ioutil"
	"log"
	"math/big"
	"os"
	"os/exec"
	"time"

	"k8s.io/client-go/util/keyutil"
	"k8s.io/kubernetes/cmd/kubeadm/app/util/pkiutil"
)

const (
	OpenVPNServerCert = "server.crt"
	OpenVPNServerKey  = "server.key"
	OpenVPNCaCert     = "ca.crt"
	OpenVPNCaKey      = "ca.key"
	OpenVPNCRL        = "ca.crl"
	OpenVPNTaKey      = "ta.key"
	OpenVPNPath       = "pki/vpn"
)

type OpenVPN struct {
	ClusterComponentData
}

func (o OpenVPN) Backup() error {
	return nil
}

func (o OpenVPN) Restore() error {
	return nil
}

func (o OpenVPN) getFileMappings() [][]string {
	return [][]string{
		[]string{OpenVPNServerKey, OpenVPNServerKey},
		[]string{OpenVPNServerCert, OpenVPNServerCert},
		[]string{OpenVPNCaKey, OpenVPNCaKey},
		[]string{OpenVPNCaCert, OpenVPNCaCert},
		[]string{OpenVPNCRL, OpenVPNCRL},
		[]string{OpenVPNTaKey, OpenVPNTaKey},
	}
}

func (o OpenVPN) Configure(overwrite bool) error {
	files := o.getFileMappings()
	return o.DownloadFilesToDirectory(files, o.OpenVPN.CertDir, OpenVPNPath, overwrite)
}

func (o OpenVPN) Init(dir string) error {
	now := time.Now()
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatal(err)
	}
	tmpl := x509.Certificate{
		SerialNumber: new(big.Int).SetInt64(0),
		Subject: pkix.Name{
			CommonName:   "openvpn",
			Organization: []string{"SIGHUP s.r.l."},
		},
		NotBefore:             now.UTC(),
		NotAfter:              now.Add(time.Hour * 24 * 365 * 10).UTC(),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}
	caDERBytes, err := x509.CreateCertificate(rand.Reader, &tmpl, &tmpl, privateKey.Public(), privateKey)
	if err != nil {
		log.Fatal(err)
	}
	ca, err := x509.ParseCertificate(caDERBytes)
	if err != nil {
		log.Fatal(err)
	}

	crl, err := ca.CreateCRL(rand.Reader, privateKey, []pkix.RevokedCertificate{}, now, now.AddDate(10, 0, 0).UTC())
	if err != nil {
		return err
	}
	crlPEMBlock := &pem.Block{
		Type:  "X509 CRL",
		Bytes: crl,
	}
	crlBuffer := new(bytes.Buffer)
	if err = pem.Encode(crlBuffer, crlPEMBlock); err != nil {
		return err
	}

	serverCert, serverKey, err := pkiutil.NewCertAndKey(ca, privateKey, &CertConfig)
	if err != nil {
		log.Fatal(err)
	}

	taKeyData, err := getTaKey()
	if err != nil {
		log.Fatal(err)
	}

	caKeyPEM, err := keyutil.MarshalPrivateKeyToPEM(privateKey)
	if err != nil {
		log.Fatal(err)
	}

	serverKeyPEM, err := keyutil.MarshalPrivateKeyToPEM(serverKey)
	if err != nil {
		log.Fatal(err)
	}

	certs := map[string][]byte{
		OpenVPNCaCert:     pkiutil.EncodeCertPEM(ca),
		OpenVPNCaKey:      caKeyPEM,
		OpenVPNServerCert: pkiutil.EncodeCertPEM(serverCert),
		OpenVPNServerKey:  serverKeyPEM,
		OpenVPNCRL:        crlBuffer.Bytes(),
		OpenVPNTaKey:      taKeyData,
	}
	if err = o.UploadFilesFromMemory(certs, OpenVPNPath); err != nil {
		log.Fatal(err)
	}
	return nil
}

func getTaKey() ([]byte, error) {
	tmpfile, err := ioutil.TempFile("", "openvpn")
	if err != nil {
		return nil, err
	}

	defer os.Remove(tmpfile.Name()) // clean up

	err = exec.Command("openvpn", "--genkey", "--secret", tmpfile.Name()).Run()
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadFile(tmpfile.Name())

	return data, nil
}
