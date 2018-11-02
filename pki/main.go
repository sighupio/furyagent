package pki

import (
	"log"

	"k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
	"k8s.io/kubernetes/cmd/kubeadm/app/phases/certs"
)

// NewPKI generates new
func NewPKI() {
	cfg := &kubeadm.InitConfiguration{
		APIEndpoint: kubeadm.APIEndpoint{
			AdvertiseAddress: "1.1.1.1",
		},
		NodeRegistration: kubeadm.NodeRegistrationOptions{
			Name: "Grande-puffo",
		},
		ClusterConfiguration: kubeadm.ClusterConfiguration{
			ControlPlaneEndpoint: "pippo-pluto",
			APIServerCertSANs:    []string{"bello-figo", "tanta-roba-bella"},
			CertificatesDir:      "./pki",
			ClusterName:          "super-mega-cluster",
			Networking: kubeadm.Networking{
				DNSDomain:     "cluster.local",
				PodSubnet:     "192.168.0.0/17",
				ServiceSubnet: "192.168.254.0/17",
			},
			KubernetesVersion: "v1.12.0",
			Etcd: kubeadm.Etcd{
				Local: &kubeadm.LocalEtcd{
					ServerCertSANs: []string{"super-cool-server"},
					PeerCertSANs:   []string{"very-cool-peer"},
				},
			},
		},
	}
	err := certs.CreatePKIAssets(cfg)
	if err != nil {
		log.Fatal(err)
	}
}
