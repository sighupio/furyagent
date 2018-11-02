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
	"fmt"
	"context"
	"time"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/clientv3/snapshot"
	"go.etcd.io/etcd/pkg/transport"
	"go.uber.org/zap"
)

// Etcd implements the ClusterComponent Interface
type Etcd struct{}

// Backup implements
func (e *Etcd) Backup(c *ClusterConfig) error {
	cfg := clientv3.Config{
		Endpoints:   []string{c.Etcd.Endpoint},
		DialTimeout: 5 * time.Second,
	}
	// Setup TLS config if CAFile is provided into configurations
	if c.Etcd.CaCertFile != "" {
		tlsInfo := transport.TLSInfo{
			CertFile:      fmt.Sprintf("%s/%s",c.Etcd.CertDir,c.Etcd.ClientCertFile),
			KeyFile:      fmt.Sprintf("%s/%s",c.Etcd.CertDir,c.Etcd.ClientKeyFile),
			TrustedCAFile: fmt.Sprintf("%s/%s",c.Etcd.CertDir,c.Etcd.CaCertFile),
		}
		tlsConfig, err := tlsInfo.ClientConfig()
		if err != nil {
			return err
		}
		cfg.TLS = tlsConfig
	}
	// Creating etcd client
	cli, err := clientv3.New(cfg)
	if err != nil {
		return err
	}
	defer cli.Close()
	
	sp := snapshot.NewV3(zap.NewExample())
	err = sp.Save(context.Background(), cfg, c.Etcd.SnapshotLocation)
	if err != nil {
		return err
	} 
	return nil
}

// Restore implements
func (e *Etcd) Restore(c *ClusterConfig) error {
	restoreConf := snapshot.RestoreConfig{
		SnapshotPath:        c.Etcd.SnapshotLocation,
		Name:                c.NodeName,
		OutputDataDir:       c.Etcd.DataDir,
		// PeerURLs:            initialAdvertisePeerUrls,
	}
	sp := snapshot.NewV3(zap.NewExample())
	if err := sp.Restore(restoreConf); err != nil {
		return err
	}
	return nil
}

// Configure implements
func (e *Etcd) Configure(c *ClusterConfig) error {
	return nil
}
