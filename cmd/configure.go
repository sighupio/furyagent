package cmd

import (
	"log"

	"github.com/sighupio/furyagent/pkg/component"
	"github.com/spf13/cobra"
)

// backupCmd represents the `furyctl backup` command
var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Executes configuration",
	Long:  ``,
}

var overwrite bool
var revoke bool

// etcdBackupCmd represents the `furyctl backup etcd` command
var etcdConfigCmd = &cobra.Command{
	Use:   "etcd",
	Short: "Configures etcd node",
	Long:  `Configures etcd node`,
	Run: func(cmd *cobra.Command, args []string) {
		// Does what is suppose to do
		var etcd component.ClusterComponent = component.Etcd{data}
		err := etcd.Configure(overwrite)
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
		var master component.ClusterComponent = component.Master{component.ClusterComponentData{&agentConfig.ClusterComponent, store}}
		err := master.Configure(overwrite)
		if err != nil {
			log.Fatal(err)
		}
	},
}

// NodeConfigureCmd represents the `furyagent configure node` command
var NodeConfigureCmd = &cobra.Command{
	Use:   "node",
	Short: "Get join.sh script from s3 and execute the join process",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		var node component.ClusterComponent = component.Node{data}
		err := node.Configure(overwrite)
		if err != nil {
			log.Fatal(err)
		}
	},
}

// openVPNConfigureCmd represents the `furyagent configure openvpn` command
var openVPNConfigCmd = &cobra.Command{
	Use:   "openvpn",
	Short: "Get OpenVPN certificates from s3",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		var openvpn component.ClusterComponent = component.OpenVPN{data}
		err := openvpn.Configure(overwrite)
		if err != nil {
			log.Fatal(err)
		}
	},
}

// openVPNConfigureCmd represents the `furyagent configure openvpn` command
var openVPNClientConfigCmd = &cobra.Command{
	Use:   "openvpn-client",
	Short: "Get OpenVPN users client from s3",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		var openvpn component.OpenVPNClient = component.OpenVPNClient{data}
		err := openvpn.Configure(overwrite, revoke)
		if err != nil {
			log.Fatal(err)
		}
	},
}

// SSHKeysConfigCmd represents the `furyagent configure ssh-keys` command
var SSHKeysConfigCmd = &cobra.Command{
	Use:   "ssh-keys",
	Short: "Setup ssh keys from s3",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		var ssh component.ClusterComponent = component.SSHComponent{data}
		err := ssh.Configure(overwrite)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(configureCmd)
	configureCmd.PersistentFlags().BoolVar(&overwrite, "overwrite", false, "overwrite config files (default is `false`)")
	configureCmd.AddCommand(etcdConfigCmd)
	configureCmd.AddCommand(masterConfigCmd)
	configureCmd.AddCommand(NodeConfigureCmd)
	configureCmd.AddCommand(openVPNConfigCmd)
	configureCmd.AddCommand(openVPNClientConfigCmd)
	configureCmd.AddCommand(SSHKeysConfigCmd)
	openVPNClientConfigCmd.PersistentFlags().BoolVar(&revoke, "revoke", false, "revoke client certificate (default is `false`)")
}
