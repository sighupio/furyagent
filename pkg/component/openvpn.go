package component

import (
	certutil "k8s.io/client-go/util/cert"
	pki "k8s.io/kubernetes/cmd/kubeadm/app/util/pkiutil"
	"log"
)

const (
	ServerCert  = "server.crt"
	ServerKey   = "server.key"
	CaCert      = "ca.crt"
	CaKey       = "ca.key"
	OpenVPNPath = "pki/vpn"
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
		[]string{ServerKey, ServerKey},
		[]string{ServerKey, ServerKey},
		[]string{CaKey, CaKey},
		[]string{CaCert, CaCert},
	}
}

func (o OpenVPN) Configure(overwrite bool) error {
	files := o.getFileMappings()
	return o.DownloadFilesToDirectory(files, o.OpenVPN.CertDir, OpenVPNPath, overwrite)
}

func (o OpenVPN) Init(dir string) error {
	ca, privateKey, err := pki.NewCertificateAuthority(&CertConfig)
	if err != nil {
		log.Fatal(err)
	}

	serverCert, serverKey, err := pki.NewCertAndKey(ca, privateKey, &CertConfig)
	if err != nil {
		log.Fatal(err)
	}

	certs := map[string][]byte{
		CaCert:     certutil.EncodeCertPEM(ca),
		CaKey:      certutil.EncodePrivateKeyPEM(privateKey),
		ServerCert: certutil.EncodeCertPEM(serverCert),
		ServerKey:  certutil.EncodePrivateKeyPEM(serverKey),
	}
	if err = o.UploadFilesFromMemory(certs, OpenVPNPath); err != nil {
		log.Fatal(err)
	}
	return nil
}
