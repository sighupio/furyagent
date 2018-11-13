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

// Configure implements
func (m Master) Configure(c *ClusterConfig, store *storage.Data) error {
	// remove, create and download new certs
	files := []string{c.Master.CaCertFile, c.Master.CaKeyFile,
		c.Master.SaKeyFile, c.Master.SaPubFile,
		c.Master.ProxyCaCertFile, c.Master.ProxyKeyCertFile,
	}
	bucketDir := filepath.Join("pki", "master")
	return downloadFilesToDirectory(files, c.Master.CertDir, bucketDir, store)
}
