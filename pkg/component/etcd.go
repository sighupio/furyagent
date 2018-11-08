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
	"log"
	"os"
	"path/filepath"
	"time"
)

// Etcd implements the ClusterComponent Interface
type Etcd struct{}

func getEtcdCfg(c *ClusterConfig) (*clientv3.Config, error) {
	cfg := clientv3.Config{
		Endpoints:   []string{c.Etcd.Endpoint},
		DialTimeout: 5 * time.Second,
	}
	// Setup TLS config if CAFile is provided into configurations
	if c.Etcd.ClientCertFilename != "" {
		tlsInfo := transport.TLSInfo{
			CertFile:      filepath.Join(c.Etcd.CertDir, c.Etcd.ClientCertFilename),
			KeyFile:       filepath.Join(c.Etcd.CertDir, c.Etcd.ClientKeyFilename),
			TrustedCAFile: filepath.Join(c.Etcd.CertDir, c.Etcd.CaCertFilename),
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

func getBucketPathEtcd(c *ClusterConfig) string {
	return filepath.Join("etcd", c.NodeName, c.Etcd.SnapshotFilename)
}

// Backup implements
func (e Etcd) Backup(c *ClusterConfig, store *storage.Data) error {
	filePath := filepath.Join(c.Etcd.SnapshotLocation, c.Etcd.SnapshotFilename)
	cfg, err := getEtcdCfg(c)
	if err != nil {
		return err
	}
	sp := snapshot.NewV3(zap.NewExample())
	err = sp.Save(context.Background(), *cfg, filePath)
	if err != nil {
		return err
	}
	err = store.UploadFile(getBucketPathEtcd(c), filePath)
	if err != nil {
		return err
	}
	for _, filename := range []string{c.Etcd.CaCertFilename, c.Etcd.CaKeyFilename} {
		bucketPath := filepath.Join("pki", c.NodeName, filename)
		err = store.UploadFile(bucketPath, filepath.Join(c.Etcd.CertDir, filename))
		if err != nil {
			return err
		}
	}
	return nil
}

// Restore implements
func (e Etcd) Restore(c *ClusterConfig, store *storage.Data) error {
	// the snapshot location path
	filePath := filepath.Join(c.Etcd.SnapshotLocation, c.Etcd.SnapshotFilename)
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	// downloading the snapshot to the snapshot location
	err = store.Download(getBucketPathEtcd(c), f)
	if err != nil {
		log.Println("no %s found in bucket", getBucketPathEtcd(c))
		return err
	}
	// removing bkups
	backupDir := c.Etcd.DataDir + ".bkup"
	err = os.RemoveAll(backupDir)
	if err != nil {
		return err
	}
	// moving old data to original_name.bkup
	err = os.Rename(c.Etcd.DataDir, backupDir)
	if err != nil {
		return err
	}

	// remove, create and download new certs
	for _, filename := range []string{c.Etcd.CaCertFilename, c.Etcd.CaKeyFilename} {
		file := filepath.Join(c.Etcd.CertDir, filename)
		os.Remove(file)
		newFile, err := os.Create(file)
		if err != nil {
			return err
		}
		bucketPath := filepath.Join("pki", c.NodeName, filename)
		err = store.Download(bucketPath, newFile)
		if err != nil {
			log.Println("no %s found in bucket", bucketPath)
			return err
		}
	}

	restoreConf := snapshot.RestoreConfig{
		SnapshotPath: filePath,
		Name:         c.NodeName,
		// probably we'll have to modify this part to handle ha etcd
		InitialCluster:      fmt.Sprintf("%s=%s", c.NodeName, c.Etcd.Endpoint),
		InitialClusterToken: c.Etcd.InitialClusterToken,
		OutputDataDir:       c.Etcd.DataDir,
		PeerURLs:            []string{c.Etcd.Endpoint},
	}

	sp := snapshot.NewV3(zap.NewExample())

	return sp.Restore(restoreConf)
}

// Configure implements
func (e Etcd) Configure(c ClusterConfig) error {
	return nil
}
