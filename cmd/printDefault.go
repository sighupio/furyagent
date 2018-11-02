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
	"fmt"

	"github.com/spf13/cobra"
)

// InitFuryfile default initial Furyfile config
const InitFuryfile = `
roles:
  - name: kube-node
    version: master

bases:
  - name: monitoring/prometheus-operated
    version: master
  - name: monitoring/prometheus-operator
    version: master
`

// printDefaultCmd represents the printDefault command
var printDefaultCmd = &cobra.Command{
	Use:   "printDefault",
	Short: "Print the basic Furyfile used to generate an INFRA project for Sighup SRL",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(InitFuryfile)
	},
}

func init() {
	rootCmd.AddCommand(printDefaultCmd)
}
