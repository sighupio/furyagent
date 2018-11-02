package cmd

import (
	"log"

	"git.incubator.sh/sighup/furyctl/pkg/component"
	"git.incubator.sh/sighup/furyctl/pkg/storage"
	"github.com/spf13/cobra"
)

var cfgFile string

// backupCmd represents the `furyctl backup` command
var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Executes backups",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

// etcdBackupCmd represents the `furyctl backup etcd` command
var etcdBackupCmd = &cobra.Command{
	Use:   "etcd",
	Short: "Backups etcd node",
	Long:  `Backups etcd node`,
	Run: func(cmd *cobra.Command, args []string) {
		// Reads the configuration file
		cfg, err := InitAgent(cfgFile)
		if err != nil {
			log.Fatal(err)
		}
		// Initializes the storage
		store, err := storage.Init(&cfg.Storage)
		if err != nil {
			log.Fatal(err)
		}
		// Does what is suppose to do
		etcd := component.Etcd{}
		err = etcd.Backup(&cfg.ClusterComponent, store)
		if err != nil {
			log.Fatal(err)
		}
	},
}

// masterBackupCmd represents the `furyctl backup master` command
var masterBackupCmd = &cobra.Command{
	Use:   "master",
	Short: "Backups master node",
	Long:  `Backups master node`,
	Run: func(cmd *cobra.Command, args []string) {

	},
}

func init() {
	rootCmd.AddCommand(backupCmd)
	backupCmd.PersistentFlags().StringVar(&cfgFile, "config", "furyagent.yml", "config file (default is `furyagent.yaml`)")
	backupCmd.AddCommand(etcdBackupCmd)
	backupCmd.AddCommand(masterBackupCmd)
}
