package component

import (
	"bytes"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"

	certutil "k8s.io/client-go/util/cert"
	pki "k8s.io/kubernetes/cmd/kubeadm/app/util/pkiutil"
)

const (
	OpenVPNClientCert   = "client.crt"
	OpenVPNClientKey    = "client.key"
	OpenVPNClientCaCert = "ca.crt"
	OpenVPNClientCaKey  = "ca.key"
	OpenVPNClientTaKey  = "ta.key"
	OpenVPNClientPath   = "pki/vpn-client"
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
		[]string{OpenVPNClientTaKey, OpenVPNClientTaKey},
	}
}

func (o OpenVPNClient) Configure(overwrite, revoke bool) error {
	files := o.getFileMappings()

	if len(o.OpenVPNClient.Users) == 0 {
		log.Fatalf("No users defined in furyagent config file passed (clusterComponent.openvpn-client.users)")
	}

	for _, v := range o.OpenVPNClient.Users {
		if revoke {
			files := []string{
				OpenVPNCaCert,
				OpenVPNCaKey,
				OpenVPNCRL,
			}
			log.Println("Downloading ca.crt, ca.key, ca.crl")
			ca, err := o.DownloadFilesToMemory(files, OpenVPNPath)
			if err != nil {
				return err
			}
			log.Println("Downloading ", v, ".crt")
			cert, err := o.DownloadFilesToMemory([]string{v + ".crt"}, OpenVPNClientPath)
			if err != nil {
				return err
			}
			log.Println("Revoking certificate for: ", v)
			err = o.revokeClientCertificate(ca[OpenVPNCaCert], ca[OpenVPNCaKey], ca[OpenVPNCRL], cert[v+".crt"])
			if err != nil {
				return err
			}
			log.Println("Removing certificate for: ", v)
			if err := o.Remove(filepath.Join(OpenVPNClientPath, v+".crt")); err != nil {
				return err
			}
		} else {
			path := filepath.Join(o.OpenVPNClient.TargetDir, v)
			log.Println("Creating directory for: ", v)
			if _, err := os.Stat(path); os.IsNotExist(err) {
				e := os.MkdirAll(path, 0755)
				if e != nil {
					return e
				}
			}
			log.Println("Downloading CA.crt and TA.key for: ", v)
			if err := o.DownloadFilesToDirectory(files, path, OpenVPNPath, overwrite); err != nil {
				return err
			}
			defer os.Remove(filepath.Join(path, OpenVPNClientCaKey)) // clean up ca.key
			log.Println("Creating client certs for: ", v)
			if err := o.createClientCertificates(v, path); err != nil {
				return err
			}
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
	files := map[string][]byte{
		username + ".crt": certutil.EncodeCertPEM(clientCert),
	}
	if err = o.UploadFilesFromMemory(files, OpenVPNClientPath); err != nil {
		return err
	}
	return nil
}

func (o OpenVPNClient) revokeClientCertificate(catPEMBytes, keyPEMBytes, crlPEMBytes, certPEMBytes []byte) error {
	now := time.Now()

	caPEMBlock, _ := pem.Decode(catPEMBytes)
	ca, err := x509.ParseCertificate(caPEMBlock.Bytes)
	if err != nil {
		return err
	}

	keyPEMBlock, _ := pem.Decode(keyPEMBytes)
	key, err := x509.ParsePKCS1PrivateKey(keyPEMBlock.Bytes)
	if err != nil {
		return err
	}

	certPEMBlock, _ := pem.Decode(certPEMBytes)
	cert, err := x509.ParseCertificate(certPEMBlock.Bytes)
	if err != nil {
		return err
	}

	certRevocation := pkix.RevokedCertificate{
		SerialNumber:   cert.SerialNumber,
		RevocationTime: now,
	}

	crl, err := x509.ParseCRL(crlPEMBytes)
	if err != nil {
		return err
	}

	revokedCerts := append(crl.TBSCertList.RevokedCertificates, certRevocation)
	newCRL, err := ca.CreateCRL(rand.Reader, key, revokedCerts, now.UTC(), now.AddDate(10, 0, 0).UTC())
	if err != nil {
		return err
	}

	newCRLPEMBlock := &pem.Block{
		Type:  "X509 CRL",
		Bytes: newCRL,
	}
	CRLBuffer := new(bytes.Buffer)
	if err = pem.Encode(CRLBuffer, newCRLPEMBlock); err != nil {
		return err
	}
	if err = o.Remove(filepath.Join(OpenVPNPath, OpenVPNCRL)); err != nil {
		return err
	}
	files := map[string][]byte{
		OpenVPNCRL: CRLBuffer.Bytes(),
	}
	if err = o.UploadFilesFromMemory(files, OpenVPNPath); err != nil {
		return err
	}

	return nil
}
