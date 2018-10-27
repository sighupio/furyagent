// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
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
	"archive/zip"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/clientv3/snapshot"
	"go.etcd.io/etcd/pkg/transport"
	"go.uber.org/zap"
)

var cacert string
var cert string
var key string
var certDir string
var zipname = "certificates.zip"

// saveCmd represents the save command
var saveCmd = &cobra.Command{
	Use:   "save <filepath>",
	Short: "Stores an etcd node backend snapshot to a given file",
	Long:  "Stores an etcd node backend snapshot to a given file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var cfg clientv3.Config
		dbPath := args[0]
		if cmd.Flags().Changed("cacert") ||
			cmd.Flags().Changed("cert") ||
			cmd.Flags().Changed("key") {
			tlsInfo := transport.TLSInfo{
				CertFile:      cert,
				KeyFile:       key,
				TrustedCAFile: cacert,
			}
			tlsConfig, err := tlsInfo.ClientConfig()
			if err != nil {
				log.Fatal(err)
			}
			cfg = clientv3.Config{
				Endpoints:   []string{endpoint},
				DialTimeout: 5 * time.Second,
				TLS:         tlsConfig,
			}
		} else {
			cfg = clientv3.Config{
				Endpoints:   []string{endpoint},
				DialTimeout: 5 * time.Second,
			}
		}
		createSnapshot(cfg, dbPath)
		saveCertificates(certDir)
	},
}

func init() {
	snapshotCmd.AddCommand(saveCmd)
	saveCmd.PersistentFlags().StringVar(&endpoint, "endpoint", "127.0.0.1:2379", "etcd host endpoint")
	saveCmd.PersistentFlags().StringVar(&cacert, "cacert", "", "Verify certificates of TLS-enabled secure servers using this CA bundle")
	saveCmd.PersistentFlags().StringVar(&cert, "cert", "", "Identify secure client using this TLS certificate file")
	saveCmd.PersistentFlags().StringVar(&key, "key", "", "Identify secure client using this TLS key file")
	saveCmd.PersistentFlags().StringVar(&certDir, "certdir", "/etc/ssl/etcd", "Etcd certificates folder")

	//saveCmd.MarkPersistentFlagRequired("endpoint")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// saveCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// saveCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func createSnapshot(cfg clientv3.Config, dbPath string) {
	cli, err := clientv3.New(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer cli.Close()

	sp := snapshot.NewV3(zap.NewExample())
	if err = sp.Save(context.Background(), cfg, dbPath); err != nil {
		fmt.Println(err)
	}
}

func saveCertificates(path string) {
	var capath = fmt.Sprintf("%s/ca.*", path)
	var cakeypath = fmt.Sprintf("%s/ca-key.*", path)
	var cafile []string
	var cakeyfile []string

	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Fatal(path, " not exist. Please provide the directory where your etcd certificates are stored")
	}
	//SHOULD DO KIND OF SANITY CHECK and size should be 1
	log.Println("Looking for certificate files under directory ", path)
	cafile, err := filepath.Glob(capath)
	if err != nil || len(cafile) == 0 {
		log.Fatal("Can't find any of these files: ca.pem, ca.cert, ca.cer, ca.crt in the specified directory.")
	} else {
		log.Println("Found ", cafile[0])
	}

	cakeyfile, err = filepath.Glob(cakeypath)
	if err != nil || len(cakeyfile) == 0 {
		log.Fatal("Can't find any of these files: ca-key.pem, ca-key.cert, ca-key.cer, ca-key.crt, ca-key.key in the specified directory.")
	} else {
		log.Println("Found ", cakeyfile[0])
	}

	log.Printf("Saving %s and %s in %s\n", cafile[0], cakeyfile[0], zipname)
	err = createZip(zipname, []string{cafile[0], cakeyfile[0]})
	if err != nil {
		log.Fatal("Failed to create archive file ", zipname)
	}
	log.Println("Created archive file: ", zipname)
}

func createZip(zipname string, files []string) error {
	newZipFile, err := os.Create(zipname)
	if err != nil {
		return err
	}
	defer newZipFile.Close()

	zipWriter := zip.NewWriter(newZipFile)
	defer zipWriter.Close()

	for _, file := range files {
		zipfile, err := os.Open(file)
		if err != nil {
			return err
		}
		defer zipfile.Close()

		info, err := zipfile.Stat()
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		header.Name = file
		header.Method = zip.Deflate

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}

		if _, err = io.Copy(writer, zipfile); err != nil {
			return err
		}
	}
	return nil
}
