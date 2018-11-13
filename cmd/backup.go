package cmd

import (
	"git.incubator.sh/sighup/furyagent/pkg/component"
	"git.incubator.sh/sighup/furyagent/pkg/storage"
	"github.com/spf13/cobra"
	"log"
)

func getConfig(cfgFile string) (*AgentConfig, *storage.Data) {
	// Executed only when another argument is passed, e.g. "backup etcd"
	// "backup" will print usage as desired
	// Reads the configuration file
	agentConfig, err := InitAgent(cfgFile)
	if err != nil {
		log.Fatal(err)
	}
	// Initializes the storage
	store, err := storage.Init(&agentConfig.Storage)
	if err != nil {
		log.Fatal(err)
	}
	return agentConfig, store

}

// backupCmd represents the `furyctl backup` command
var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Executes backups",
	Long:  ``,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		agentConfig, store = getConfig(cfgFile)
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
