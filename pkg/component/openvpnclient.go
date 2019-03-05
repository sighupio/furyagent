package component

import (
	"crypto/x509"
	"log"
	"net"
	"os"
	"path/filepath"

	certutil "k8s.io/client-go/util/cert"
	pki "k8s.io/kubernetes/cmd/kubeadm/app/util/pkiutil"
)

const (
	OpenVPNClientCert   = "client.crt"
	OpenVPNClientKey    = "client.key"
	OpenVPNClientCaCert = "ca.crt"
	OpenVPNClientCaKey  = "ca.key"
	OpenVPNClinetTaKey  = "ta.key"
)

type OpenVPNClient struct {
	ClusterComponentData
}

func (o OpenVPNClient) Backup() error {
	return nil
}

func (o OpenVPNClient) Restore() error {
	return nil
}

func (o OpenVPNClient) getFileMappings() [][]string {
	return [][]string{
		[]string{OpenVPNClientCaCert, OpenVPNClientCaCert},
		[]string{OpenVPNClientCaKey, OpenVPNClientCaKey},
		[]string{OpenVPNClinetTaKey, OpenVPNClinetTaKey},
	}
}

func (o OpenVPNClient) Configure(overwrite bool) error {
	files := o.getFileMappings()
	for _, v := range o.OpenVPNClient.Users {
		path := filepath.Join(o.OpenVPNClient.TargetDir, v)

		log.Println("Creating directory for: ", v)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			e := os.Mkdir(path, 0755)
			if e != nil {
				return e
			}
		}
		log.Println("Downloading CA.crt and TA.key for: ", v)
		err := o.DownloadFilesToDirectory(files, path, OpenVPNPath, overwrite)
		if err != nil {
			return err
		}
		defer os.Remove(filepath.Join(path, OpenVPNClientCaKey)) // clean up ca.key
		log.Println("Creating client certs for: ", v)
		err = o.createClientCertificates(v, path)
		if err != nil {
			return err
		}
	}
	return nil
}

func (o OpenVPNClient) Init(dir string) error {
	return nil
}

func (o OpenVPNClient) createClientCertificates(username, path string) error {
	caCert, caKey, err := pki.TryLoadCertAndKeyFromDisk(path, "ca")
	if err != nil {
		return err
	}
	clientCertConfig := certutil.Config{
		CommonName:   username,
		Organization: []string{"SIGHUP s.r.l."},
		AltNames:     certutil.AltNames{DNSNames: []string{}, IPs: []net.IP{}},
		Usages:       []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}

	clientCert, clientKey, err := pki.NewCertAndKey(caCert, caKey, &clientCertConfig)
	if err != nil {
		return err
	}
	err = pki.WriteCertAndKey(path, "client", clientCert, clientKey)
	if err != nil {
		return err
	}
	return nil
}
