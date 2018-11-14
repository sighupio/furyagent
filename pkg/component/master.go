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
	"path/filepath"
)

const (
	MasterSaKey     = "sa.key"
	MasterSaPub     = "sa.pub"
	MasterFProxyCrt = "front-proxy-ca.crt"
	MasterFProxyKey = "front-proxy-ca.key"
	MasterCaKey     = "ca.key"
	MasterCaCrt     = "ca.pub"
)

// Master implements the ClusterComponent interface
type Master struct{}

// Backup implements
func (m Master) Backup(c *ClusterConfig, store *storage.Data) error {
	return nil
}

// Restore implements
func (m Master) Restore(c *ClusterConfig, store *storage.Data) error {
	return nil
}
func (m Master) getFileMappings(c *ClusterConfig) [][]string {
	return [][]string{
		[]string{c.Master.CaCertFile, MasterCaCrt},
		[]string{c.Master.CaKeyFile, MasterCaKey},
		[]string{c.Master.SaKeyFile, MasterSaKey},
		[]string{c.Master.SaPubFile, MasterSaPub},
		[]string{c.Master.ProxyCaCertFile, MasterFProxyCrt},
		[]string{c.Master.ProxyKeyCertFile, MasterFProxyKey},
	}
}

// Configure implements
func (m Master) Configure(c *ClusterConfig, store *storage.Data, overwrite bool) error {
	// remove, create and download new certs
	files := m.getFileMappings(c)
	bucketDir := filepath.Join("pki", "master")
	return store.DownloadFilesToDirectory(files, c.Master.CertDir, bucketDir, overwrite)
}

func (m Master) Init(c *ClusterConfig, store *storage.Data, dir string) error {
	// remove, create and download new certs
	files := m.getFileMappings(c)
	bucketDir := "pki/master"
	return store.UploadFilesFromDirectory(files, dir, bucketDir)
}
