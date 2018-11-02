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
	"log"
	"os"
	"path/filepath"

	"github.com/graymeta/stow"
	"github.com/graymeta/stow/s3"
	"github.com/spf13/cobra"
)

const KIND_S3 = "s3"
const KIND_GOOGLE = "google"
const DEFAULT_CONTAINER = "furyctl-etcd-snapshots"

var kind string
var setEndpoint bool
var s3Endpoint string
var containerName string

// uploadCmd represents the upload command
var uploadCmd = &cobra.Command{
	Use:   "upload <filepath>",
	Short: "Upload file(s) to a S3(or S3 compliant) storage",
	Long:  "Upload file(s) to a S3(or S3 compliant) storage",
	Args:  cobra.RangeArgs(1, 10),
	Run: func(cmd *cobra.Command, args []string) {
		kind := KIND_S3
		paths := args
		setEndpoint = cmd.Flags().Changed("endpoint")
		log.Printf("Furyctl is going to upload %d objects : %s", len(paths), paths)
		for i, path := range paths {
			log.Printf("%d. object: ", i+1)
			upload(kind, path)
		}
	},
}

func init() {
	etcdCmd.AddCommand(uploadCmd)
	uploadCmd.PersistentFlags().StringVar(&s3Endpoint, "endpoint", "",
		"Optional config value for changing s3 endpoint used for e.g. minio.io")
	uploadCmd.PersistentFlags().StringVar(&containerName, "container-name", DEFAULT_CONTAINER, "Name of container to upload file into.")
}

func upload(kind string, path string) {
	var config stow.ConfigMap
	switch kind {
	case KIND_S3:
		accessKeyID := os.Getenv("AWS_ACCESS_KEY_ID")
		secretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
		defaultRegion := os.Getenv("AWS_DEFAULT_REGION")

		if setEndpoint {
			//S3 Compliant
			config = stow.ConfigMap{
				s3.ConfigAccessKeyID: accessKeyID,
				s3.ConfigSecretKey:   secretAccessKey,
				s3.ConfigEndpoint:    s3Endpoint,
			}
		} else {
			//S3
			config = stow.ConfigMap{
				s3.ConfigAccessKeyID: accessKeyID,
				s3.ConfigSecretKey:   secretAccessKey,
				s3.ConfigRegion:      defaultRegion,
			}
		}
		uploadS3(config, path)
	case KIND_GOOGLE:
		fmt.Println("not implemented yet.")
	}
}

func uploadS3(config stow.ConfigMap, path string) {
	location, err := stow.Dial(KIND_S3, config)
	if err != nil {
		log.Fatal(err)
	}
	defer location.Close()

	//read snapshot file
	reader, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}

	name := filepath.Base(path)
	size := fileSize(path)

	log.Printf("Uploading... [file: %s , size: %d]\n", path, size)

	//create container containerName if not exist
	container, err := location.Container(containerName)
	if err == stow.ErrNotFound {
		container, err = location.CreateContainer(containerName)
		if err != nil {
			log.Fatal(err)
		} else {
			log.Printf("Created container %s\n", container.Name())
		}
	} else if err != nil {
		log.Fatal(err)
	}

	//upload snapshot to container with given name
	item, err := container.Put(name, reader, size, nil)
	log.Println("Item URL: ", item.URL())
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Uploaded: ", name)
}

func fileSize(path string) int64 {
	fi, err := os.Stat(path)
	if err != nil {
		log.Fatal(err)
	}

	return fi.Size()
}
