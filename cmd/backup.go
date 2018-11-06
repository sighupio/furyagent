package cmd

import (
	"git.incubator.sh/sighup/furyctl/pkg/component"
	"git.incubator.sh/sighup/furyctl/pkg/storage"
	"github.com/spf13/cobra"
	"log"
)

var cfgFile string
var store *storage.Data
var agentConfig *AgentConfig

// backupCmd represents the `furyctl backup` command
var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Executes backups",
	Long:  ``,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Executed only when another argument is passed, e.g. "backup etcd"
		// "backup" will print usage as desired
		// Reads the configuration file
		ac, err := InitAgent(cfgFile)
		agentConfig = ac
		if err != nil {
			log.Fatal(err)
		}
		// Initializes the storage
		s, err := storage.Init(&agentConfig.Storage)
		store = s
		if err != nil {
			log.Fatal(err)
		}
	},
}

// etcdBackupCmd represents the `furyctl backup etcd` command
var etcdBackupCmd = &cobra.Command{
	Use:   "etcd",
	Short: "Backups etcd node",
	Long:  `Backups etcd node`,
	Run: func(cmd *cobra.Command, args []string) {
		// Does what is suppose to do
		var etcd component.ClusterComponent = component.Etcd{}
		err := etcd.Backup(&agentConfig.ClusterComponent, store)
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
