package cmd

import (
	"github.com/spf13/cobra"
)

// restoreCmd represents the `furyctl restore` subcommand
var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Executes restores",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

// etcdRestoreCmd represents the `furyctl restore etcd` command
var etcdRestoreCmd = &cobra.Command{
	Use:   "etcd",
	Short: "Restores etcd node",
	Long:  `Restores etcd node`,
	Run: func(cmd *cobra.Command, args []string) {

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
	restoreCmd.AddCommand(etcdRestoreCmd)
	restoreCmd.AddCommand(masterRestoreCmd)
}
