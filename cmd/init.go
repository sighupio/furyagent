package cmd

import (
	"git.incubator.sh/sighup/furyagent/pkg/component"
	"github.com/spf13/cobra"
	"log"
)

// backupCmd represents the `furyctl backup` command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Executes initialization, uploads ca files",
	Long:  ``,
}
var initDir string

// etcdBackupCmd represents the `furyctl backup etcd` command
var etcdInitCmd = &cobra.Command{
	Use:   "etcd",
	Short: "uploads etcd certificates to s3",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		// Does what is suppose to do
		var etcd component.ClusterComponent = component.Etcd{}
		err := etcd.Init(&agentConfig.ClusterComponent, store, initDir)
		if err != nil {
			log.Fatal(err)
		}
	},
}

// masterBackupCmd represents the `furyctl backup master` command
var masterInitCmd = &cobra.Command{
	Use:   "master",
	Short: "uploads master certificates to s3",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		var master component.ClusterComponent = component.Master{}
		err := master.Init(&agentConfig.ClusterComponent, store, initDir)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.PersistentFlags().StringVar(&cfgFile, "config", "furyagent.yml", "config file (default is `furyagent.yaml`)")
	initCmd.PersistentFlags().StringVarP(&initDir, "directory", "d", ".", "directory with files to be uploaded (default is .)")
	initCmd.AddCommand(etcdInitCmd)
	initCmd.AddCommand(masterInitCmd)
}
