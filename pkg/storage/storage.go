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

package storage

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/graymeta/stow"
	"github.com/graymeta/stow/azure"
	"github.com/graymeta/stow/google"
	"github.com/graymeta/stow/local"
	"github.com/graymeta/stow/s3"
)

type bufferWriteCloser struct {
	Buf *bytes.Buffer
}

func (bwc bufferWriteCloser) Write(p []byte) (n int, err error) {
	return bwc.Buf.Write(p)
}

func (bwc bufferWriteCloser) Close() error {
	return nil
}

// Data represent where to put whatever you're downloading
type Data struct {
	location      stow.Location
	containerName string
	container     stow.Container
}

// Init tests the credentials, the write access and list access
func Init(cfg *Config) (*Data, error) {
	s := new(Data)

	config := stow.ConfigMap{}
	switch cfg.Provider {
	case "s3":
		s.containerName = cfg.BucketName
		if cfg.URL != "" {
			config = stow.ConfigMap{
				s3.ConfigAccessKeyID: cfg.AccessKey,
				s3.ConfigSecretKey:   cfg.SecretKey,
				s3.ConfigEndpoint:    cfg.URL,
				s3.ConfigRegion:      cfg.Region,
			}
		} else {
			config = stow.ConfigMap{
				s3.ConfigAccessKeyID: cfg.AccessKey,
				s3.ConfigSecretKey:   cfg.SecretKey,
				s3.ConfigRegion:      cfg.Region,
			}
		}
	case "azure":
		s.containerName = cfg.BucketName
		config = stow.ConfigMap{
			azure.ConfigAccount: cfg.AzureStorageAccount,
			azure.ConfigKey:     cfg.AzureStorageKey,
		}
	case "google":
		s.containerName = cfg.BucketName
		sa, err := ioutil.ReadFile(cfg.GoogleServiceAccount)
		if err != nil {
			return nil, fmt.Errorf("Cannot read Google Service Account file %s: %v", cfg.GoogleServiceAccount, err)
		}
		config = stow.ConfigMap{
			google.ConfigJSON:      string(sa),
			google.ConfigProjectId: cfg.GoogleProjectId,
		}
	case "local":
		config = stow.ConfigMap{
			local.ConfigKeyPath: cfg.LocalPath,
		}
		s.containerName = cfg.LocalPath
	default:
		return nil, fmt.Errorf("provider \"%s\" not supported", cfg.Provider)
	}
	location, err := stow.Dial(cfg.Provider, config)
	if err != nil {
		return nil, fmt.Errorf("Cannot dial to %s: %v", cfg.Provider, err)
	}
	s.location = location
	container, err := s.getContainer()
	if err != nil {
		return nil, fmt.Errorf("Cannot get container %s: %v", s.containerName, err)
	}
	s.container = container
	return s, nil
}

func (s *Data) getContainer() (stow.Container, error) {
	container, err := s.location.Container(s.containerName)
	if err == stow.ErrNotFound {
		log.Println("Container not found, trying to create one!")
		container, err = s.location.CreateContainer(s.containerName)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		log.Println("Generic error accessing the container: ", s.containerName)
		return nil, err
	}

	return container, nil
}

// Close closes the open connection to the remote or local stow.Location
func (s *Data) Close() error {
	return s.location.Close()
}

// Download is the single interface to download something from Object Storage
func (s *Data) Download(filename string, obj io.WriteCloser) error {
	item, err := s.container.Item(filename)
	if err == stow.ErrNotFound {
		return fmt.Errorf("Item %s not found", filename)
	} else if err != nil {
		return err
	}
	name := item.Name()
	size, err := item.Size()
	if err != nil {
		return err
	}
	log.Printf("Item %s found [size: %d]\n", name, size)
	log.Printf("Saving item %s ...", name)
	reader, err := item.Open()
	if err != nil {
		return err
	}
	defer reader.Close()
	_, err = io.Copy(obj, reader)
	if err != nil {
		return err
	}
	return nil
}

// Upload is the single interface to upload something to Object Storage
func (s *Data) Upload(filename string, size int64, obj io.ReadCloser) error {
	//upload snapshot to container with given name
	defer obj.Close()
	if _, err := s.container.Item(filename); err == nil {
		log.Fatalf("%s exists already", filename)
	}
	item, err := s.container.Put(filename, obj, size, nil)
	if err != nil {
		return err
	}
	log.Println("Item URL: ", item.URL())
	if err != nil {
		return err
	}
	return nil
}

func (s *Data) UploadForce(filename string, size int64, obj io.ReadCloser) error {
	//upload snapshot to container with given name
	defer obj.Close()
	item, err := s.container.Put(filename, obj, size, nil)
	if err != nil {
		return err
	}
	log.Println("Item URL: ", item.URL())
	if err != nil {
		return err
	}
	return nil
}

// Remove removes the filename with the given path
func (s *Data) Remove(filename string) error {
	return s.container.RemoveItem(filename)
}

func (s *Data) UploadFile(filename, localPath string) error {
	log.Printf("uploading %s to %s", localPath, filename)
	fileSize, err := FileSize(localPath)
	if err != nil {
		return err
	}
	r, err := os.Open(localPath)
	if err != nil {
		return err
	}
	return s.Upload(filename, fileSize, r)
}

func (s *Data) UploadFileForce(filename, localPath string) error {
	log.Printf("uploading %s to %s", localPath, filename)
	fileSize, err := FileSize(localPath)
	if err != nil {
		return err
	}
	r, err := os.Open(localPath)
	if err != nil {
		return err
	}
	return s.UploadForce(filename, fileSize, r)
}

func (store *Data) UploadFilesFromDirectory(files [][]string, localDir string, toPath string) error {
	for _, fileSrcDest := range files {
		local, remote := filepath.Join(localDir, fileSrcDest[0]), filepath.Join(toPath, fileSrcDest[1])
		log.Printf("trying to upload %s to %s", local, remote)
		err := store.UploadFile(remote, local)
		if err != nil {
			return err
		}
	}
	return nil
}

func (store *Data) UploadFilesFromDirectoryWithForce(files [][]string, localDir string, toPath string) error {
	for _, fileSrcDest := range files {
		local, remote := filepath.Join(localDir, fileSrcDest[0]), filepath.Join(toPath, fileSrcDest[1])
		log.Printf("trying to upload %s to %s", local, remote)
		err := store.UploadFileForce(remote, local)
		if err != nil {
			return err
		}
	}
	return nil
}

func (store *Data) UploadFilesFromMemory(files map[string][]byte, dir string) error {
	for filename, file := range files {
		path := filepath.Join(dir, filename)
		if _, err := store.container.Item(path); err == nil {
			log.Fatalf("%s exists already", path)
		}
		if _, err := store.container.Put(path, ioutil.NopCloser(bytes.NewReader(file)), int64(len(file)), nil); err != nil {
			log.Fatal(err)
		}
	}
	return nil
}

func (store *Data) DownloadFilesToDirectory(files [][]string, localDir string, fromPath string, overwrite bool) error {
	os.MkdirAll(localDir, 0750)
	for _, fileSrcDst := range files {
		local, remote := fileSrcDst[0], fileSrcDst[1]
		file := filepath.Join(localDir, local)
		if overwrite {
			os.Remove(file)
		} else if _, err := os.Stat(file); !os.IsNotExist(err) {
			log.Fatalf("file %s already exists, use --overwrite=true", file)
		}
		newFile, err := os.Create(file)
		if err != nil {
			return err
		}
		bucketPath := filepath.Join(fromPath, remote)
		err = store.Download(bucketPath, newFile)
		if err != nil {
			log.Println("no %s found in bucket", bucketPath)
			return err
		}
	}
	return nil
}

func (store *Data) DownloadFilesToMemory(files []string, fromPath string) (map[string][]byte, error) {
	out := make(map[string][]byte)
	for _, fn := range files {
		bwc := bufferWriteCloser{new(bytes.Buffer)}
		err := store.Download(filepath.Join(fromPath, fn), bwc)
		if err != nil {
			return nil, err
		}
		fileContent := make([]byte, bwc.Buf.Len())
		_, err = bwc.Buf.Read(fileContent)
		if err != nil {
			return nil, err
		}
		out[fn] = fileContent
	}
	return out, nil
}
