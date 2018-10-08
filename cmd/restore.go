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
	"fmt"
	"math/rand"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"go.etcd.io/etcd/clientv3/snapshot"
	"go.etcd.io/etcd/embed"

	"go.uber.org/zap"
)

// restoreCmd represents the restore command
var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("restore called")
		restoreCluster(1, "/tmp/snapshot747756240.db")
	},
}

func init() {
	etcdCmd.AddCommand(restoreCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// restoreCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// restoreCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func restoreCluster(clusterN int, dbPath string) (
	cURLs []url.URL,
	pURLs []url.URL,
	srvs []*embed.Etcd) {
	urls := newEmbedURLs(clusterN * 2)
	cURLs, pURLs = urls[:clusterN], urls[clusterN:]

	const testClusterTkn = "tkn"

	ics := ""
	for i := 0; i < clusterN; i++ {
		ics += fmt.Sprintf(",%d=%s", i, pURLs[i].String())
	}
	fmt.Println("icd: ", ics)
	ics = ics[1:]

	cfgs := make([]*embed.Config, clusterN)
	for i := 0; i < clusterN; i++ {
		cfg := embed.NewConfig()
		cfg.Logger = "zap"
		cfg.LogOutputs = []string{"/dev/null"}
		cfg.Debug = false
		cfg.Name = fmt.Sprintf("%d", i)
		cfg.InitialClusterToken = testClusterTkn
		cfg.ClusterState = "existing"
		//lcurls -> listen-client-urls
		//lpurls -> listen-peer-urls
		//acurls -> adverties-client-urls
		//apurls -> adverties-peer-urls
		cfg.LCUrls, cfg.ACUrls = []url.URL{cURLs[i]}, []url.URL{cURLs[i]}
		cfg.LPUrls, cfg.APUrls = []url.URL{pURLs[i]}, []url.URL{pURLs[i]}
		cfg.InitialCluster = ics
		cfg.Dir = filepath.Join(os.TempDir(), fmt.Sprint(time.Now().Nanosecond()+i))

		sp := snapshot.NewV3(zap.NewExample())
		if err := sp.Restore(snapshot.RestoreConfig{
			SnapshotPath:        dbPath,
			Name:                cfg.Name,
			OutputDataDir:       cfg.Dir,
			PeerURLs:            []string{pURLs[i].String()},
			InitialCluster:      ics,
			InitialClusterToken: cfg.InitialClusterToken,
		}); err != nil {
			fmt.Println(err)
		}
		cfgs[i] = cfg
	}

	sch := make(chan *embed.Etcd)
	for i := range cfgs {
		go func(idx int) {
			srv, err := embed.StartEtcd(cfgs[idx])
			if err != nil {
				fmt.Println(err)
			}

			<-srv.Server.ReadyNotify()
			sch <- srv
		}(i)
	}

	srvs = make([]*embed.Etcd, clusterN)
	for i := 0; i < clusterN; i++ {
		select {
		case srv := <-sch:
			srvs[i] = srv
		case <-time.After(5 * time.Second):
			fmt.Printf("#%d: failed to start embed.Etcd\n", i)
		}
	}
	return cURLs, pURLs, srvs
}

func newEmbedURLs(n int) (urls []url.URL) {
	urls = make([]url.URL, n)
	for i := 0; i < n; i++ {
		rand.Seed(int64(time.Now().Nanosecond()))
		u, _ := url.Parse(fmt.Sprintf("unix://localhost:%d", rand.Intn(45000)))
		urls[i] = *u
	}
	return urls
}
