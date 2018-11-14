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
)

// ClusterComponent interface represent the basic concept of the componet: etcd, master, node
type ClusterComponent interface {
	Backup(*ClusterConfig, *storage.Data) error
	Restore(*ClusterConfig, *storage.Data) error
	Configure(*ClusterConfig, *storage.Data, bool) error
	Init(*ClusterConfig, *storage.Data, string) error
}

// ClusterConfig represents the configuration for the whole cluster
type ClusterConfig struct {
	NodeName string       `mapstructure:"nodeName"`
	Etcd     EtcdConfig   `mapstructure:"etcd"`
	Master   MasterConfig `mapstructure:"master"`
	Node     NodeConfig   `mapstructure:"node"`
}

// EtcdConfig is used to backup/restore/configure etcd nodes
type EtcdConfig struct {
	DataDir             string `mapstructure:"dataDir"`
	CertDir             string `mapstructure:"certDir"`
	CaCertFilename      string `mapstructure:"caCertFilename"`
	CaKeyFilename       string `mapstructure:"caKeyFilename"`
	ClientCertFilename  string `mapstructure:"clientCertFilename"`
	InitialClusterToken string `mapstructure:"initialClusterToken"`
	SnapshotFilename    string `mapstructure:"snapshotFilename"`
	ClientKeyFilename   string `mapstructure:"clientKeyFilename"`
	Endpoint            string `mapstructure:"endpoint"`
	SnapshotLocation    string `mapstructure:"snapshotLocation"`
	BackupConfig
}

// MasterConfig is used to backup/restore/configure master nodes
type MasterConfig struct {
	CertDir          string `mapstructure:"certDir"`
	CaCertFile       string `mapstructure:"caCertFilename"`
	CaKeyFile        string `mapstructure:"caKeyFilename"`
	SaPubFile        string `mapstructure:"saPubFilename"`
	SaKeyFile        string `mapstructure:"saKeyFilename"`
	ProxyCaCertFile  string `mapstructure:"proxyCaCertFilename"`
	ProxyKeyCertFile string `mapstructure:"proxyKeyCertFilename"`
	BackupConfig
}

// NodeConfig is used to backup/restore/configure worker nodes (backup and restore have an empty implementation right now)
type NodeConfig struct {
	CloudProvider string `mapstructure:"caKeyFilename"`
}

// BackupConfig are used to generalyze backuconfiguration
type BackupConfig struct {
	BackupFrequency string `mapstructure:"backupFrequency"`
	BackupRetention string `mapstructure:"backupRetention"`
}
