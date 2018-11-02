package cmd

import (
	"github.com/spf13/cobra"
)

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
	backupCmd.AddCommand(etcdBackupCmd)
	backupCmd.AddCommand(masterBackupCmd)
}
