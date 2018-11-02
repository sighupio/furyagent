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

// import (
// 	"log"

// 	"github.com/spf13/cobra"
// 	"go.etcd.io/etcd/clientv3/snapshot"
// 	"go.uber.org/zap"
// )

// var name string
// var dataDir string
// var initialCluster string
// var initialClusterToken string
// var initialAdvertisePeerUrls []string

// // restoreCmd represents the restore command
// var restoreCmd = &cobra.Command{
// 	Use:   "restore <filename>",
// 	Short: "Restores an etcd member snapshot to an etcd directory",
// 	Long:  "Restores an etcd member snapshot to an etcd directory",
// 	Args:  cobra.ExactArgs(1),
// 	Run: func(cmd *cobra.Command, args []string) {
// 		//fmt.Println("restore called")
// 		dbPath := args[0]
// 		cfg := snapshot.RestoreConfig{
// 			SnapshotPath:        dbPath,
// 			Name:                name,
// 			OutputDataDir:       dataDir,
// 			InitialCluster:      initialCluster,
// 			InitialClusterToken: initialClusterToken,
// 			PeerURLs:            initialAdvertisePeerUrls,
// 		}
// 		restore(dbPath, cfg)
// 	},
// }

// func init() {
// 	snapshotCmd.AddCommand(restoreCmd)
// 	restoreCmd.PersistentFlags().StringVar(&name, "name", "default", "Human-readable name for this member")
// 	restoreCmd.PersistentFlags().StringVar(&dataDir, "data-dir", "", "Path to the data directory")
// 	restoreCmd.PersistentFlags().StringVar(&initialCluster, "initial-cluster", "default=http://localhost:2380", "Initial cluster configuration for restore bootstrap")
// 	restoreCmd.PersistentFlags().StringVar(&initialClusterToken, "initial-cluster-token", "etcd-cluster", "Initial cluster token for the etcd cluster during restore bootstrap")
// 	restoreCmd.PersistentFlags().StringArrayVar(&initialAdvertisePeerUrls, "initial-advertise-peer-urls", []string{"http://localhost:2380"}, "List of this member's peer URLs to advertise to the rest of the cluster")
// }

// func restore(dbPath string, cfg snapshot.RestoreConfig) {
// 	sp := snapshot.NewV3(zap.NewExample())
// 	if err := sp.Restore(cfg); err != nil {
// 		log.Fatal(err)
// 	}
// }

/*func checkKeyvalue(key string, value string) {
	var cli *clientv3.Client
	cli, err := clientv3.New(clientv3.Config{Endpoints: []string{"http://localhost:2379"}})
	if err != nil {
		log.Fatal(err)
	}
	defer cli.Close()
	var gresp *clientv3.GetResponse
	gresp, err = cli.Get(context.Background(), key)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("val: ", gresp.Count)
	log.Println("val: ", gresp.Kvs)
}*/
