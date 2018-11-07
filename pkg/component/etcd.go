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
	"context"
	"fmt"
	"git.incubator.sh/sighup/furyctl/pkg/storage"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/clientv3/snapshot"
	"go.etcd.io/etcd/pkg/transport"
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"time"
)

// Etcd implements the ClusterComponent Interface
type Etcd struct{}

func getEtcdCfg(c *ClusterConfig, store *storage.Data) (*clientv3.Config, error) {
	cfg := clientv3.Config{
		Endpoints:   []string{c.Etcd.Endpoint},
		DialTimeout: 5 * time.Second,
	}
	// Setup TLS config if CAFile is provided into configurations
	if c.Etcd.CaCertFilename != "" {
		tlsInfo := transport.TLSInfo{
			CertFile:      fmt.Sprintf("%s/%s", c.Etcd.CertDir, c.Etcd.ClientCertFilename),
			KeyFile:       fmt.Sprintf("%s/%s", c.Etcd.CertDir, c.Etcd.ClientKeyFilename),
			TrustedCAFile: fmt.Sprintf("%s/%s", c.Etcd.CertDir, c.Etcd.CaCertFilename),
		}
		tlsConfig, err := tlsInfo.ClientConfig()
		if err != nil {
			return nil, err
		}
		cfg.TLS = tlsConfig
	}
	// Creating etcd client
	return &cfg, nil
}

// Backup implements
func (e Etcd) Backup(c *ClusterConfig, store *storage.Data) error {
	cfg, err := getEtcdCfg(c, store)
	filePath := filepath.Join(c.Etcd.SnapshotLocation, c.Etcd.SnapshotFilename)
	if err != nil {
		return err
	}
	sp := snapshot.NewV3(zap.NewExample())
	err = sp.Save(context.Background(), *cfg, filePath)
	if err != nil {
		return err
	}
	return store.UploadFile(fmt.Sprintf("%s/%s", c.NodeName, c.Etcd.SnapshotFilename), filePath)
}

// Restore implements
func (e Etcd) Restore(c *ClusterConfig, store *storage.Data) error {
	filePath := filepath.Join(c.Etcd.SnapshotLocation, c.Etcd.SnapshotFilename)
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	err = store.Download(fmt.Sprintf("%s/%s", c.NodeName, c.Etcd.SnapshotFilename), f)
	if err != nil {
		return err
	}
	restoreConf := snapshot.RestoreConfig{
		SnapshotPath: filePath,
		Name:         c.NodeName,
		// probably we'll have to modify this part to handle ha etcd
		InitialCluster: fmt.Sprintf("%s=%s", c.NodeName, c.Etcd.Endpoint),
		OutputDataDir:  filepath.Join(os.TempDir(), fmt.Sprint(time.Now().Nanosecond())),
		PeerURLs:       []string{c.Etcd.Endpoint},
	}
	sp := snapshot.NewV3(zap.NewExample())
	return sp.Restore(restoreConf)
}

// Configure implements
func (e Etcd) Configure(c ClusterConfig) error {
	return nil
}
