package cmd

import (
	"git.incubator.sh/sighup/furyagent/pkg/component"
	"github.com/spf13/cobra"
	"log"
)

// restoreCmd represents the `furyctl restore` subcommand
var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Executes restores",
	Long:  ``,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		agentConfig, store = getConfig(cfgFile)
	},
}

// etcdRestoreCmd represents the `furyctl restore etcd` command
var etcdRestoreCmd = &cobra.Command{
	Use:   "etcd",
	Short: "Restores etcd node",
	Long:  `Restores etcd node`,
	Run: func(cmd *cobra.Command, args []string) {
		var etcd component.ClusterComponent = component.Etcd{}
		err := etcd.Restore(&agentConfig.ClusterComponent, store)
		if err != nil {
			log.Fatal(err)
		}
	},
}

// masterBackupCmd represents the `furyctl restore master` command
var masterRestoreCmd = &cobra.Command{
	Use:   "master",
	Short: "Restores master node",
	Long:  `Restores master node`,
	Run: func(cmd *cobra.Command, args []string) {

	},
}

func init() {
	rootCmd.AddCommand(restoreCmd)
	restoreCmd.PersistentFlags().StringVar(&cfgFile, "config", "furyagent.yml", "config file (default is `furyagent.yaml`)")
	restoreCmd.AddCommand(etcdRestoreCmd)
	restoreCmd.AddCommand(masterRestoreCmd)
}
