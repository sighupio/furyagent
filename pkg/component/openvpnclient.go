package component

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"text/template"
	"time"

	"github.com/olekukonko/tablewriter"
	certutil "k8s.io/client-go/util/cert"
	"k8s.io/kubernetes/cmd/kubeadm/app/util/pkiutil"
)

const (
	OpenVPNClientCert        = "client.crt"
	OpenVPNClientKey         = "client.key"
	OpenVPNClientCaCert      = "ca.crt"
	OpenVPNClientCaKey       = "ca.key"
	OpenVPNClientTaKey       = "ta.key"
	OpenVPNClientPath        = "pki/vpn-client"
	OpenVPNClientRevokedPath = "pki/vpn-client/revoked"

	openVPNClientConfigTmpl = `
client
dev tun
proto udp
remote-random
remote-cert-tls server
tls-version-min 1.2
tls-cipher TLS-ECDHE-RSA-WITH-AES-128-GCM-SHA256:TLS-ECDHE-ECDSA-WITH-AES-128-GCM-SHA256:TLS-ECDHE-RSA-WITH-AES-256-GCM-SHA384:TLS-DHE-RSA-WITH-AES-256-CBC-SHA256
cipher AES-256-CBC
auth SHA512
resolv-retry infinite
auth-retry none
nobind
key-direction 1
persist-key
persist-tun
compress lzo
verb 3
tls-client

{{ range $server := .Server }}remote {{ $server }} 1194
{{ end }}

<ca>
{{ .CACert }}</ca>

<cert>
{{ .ClientCert }}</cert>

<key>
{{ .ClientKey }}</key>

<tls-auth>
{{ .TLSAuthKey }}</tls-auth>
`
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

type RevocationResponse struct {
	Revoked    bool
	RevokeTime time.Time
}

type ListOutput struct {
	User       string
	Valid_from string
	Valid_to   string
	Expired    bool
	Revoked    RevocationResponse
}

func (o OpenVPNClient) ListUserCertificates(output string) error {
	s3files, err := o.List(OpenVPNClientPath)
	if err != nil {
		return err
	}
	data := [][]string{}
	files, err := o.DownloadFilesToMemory(s3files, OpenVPNClientPath)
	var jsonOutput ListOutput

	for _, file := range files {
		cpb, _ := pem.Decode(file)
		crt, err := x509.ParseCertificate(cpb.Bytes)
		if err != nil {
			return err
		}
		name := crt.Subject.CommonName
		vt := crt.NotAfter.Format("2006-01-02")
		vf := crt.NotBefore.Format("2006-01-02")

		now := time.Now()
		var expired bool
		if now.After(crt.NotAfter) {
			expired = true
		}

		filenames := []string{
			OpenVPNCRL,
		}
		ca, err := o.DownloadFilesToMemory(filenames, OpenVPNPath)
		if err != nil {
			return err
		}
		// Parse ceritifcate revocation list
		crl, err := x509.ParseCRL(ca[OpenVPNCRL])

		if err != nil {
			return err
		}
		revoke := getRevocationInfo(crt, crl.TBSCertList.RevokedCertificates)

		data = append(data, []string{name, fmt.Sprintln(vf), fmt.Sprintln(vt), fmt.Sprintf("%v", expired), fmt.Sprintf("%v %v", revoke.Revoked, revoke.RevokeTime)})

		jsonOutput = ListOutput{
			User:       name,
			Valid_from: vf,
			Valid_to:   vt,
			Expired:    expired,
			Revoked:    revoke,
		}
	}

	switch output {
	case "json":
		resp, _ := json.Marshal(jsonOutput)
		fmt.Println(string(resp))
	default:
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"User", "Valid from", "Valid to", "Expired", "Revoked"})
		table.SetRowLine(true)
		for _, v := range data {
			table.Append(v)
		}
		table.Render()
	}
	return nil
}

func getRevocationInfo(cert *x509.Certificate, revokedList []pkix.RevokedCertificate) RevocationResponse {
	for _, rc := range revokedList {
		if rc.SerialNumber.String() == cert.SerialNumber.String() {
			return RevocationResponse{
				Revoked:    true,
				RevokeTime: rc.RevocationTime,
			}
		}
	}
	return RevocationResponse{}
}

func (o OpenVPNClient) CreateUser(clientName string) error {
	if o.Exists(filepath.Join(OpenVPNClientPath, clientName+".crt")) {
		return errors.New(fmt.Sprintf("client certificate for %s already exists", clientName))
	}
	filenames := []string{
		OpenVPNCaCert,
		OpenVPNCaKey,
		OpenVPNTaKey,
	}
	log.Println("Downloading ca.crt, ca.key and ta.key")
	files, err := o.DownloadFilesToMemory(filenames, OpenVPNPath)
	if err != nil {
		return err
	}
	log.Println("Creating client cert for: ", clientName)
	clientCert, err := o.createClientCertificate(clientName, files[OpenVPNCaCert], files[OpenVPNCaKey])
	if err != nil {
		return err
	}
	type openVPNClientConfig struct {
		Server     []string
		CACert     string
		ClientCert string
		ClientKey  string
		TLSAuthKey string
	}
	clientConfig := openVPNClientConfig{
		Server:     o.OpenVPN.Servers,
		CACert:     string(files[OpenVPNCaCert]),
		ClientCert: string(clientCert[clientName+".crt"]),
		ClientKey:  string(clientCert[clientName+".key"]),
		TLSAuthKey: string(files[OpenVPNTaKey]),
	}
	delete(clientCert, clientName+".key")
	log.Println("Uploading client cert for: ", clientName)
	if err = o.UploadFilesFromMemory(clientCert, OpenVPNClientPath); err != nil {
		return err
	}
	t := template.Must(template.New("openVPNClientConfig").Parse(openVPNClientConfigTmpl))
	err = t.Execute(os.Stdout, clientConfig)
	if err != nil {
		return err
	}
	return nil
}

func (o OpenVPNClient) createClientCertificate(clientName string, ca, key []byte) (map[string][]byte, error) {
	certs, err := certutil.ParseCertsPEM(ca)
	if err != nil {
		return nil, err
	}
	caCert := certs[0]
	caKey, err := certutil.ParsePrivateKeyPEM(key)
	clientCertConfig := certutil.Config{
		CommonName:   clientName,
		Organization: []string{"SIGHUP s.r.l."},
		AltNames:     certutil.AltNames{DNSNames: []string{}, IPs: []net.IP{}},
		Usages:       []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}
	clientCert, clientKey, err := pkiutil.NewCertAndKey(caCert, caKey.(*rsa.PrivateKey), &clientCertConfig)
	if err != nil {
		return nil, err
	}
	files := map[string][]byte{
		clientName + ".crt": certutil.EncodeCertPEM(clientCert),
		clientName + ".key": certutil.EncodePrivateKeyPEM(clientKey),
	}
	return files, nil
}

func (o OpenVPNClient) RevokeUser(clientName string) error {
	clientCert := clientName + ".crt"
	filenames := []string{
		OpenVPNCaCert,
		OpenVPNCaKey,
		OpenVPNCRL,
	}
	log.Println("Downloading ca.crt, ca.key, ca.crl")
	ca, err := o.DownloadFilesToMemory(filenames, OpenVPNPath)
	if err != nil {
		return err
	}
	log.Println("Downloading ", clientCert)
	cert, err := o.DownloadFilesToMemory([]string{clientCert}, OpenVPNClientPath)
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
	log.Println("Moving certificate to revoked folder for: ", clientName)
	if err := o.Move(clientName+".crt", OpenVPNClientPath, OpenVPNClientRevokedPath); err != nil {
		return err
	}
	return nil
}

func (o OpenVPNClient) revokeClientCertificate(caPEMBytes, keyPEMBytes, crlPEMBytes, certPEMBytes []byte) (map[string][]byte, error) {
	now := time.Now()
	// Parse ca certificate
	caPEMBlock, _ := pem.Decode(caPEMBytes)
	ca, err := x509.ParseCertificate(caPEMBlock.Bytes)
	if err != nil {
		return nil, err
	}
	// Parse ca private key
	keyPEMBlock, _ := pem.Decode(keyPEMBytes)
	key, err := x509.ParsePKCS1PrivateKey(keyPEMBlock.Bytes)
	if err != nil {
		return nil, err
	}
	// Parse client certifiate
	certPEMBlock, _ := pem.Decode(certPEMBytes)
	cert, err := x509.ParseCertificate(certPEMBlock.Bytes)
	if err != nil {
		return nil, err
	}
	// Create certificate revocation
	certRevocation := pkix.RevokedCertificate{
		SerialNumber:   cert.SerialNumber,
		RevocationTime: now,
	}
	// Parse ceritifcate revocation list
	crl, err := x509.ParseCRL(crlPEMBytes)
	if err != nil {
		return nil, err
	}
	// Create updated CRL
	revokedCerts := append(crl.TBSCertList.RevokedCertificates, certRevocation)
	newCRL, err := ca.CreateCRL(rand.Reader, key, revokedCerts, now.UTC(), now.AddDate(10, 0, 0).UTC())
	if err != nil {
		return nil, err
	}
	// Marshal updated CRL
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
