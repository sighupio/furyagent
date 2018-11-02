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
	"bufio"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/graymeta/stow"
	"github.com/graymeta/stow/s3"
	"github.com/spf13/cobra"
)

//var containerName string
//var s3Endpoint string

const BUFFERSIZE = 1024

// downloadCmd represents the download command
var downloadCmd = &cobra.Command{
	Use:   "download <itemname>",
	Short: "Download a db snapshot from a S3(or S3 compliant) storage",
	Long:  "Download a db snapshot from a S3(or S3 compliant) storage",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("download called")
		kind = KIND_S3
		itemName := args[0]
		setEndpoint = cmd.Flags().Changed("endpoint")
		download(itemName)
	},
}

func init() {
	etcdCmd.AddCommand(downloadCmd)
	downloadCmd.PersistentFlags().StringVar(&s3Endpoint, "endpoint", "",
		"Optional config value for changing s3 endpoint used for e.g. minio.io")
	downloadCmd.PersistentFlags().StringVar(&containerName, "container-name", DEFAULT_CONTAINER, "Name of container to download file from.")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// downloadCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// downloadCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func download(itemName string) {
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
		downloadS3(config, itemName)
	case KIND_GOOGLE:
		fmt.Println("not implemented yet.")
	}
}

func downloadS3(config stow.ConfigMap, itemName string) {
	location, err := stow.Dial(KIND_S3, config)
	if err != nil {
		log.Fatal(err)
	}
	defer location.Close()

	container, err := location.Container(containerName)
	if err == stow.ErrNotFound {
		log.Fatalf("Container %s not found", containerName)
	} else if err != nil {
		log.Fatal(err)
	}

	item, err := container.Item(itemName)
	if err == stow.ErrNotFound {
		log.Fatalf("Item %s not found", itemName)
	} else if err != nil {
		log.Fatal(err)
	}

	name := item.Name()
	size, err := item.Size()
	log.Printf("Item %s found [size: %d]\n", name, size)

	log.Printf("Saving item %s ...", name)
	err = saveItem(item, name)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Saved item %s [size: %d]\n", name, fileSize(name))

}

func saveItem(item stow.Item, path string) error {
	data := make([]byte, BUFFERSIZE)

	reader, err := item.Open()
	if err != nil {
		return err
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}

	defer func() {
		reader.Close()
		file.Close()
	}()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	for {
		n, err := reader.Read(data)
		if err != nil {
			if err == io.EOF {
				_, err := writer.Write(data[:n])
				if err != nil {
					return err
				}
				break
			}
			return err
		}
		_, err = writer.Write(data[:n])
	}
	return nil
}
