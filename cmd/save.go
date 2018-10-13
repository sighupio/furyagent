// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
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

package cmd

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/spf13/cobra"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/clientv3/snapshot"
	"go.etcd.io/etcd/pkg/transport"
	"go.uber.org/zap"
)

var cacert string
var cert string
var key string

// saveCmd represents the save command
var saveCmd = &cobra.Command{
	Use:   "save <filepath>",
	Short: "Stores an etcd node backend snapshot to a given file",
	Long:  "Stores an etcd node backend snapshot to a given file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var cfg clientv3.Config
		dbPath := args[0]
		if cmd.Flags().Changed("cacert") ||
			cmd.Flags().Changed("cert") ||
			cmd.Flags().Changed("key") {
			tlsInfo := transport.TLSInfo{
				CertFile:      cert,
				KeyFile:       key,
				TrustedCAFile: cacert,
			}
			tlsConfig, err := tlsInfo.ClientConfig()
			if err != nil {
				log.Fatal(err)
			}
			cfg = clientv3.Config{
				Endpoints:   []string{endpoint},
				DialTimeout: 5 * time.Second,
				TLS:         tlsConfig,
			}
		} else {
			cfg = clientv3.Config{
				Endpoints:   []string{endpoint},
				DialTimeout: 5 * time.Second,
			}
		}
		createSnapshot(cfg, dbPath)
	},
}

func init() {
	snapshotCmd.AddCommand(saveCmd)
	saveCmd.PersistentFlags().StringVar(&endpoint, "endpoint", "127.0.0.1:2379", "etcd host endpoint")
	saveCmd.PersistentFlags().StringVar(&cacert, "cacert", "", "Verify certificates of TLS-enabled secure servers using this CA bundle")
	saveCmd.PersistentFlags().StringVar(&cert, "cert", "", "Identify secure client using this TLS certificate file")
	saveCmd.PersistentFlags().StringVar(&key, "key", "", "Identify secure client using this TLS key file")

	//saveCmd.MarkPersistentFlagRequired("endpoint")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// saveCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// saveCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func createSnapshot(cfg clientv3.Config, dbPath string) {
	cli, err := clientv3.New(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer cli.Close()

	sp := snapshot.NewV3(zap.NewExample())
	if err = sp.Save(context.Background(), cfg, dbPath); err != nil {
		fmt.Println(err)
	}
}
