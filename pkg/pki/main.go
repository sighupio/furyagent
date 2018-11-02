// Copyright © 2018 Sighup SRL support@sighup.io
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
