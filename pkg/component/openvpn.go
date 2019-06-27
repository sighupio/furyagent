package component

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"

	certutil "k8s.io/client-go/util/cert"
	pki "k8s.io/kubernetes/cmd/kubeadm/app/util/pkiutil"
)

const (
	OpenVPNServerCert = "server.crt"
	OpenVPNServerKey  = "server.key"
	OpenVPNCaCert     = "ca.crt"
	OpenVPNCaKey      = "ca.key"
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
		[]string{OpenVPNTaKey, OpenVPNTaKey},
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

	taKeyData, err := getTaKey()
	if err != nil {
		log.Fatal(err)
	}

	certs := map[string][]byte{
		OpenVPNCaCert:     certutil.EncodeCertPEM(ca),
		OpenVPNCaKey:      certutil.EncodePrivateKeyPEM(privateKey),
		OpenVPNServerCert: certutil.EncodeCertPEM(serverCert),
		OpenVPNServerKey:  certutil.EncodePrivateKeyPEM(serverKey),
		OpenVPNTaKey:      taKeyData,
	}
	if err = o.UploadFilesFromMemory(certs, OpenVPNPath, false); err != nil {
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

	// fmt.Println("Temporary file is: ", tmpfile.Name())

	// fmt.Println("With content: ", string(data))

	return data, nil
}
