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
	OpenVPNClientCert        = "client.crt"
	OpenVPNClientKey         = "client.key"
	OpenVPNClientCaCert      = "ca.crt"
	OpenVPNClientCaKey       = "ca.key"
	OpenVPNClientTaKey       = "ta.key"
	OpenVPNClientPath        = "pki/vpn-client"
	OpenVPNClientRevokedPath = "pki/vpn-client/revoked"
)

type OpenVPNClient struct {
	ClusterComponentData
}

func (o OpenVPNClient) getFileMappings() [][]string {
	return [][]string{
		[]string{OpenVPNClientCaCert, OpenVPNClientCaCert},
		[]string{OpenVPNClientCaKey, OpenVPNClientCaKey},
		[]string{OpenVPNClientTaKey, OpenVPNClientTaKey},
	}
}

func (o OpenVPNClient) CreateUser(clientName string) error {
	path := filepath.Join("./", clientName)
	log.Println("Creating directory for: ", clientName)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		e := os.MkdirAll(path, 0755)
		if e != nil {
			return e
		}
	}
	log.Println("Downloading ca.crt, ca.key and ta.key for: ", clientName)
	if err := o.DownloadFilesToDirectory(o.getFileMappings(), path, OpenVPNPath, false); err != nil {
		return err
	}
	defer os.Remove(filepath.Join(path, OpenVPNClientCaKey)) // clean up ca.key
	log.Println("Creating client cert for: ", clientName)
	clientCert, err := o.createClientCertificate(clientName, path)
	if err != nil {
		return err
	}
	log.Println("uploading client cert for: ", clientName)
	if err = o.UploadFilesFromMemory(clientCert, OpenVPNClientPath); err != nil {
		return err
	}
	return nil
}

func (o OpenVPNClient) RevokeUser(clientName string) error {
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
	log.Println("Downloading ", clientName, ".crt")
	cert, err := o.DownloadFilesToMemory([]string{clientName + ".crt"}, OpenVPNClientPath)
	if err != nil {
		return err
	}
	log.Println("Revoking certificate for: ", clientName)
	newCRL, err := o.revokeClientCertificate(ca[OpenVPNCaCert], ca[OpenVPNCaKey], ca[OpenVPNCRL], cert[clientName+".crt"])
	if err != nil {
		return err
	}
	log.Println("Removing old ca.crl")
	if err = o.Remove(filepath.Join(OpenVPNPath, OpenVPNCRL)); err != nil {
		return err
	}
	log.Println("Uploading new ca.crl")
	if err = o.UploadFilesFromMemory(newCRL, OpenVPNPath); err != nil {
		return err
	}
	log.Println("Moving certificate for: ", clientName)
	if err := o.Move(clientName+".crt", OpenVPNClientPath, OpenVPNClientRevokedPath); err != nil {
		return err
	}
	return nil
}

func (o OpenVPNClient) createClientCertificate(clientName, path string) (map[string][]byte, error) {
	caCert, caKey, err := pki.TryLoadCertAndKeyFromDisk(path, "ca")
	if err != nil {
		return nil, err
	}
	clientCertConfig := certutil.Config{
		CommonName:   clientName,
		Organization: []string{"SIGHUP s.r.l."},
		AltNames:     certutil.AltNames{DNSNames: []string{}, IPs: []net.IP{}},
		Usages:       []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}
	clientCert, clientKey, err := pki.NewCertAndKey(caCert, caKey, &clientCertConfig)
	if err != nil {
		return nil, err
	}
	err = pki.WriteCertAndKey(path, "client", clientCert, clientKey)
	if err != nil {
		return nil, err
	}
	files := map[string][]byte{
		clientName + ".crt": certutil.EncodeCertPEM(clientCert),
	}
	return files, nil
}

func (o OpenVPNClient) revokeClientCertificate(caPEMBytes, keyPEMBytes, crlPEMBytes, certPEMBytes []byte) (map[string][]byte, error) {
	now := time.Now()
	caPEMBlock, _ := pem.Decode(caPEMBytes)
	ca, err := x509.ParseCertificate(caPEMBlock.Bytes)
	if err != nil {
		return nil, err
	}
	keyPEMBlock, _ := pem.Decode(keyPEMBytes)
	key, err := x509.ParsePKCS1PrivateKey(keyPEMBlock.Bytes)
	if err != nil {
		return nil, err
	}
	certPEMBlock, _ := pem.Decode(certPEMBytes)
	cert, err := x509.ParseCertificate(certPEMBlock.Bytes)
	if err != nil {
		return nil, err
	}
	certRevocation := pkix.RevokedCertificate{
		SerialNumber:   cert.SerialNumber,
		RevocationTime: now,
	}
	crl, err := x509.ParseCRL(crlPEMBytes)
	if err != nil {
		return nil, err
	}
	revokedCerts := append(crl.TBSCertList.RevokedCertificates, certRevocation)
	newCRL, err := ca.CreateCRL(rand.Reader, key, revokedCerts, now.UTC(), now.AddDate(10, 0, 0).UTC())
	if err != nil {
		return nil, err
	}
	newCRLPEMBlock := &pem.Block{
		Type:  "X509 CRL",
		Bytes: newCRL,
	}
	newCRLBuffer := new(bytes.Buffer)
	if err = pem.Encode(newCRLBuffer, newCRLPEMBlock); err != nil {
		return nil, err
	}
	newCRLBytes := make([]byte, newCRLBuffer.Len())
	_, err = newCRLBuffer.Read(newCRLBytes)
	if err != nil {
		return nil, err
	}
	files := map[string][]byte{
		OpenVPNCRL: newCRLBytes,
	}
	return files, nil
}
