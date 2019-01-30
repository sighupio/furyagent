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

package component

import (
	"bytes"
	"k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/v1beta1"
	"log"
	"os/exec"
	"path/filepath"
)

// Node represent the object that reflects what nodes need (implements ClusterComponent)
type Node struct {
	ClusterComponentData
}

const (
	kubeletBootstrapConfig = "bootstrap-kubelet.conf"
)

// Backup of a node is Empty
func (n *Node) Backup() error {
	return nil
}

// Restore of a node is Empty
func (n *Node) Restore() error {
	return nil
}

// Configure basically joins the nodes to the cluster, configures KUBELET_EXTRA_ARGS and restart kubelet and docker in case of necessity
func (n *Node) Configure(overwrite bool) error {
	//download kubelet bootstrap file
	if err := n.DownloadFile(filepath.Join(ConfigurationRemoteDir, kubeletBootstrapConfig), n.Node.KubeletBootstrapConfig, overwrite); err != nil {
		log.Fatal(err)
	}

	return nil
}

func (n *Node) Init(dir string) error {
	initCmd := exec.Command("kubeadm", "init", "--config=", n.Master.KubeadmConfig)
	if err := initCmd.Run(); err != nil {
		log.Fatal(err)
	}

	//upload kubelet bootstrap configfile
	if err := UploadFilesFromMemory(map[string][]byte{
		//KubeletBootstrapConfig:
	}, ConfigurationRemoteDir); err != nil {
		log.Fatal(err)
	}

	// TOBEREMOVED
	tokenCmd := exec.Command("kubeadm", "token", "create", "--print-join-command", "--ttl=0")
	joinCommand := &bytes.Buffer{}
	tokenCmd.Stdout = joinCommand
	if err := tokenCmd.Run(); err != nil {
		log.Fatal(err)
	}

	if err := n.UploadFilesFromMemory(map[string][]byte{
		Token: joinCommand.Bytes(),
	}, NodePath); err != nil {
		log.Fatal(err)
	}
	// ENDTOBEREMOVED
	return nil
}
