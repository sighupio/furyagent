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

// ClusterComponent interface represent the basic concept of the componet: etcd, master, node
type ClusterComponent interface {
	Backup(ClusterConfig) error
	Restore(ClusterConfig) error
	Configure(ClusterConfig) error
}

// ClusterConfig represents the configuration for the whole cluster
type ClusterConfig struct {
	NodeName string
	Etcd     EtcdConfig
	Master   MasterConfig
	Node     NodeConfig
}

// EtcdConfig is used to backup/restore/configure etcd nodes
type EtcdConfig struct {
	DataDir          string
	CertDir          string
	CaCertFile       string
	CaKeyFile        string
	ClientCertFile   string
	ClientKeyFile    string
	Endpoint         string
	SnapshotLocation string
	BackupConfig
}

// MasterConfig is used to backup/restore/configure master nodes
type MasterConfig struct {
	CaCertFile string
	CaKeyFile  string
	BackupConfig
}

// NodeConfig is used to backup/restore/configure worker nodes (backup and restore have an empty implementation right now)
type NodeConfig struct {
	CloudProvider string
}

// BackupConfig are used to generalyze backuconfiguration
type BackupConfig struct {
	BackupFrequency string
	BackupRetention string
}
