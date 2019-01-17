package component

import (
	"crypto/x509"
	certutil "k8s.io/client-go/util/cert"
	pki "k8s.io/kubernetes/cmd/kubeadm/app/util/pkiutil"
	"log"
	"net"
)

const (
	ServerCert = "server.crt"
	ServerKey  = "server.key"
	CaCert     = "ca.crt"
)

var (
	config = certutil.Config{
		CommonName:   "SIGHUP s.r.l. OpenVPN Server",
		Organization: []string{"SIGHUP s.r.l."},
		AltNames:     certutil.AltNames{DNSNames: []string{}, IPs: []net.IP{}},
		Usages:       []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
)

type OpenVPN struct {
	ClusterComponentData
}

func (o OpenVPN) Backup() error {
	return nil
}

// Restore implements
func (o OpenVPN) Restore() error {
	return nil
}

func (o OpenVPN) getFileMappings() [][]string {
	return [][]string{
		[]string{ServerCert, ServerCert},
		[]string{ServerKey, ServerKey},
		[]string{CaCert, CaCert},
	}
}

func (o OpenVPN) Configure(overwrite bool) error {
	files := o.getFileMappings()
	bucketDir := "pki/openvpn"
	return o.DownloadFilesToDirectory(files, o.OpenVPN.CertDir, bucketDir, overwrite)
}

func (o OpenVPN) Init(dir string) error {
	o.createServerCerts(dir)
	files := o.getFileMappings()
	bucketDir := "pki/openvpn"
	return o.UploadFilesFromDirectory(files, dir, bucketDir)
}

func (o OpenVPN) createServerCerts(dir string) error {

	ca, privateKey, err := pki.NewCertificateAuthority(&config)
	if err != nil {
		log.Fatal(err)
	}

	serverCert, serverKey, err := pki.NewCertAndKey(ca, privateKey, &config)
	if err != nil {
		log.Fatal(err)
	}

	if err = pki.WriteCertAndKey(dir, "ca", ca, privateKey); err != nil {
		log.Fatal(err)
	}

	if err = pki.WriteCertAndKey(dir, "server", serverCert, serverKey); err != nil {
		log.Fatal(err)
	}

	return nil
}
