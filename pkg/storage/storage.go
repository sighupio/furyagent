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
	"fmt"
	"io"
	"log"
	"os"

	"github.com/graymeta/stow"
	"github.com/graymeta/stow/local"
	"github.com/graymeta/stow/s3"
	//  "github.com/graymeta/stow/azure"
	//  "github.com/graymeta/stow/google"
	// "github.com/graymeta/stow/swift"
)

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
				// s3.ConfigRegion:      "",
			}
		} else {
			config = stow.ConfigMap{
				s3.ConfigAccessKeyID: cfg.AccessKey,
				s3.ConfigSecretKey:   cfg.SecretKey,
				s3.ConfigRegion:      cfg.Region,
			}
		}

	// case "google":
	// 	config = stow.ConfigMap{
	// 		google.ConfigJSON:      os.Getenv("GOOGLE_CLOUD_KEYFILE_JSON"),
	// 		google.ConfigProjectId: os.Getenv("GOOGLE_PROJECT"),
	// 		google.ConfigScopes:    "read-write",
	// 	}
	// case "azure":
	// 	config = stow.ConfigMap{
	// 		azure.ConfigAccount: os.Getenv("AZURE_CONFIG_ID"),
	// 		azure.ConfigKey:     os.Getenv("AZURE_CONFIG_KEY"),
	// 	}
	// case "swift":
	// 	config = stow.ConfigMap{
	// 		swift.ConfigUsername:      os.Getenv("OS_USERNAME"),
	// 		swift.ConfigKey:           os.Getenv("OS_TOKEN"),
	// 		swift.ConfigTenantName:    os.Getenv("OS_TENANT_NAME"),
	// 		swift.ConfigTenantAuthURL: os.Getenv("OS_AUTH_URL"),
	// 	}
	case "local":
		config = stow.ConfigMap{
			local.ConfigKeyPath: cfg.Path,
		}
		s.containerName = cfg.BackupFolder
	default:
		return nil, fmt.Errorf("provider \"%s\" not supported", cfg.Provider)
	}
	location, err := stow.Dial(cfg.Provider, config)
	if err != nil {
		return nil, fmt.Errorf("cannot dial to %s: %v", cfg.Provider, err)
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
		container, err = s.location.CreateContainer(s.containerName)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
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

func (s *Data) UploadFile(filename, localPath string) error {
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
