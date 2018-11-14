package cmd

import (
	"git.incubator.sh/sighup/furyagent/pkg/component"
	"github.com/spf13/cobra"
	"log"
)

// backupCmd represents the `furyctl backup` command
var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Executes configuration",
	Long:  ``,
}
var overwrite bool

// etcdBackupCmd represents the `furyctl backup etcd` command
var etcdConfigCmd = &cobra.Command{
	Use:   "etcd",
	Short: "Configures etcd node",
	Long:  `Configures etcd node`,
	Run: func(cmd *cobra.Command, args []string) {
		// Does what is suppose to do
		var etcd component.ClusterComponent = component.Etcd{}
		err := etcd.Configure(&agentConfig.ClusterComponent, store, overwrite)
		if err != nil {
			log.Fatal(err)
		}
	},
}

// masterBackupCmd represents the `furyctl backup master` command
var masterConfigCmd = &cobra.Command{
	Use:   "master",
	Short: "Configures master node",
	Long:  `Configures master node`,
	Run: func(cmd *cobra.Command, args []string) {
		var master component.ClusterComponent = component.Master{}
		err := master.Configure(&agentConfig.ClusterComponent, store, overwrite)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(backupCmd)
	configureCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "furyagent.yml", "config file (default is `furyagent.yaml`)")
	configureCmd.PersistentFlags().BoolVarP(&overwrite, "overwite", "ow", false, "overwrite config files (default is `false`)")
	configureCmd.AddCommand(etcdConfigCmd)
	configureCmd.AddCommand(masterConfigCmd)
}
