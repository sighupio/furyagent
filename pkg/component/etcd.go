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
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/clientv3/snapshot"
	"go.etcd.io/etcd/pkg/transport"
	"go.uber.org/zap"
	"log"
	"os"
	"path/filepath"
	"time"
)

const (
	EtcdCaCrt = "ca.crt"
	EtcdCaKey = "ca.key"
)

// Etcd implements the ClusterComponent Interface
type Etcd struct {
	ClusterComponentData
}

func getEtcdCfg(c EtcdConfig) (*clientv3.Config, error) {
	cfg := clientv3.Config{
		Endpoints:   []string{c.Endpoint},
		DialTimeout: 5 * time.Second,
	}
	// Setup TLS config if CAFile is provided into configurations
	if c.ClientCertFilename != "" {
		tlsInfo := transport.TLSInfo{
			CertFile:      filepath.Join(c.CertDir, c.ClientCertFilename),
			KeyFile:       filepath.Join(c.CertDir, c.ClientKeyFilename),
			TrustedCAFile: filepath.Join(c.CertDir, c.CaCertFilename),
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
func (e Etcd) Backup() error {
	filePath := filepath.Join(e.Etcd.SnapshotLocation, e.Etcd.SnapshotFilename)
	cfg, err := getEtcdCfg(e.Etcd)
	if err != nil {
		return err
	}
	sp := snapshot.NewV3(zap.NewExample())
	err = sp.Save(context.Background(), *cfg, filePath)
	if err != nil {
		return err
	}
	e.UploadFile(getBucketPathEtcd(e.ClusterConfig), filePath)
	return err
}

// Restore implements
func (e Etcd) Restore() error {
	// the snapshot location path
	filePath := filepath.Join(e.Etcd.SnapshotLocation, e.Etcd.SnapshotFilename)
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	// downloading the snapshot to the snapshot location
	bucketPath := getBucketPathEtcd(e.ClusterConfig)
	err = e.Download(bucketPath, f)
	if err != nil {
		log.Println("no %s found in bucket", bucketPath)
		return err
	}
	// removing bkups
	backupDir := e.Etcd.DataDir + ".bkup"
	err = os.RemoveAll(backupDir)
	if err != nil {
		return err
	}
	// moving old data to original_name.bkup
	err = os.Rename(e.Etcd.DataDir, backupDir)
	if err != nil {
		return err
	}
	restoreConf := snapshot.RestoreConfig{
		SnapshotPath: filePath,
		Name:         e.NodeName,
		// probably we'll have to modify this part to handle ha etcd
		InitialCluster:      fmt.Sprintf("%s=%s", e.NodeName, e.Etcd.Endpoint),
		InitialClusterToken: e.Etcd.InitialClusterToken,
		OutputDataDir:       e.Etcd.DataDir,
		PeerURLs:            []string{e.Etcd.Endpoint},
	}

	sp := snapshot.NewV3(zap.NewExample())

	return sp.Restore(restoreConf)
}

func (e Etcd) getFileMappings() [][]string {
	return [][]string{
		[]string{e.Etcd.CaCertFilename, EtcdCaCrt},
		[]string{e.Etcd.CaKeyFilename, EtcdCaKey},
	}
}

func (e Etcd) Configure(overwrite bool) error {
	// remove, create and download new certs
	files := e.getFileMappings()
	bucketDir := "pki/etcd"
	return e.DownloadFilesToDirectory(files, e.Etcd.CertDir, bucketDir, overwrite)
}

func (e Etcd) Init(dir string) error {
	// uploads new certs
	files := e.getFileMappings()
	bucketDir := "pki/etcd"
	return e.UploadFilesFromDirectory(files, dir, bucketDir)
}
