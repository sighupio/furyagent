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
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/clientv3/snapshot"
	"go.etcd.io/etcd/embed"
	"go.uber.org/zap"
)

// snapshotCmd represents the snapshot command
var snapshotCmd = &cobra.Command{
	Use:   "snapshot",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("called snapshot")
		dbpath := createSnapshotFile()
		fmt.Println(dbpath)
	},
}

func init() {
	etcdCmd.AddCommand(snapshotCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// snapshotCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:55
	// snapshotCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func createSnapshotFile() string {
	//kvs := []kv{{"foo1", "bar1"}, {"foo2", "bar2"}, {"foo3", "bar3"}}
	//Crea un local etcd server
	clusterN := 1
	urls := newEmbedURLs(clusterN * 2)
	cURLs, pURLs := urls[:clusterN], urls[clusterN:]
	cfg := embed.NewConfig()
	cfg.Logger = "zap"
	cfg.LogOutputs = []string{"/dev/null"}
	cfg.Debug = false
	cfg.Name = "default"
	cfg.ClusterState = "new"
	cfg.LCUrls, cfg.ACUrls = cURLs, cURLs
	cfg.LPUrls, cfg.APUrls = pURLs, pURLs
	cfg.InitialCluster = fmt.Sprintf("%s=%s", cfg.Name, pURLs[0].String())
	cfg.Dir = filepath.Join(os.TempDir(), fmt.Sprint(time.Now().Nanosecond()))
	srv, err := embed.StartEtcd(cfg) //-> launch server

	if err != nil {
		fmt.Println(err.Error)
	}
	defer func() {
		os.RemoveAll(cfg.Dir)
		srv.Close()
	}()
	select {
	case <-srv.Server.ReadyNotify():
		log.Printf("Server is ready!")
	case <-time.After(3 * time.Second):
		fmt.Println("failed to start embed.Etcd for creating snapshots")
	}

	//configure and create etcd client
	ccfg := clientv3.Config{Endpoints: []string{cfg.ACUrls[0].String()}}
	cli, err := clientv3.New(ccfg)
	if err != nil {
		fmt.Println(err.Error)
	}
	defer cli.Close()
	//put some key value pairs to etcd
	//for i := range kvs {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	//_, err = cli.Put(ctx, kvs[i].k, kvs[i].v)
	_, err = cli.Put(ctx, "key1", "value1")
	_, err = cli.Put(ctx, "key2", "value2")
	_, err = cli.Put(ctx, "key3", "value3")
	cancel()
	if err != nil {
		fmt.Println(err)
	}
	//	}

	//create snapshot manager
	sp := snapshot.NewV3(zap.NewExample())

	//determine snapshot path
	dpPath := filepath.Join(os.TempDir(), fmt.Sprintf("snapshot%d.db", time.Now().Nanosecond()))
	//save it
	if err = sp.Save(context.Background(), ccfg, dpPath); err != nil {
		fmt.Println(err)
	}

	os.RemoveAll(cfg.Dir)
	srv.Close()
	return dpPath
}
