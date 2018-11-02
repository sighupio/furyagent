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
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var dev bool

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Download dependencies specified in Furyfile.yml",
	Long:  "Download dependencies specified in Furyfile.yml",
	Run: func(cmd *cobra.Command, args []string) {
		dev = cmd.Flag("dev").Changed
		install(dev)
	},
}

func install(dev bool) {
	viper.SetConfigType("yml")
	viper.AddConfigPath(".")
	viper.SetConfigName(configFile)
	configuration := new(Furyconf)
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}
	err := viper.Unmarshal(configuration)
	if err != nil {
		log.Fatalf("unable to decode into struct, %v", err)
	}

	err = configuration.Validate()
	if err != nil {
		log.Println("ERROR VALIDATING: ", err)
	}

	err = configuration.Download(dev)
	if err != nil {
		log.Println("ERROR DOWNLOADING: ", err)
	}

}

func init() {
	rootCmd.AddCommand(installCmd)
	installCmd.PersistentFlags().Bool("dev", false, "Download from development repo (git.incubator.sh)")
}
