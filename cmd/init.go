package cmd

import (
	"log"

	"github.com/sighup-io/furyagent/pkg/component"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Executes initialization, uploads ca files",
	Long:  ``,
}
var initDir string
var data component.ClusterComponentData

var etcdInitCmd = &cobra.Command{
	Use:   "etcd",
	Short: "uploads etcd certificates to s3",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		// Does what is suppose to do
		var etcd component.ClusterComponent = component.Etcd{data}
		err := etcd.Init(initDir)
		if err != nil {
			log.Fatal(err)
		}
	},
}

var masterInitCmd = &cobra.Command{
	Use:   "master",
	Short: "uploads master certificates to s3",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		var master component.ClusterComponent = component.Master{data}
		err := master.Init(initDir)
		if err != nil {
			log.Fatal(err)
		}
	},
}

var openVpnInitCmd = &cobra.Command{
	Use:   "openvpn",
	Short: "uploads openvpn certificates to s3",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		var openvpn component.ClusterComponent = component.OpenVPN{data}
		err := openvpn.Init(initDir)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.PersistentFlags().StringVarP(&initDir, "directory", "d", ".", "directory with files to be uploaded (default is .)")

	initCmd.AddCommand(etcdInitCmd)
	initCmd.AddCommand(masterInitCmd)
	initCmd.AddCommand(openVpnInitCmd)
}
