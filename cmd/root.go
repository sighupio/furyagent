// Copyright © 2018 Sighup SRL support@sighup.io
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/sighup-io/furyagent/pkg/component"
	"github.com/sighup-io/furyagent/pkg/storage"
	"github.com/spf13/cobra"
	"log"
	"os"
)

var cfgFile string
var store *storage.Data
var agentConfig *AgentConfig

// Execute is the main entrypoint of furyctl
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func getConfig(cfgFile string) (*AgentConfig, *storage.Data) {
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

func init() {
	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.furyctl.yaml)")
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "furyagent.yml", "config file (default is `furyagent.yaml`)")
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(printParsedConfig)
	rootCmd.AddCommand(printDefaultCmd)
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "furyagent",
	Short: "A command line tool to manage cluster deployment with kubernetes",
	Long:  ``,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		agentConfig, store = getConfig(cfgFile)
		data = component.ClusterComponentData{&agentConfig.ClusterComponent, store}
	},
}

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Prints the client version information",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Furyagent version ", FuryctlVersion)
	},
}

// printDefaultCmd represents the printDefault command
var printDefaultCmd = &cobra.Command{
	Use:   "printDefault",
	Short: "Prints a basic Furyfile used to generate an INFRA project",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		//fmt.Println(InitFuryfile)
	},
}

var printParsedConfig = &cobra.Command{
	Use:   "parsed-config",
	Short: "Prints the parsed furyagent.yaml file",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		conf, err := json.MarshalIndent(agentConfig, "", " ")
		if err != nil {
			log.Print(err)
		}
		fmt.Print(string(conf))
	},
}
