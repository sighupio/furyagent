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
	"git.incubator.sh/sighup/furyagent/pkg/storage"
	"log"
	"os"
	"path/filepath"
)

// ClusterComponent interface represent the basic concept of the componet: etcd, master, node
type ClusterComponent interface {
	Backup(*ClusterConfig, *storage.Data) error
	Restore(*ClusterConfig, *storage.Data) error
	Configure(*ClusterConfig, *storage.Data) error
}

// ClusterConfig represents the configuration for the whole cluster
type ClusterConfig struct {
	NodeName string       `yaml:"nodeName"`
	Etcd     EtcdConfig   `yaml:"etcd"`
	Master   MasterConfig `yaml:"master"`
	Node     NodeConfig   `yaml:"node"`
}

// EtcdConfig is used to backup/restore/configure etcd nodes
type EtcdConfig struct {
	DataDir             string `yaml:"dataDir"`
	CertDir             string `yaml:"certDir"`
	CaCertFilename      string `yaml:"caCertFilename"`
	CaKeyFilename       string `yaml:"caKeyFilename"`
	ClientCertFilename  string `yaml:"clientCertFilename"`
	InitialClusterToken string `yaml:"initialClusterToken"`
	SnapshotFilename    string `yaml:"snapshotFilename"`
	ClientKeyFilename   string `yaml:"clientKeyFilename"`
	Endpoint            string `yaml:"endpoint"`
	SnapshotLocation    string `yaml:"snapshotLocation"`
	BackupConfig
}

// MasterConfig is used to backup/restore/configure master nodes
type MasterConfig struct {
	CertDir          string `yaml:"certDir"`
	CaCertFile       string `yaml:"caCertFilename"`
	CaKeyFile        string `yaml:"caKeyFilename"`
	SaPubFile        string `yaml:"saPubFilename"`
	SaKeyFile        string `yaml:"saKeyFilename"`
	ProxyCaCertFile  string `yaml:"proxyCaCertFilename"`
	ProxyKeyCertFile string `yaml:"proxyKeyCertFilename"`
	BackupConfig
}

// NodeConfig is used to backup/restore/configure worker nodes (backup and restore have an empty implementation right now)
type NodeConfig struct {
	CloudProvider string `yaml:"caKeyFilename"`
}

// BackupConfig are used to generalyze backuconfiguration
type BackupConfig struct {
	BackupFrequency string `yaml:"backupFrequency"`
	BackupRetention string `yaml:"backupRetention"`
}

func downloadFilesToDirectory(files []string, directory string, fromPath string, store *storage.Data) error {
	for _, filename := range files {
		file := filepath.Join(directory, filename)
		//os.Remove(file)
		newFile, err := os.Create(file)
		if err != nil {
			return err
		}
		bucketPath := filepath.Join(fromPath, filename)
		err = store.Download(bucketPath, newFile)
		if err != nil {
			log.Println("no %s found in bucket", bucketPath)
			return err
		}
	}
	return nil
}
