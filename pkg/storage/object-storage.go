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
	"os"

	"github.com/graymeta/stow"
	"github.com/graymeta/stow/azure"
	"github.com/graymeta/stow/google"
	"github.com/graymeta/stow/local"
	"github.com/graymeta/stow/s3"
	"github.com/graymeta/stow/swift"
)

// Storage represent where to put whathever you're downloading
type Storage struct {
	Provider string        `yaml:"provider"`
	Address  string        `yaml:"address"`
	Prefix   string        `yaml:"prefix"`
	location stow.Location `yaml:"-"`
}

// Init tests the credentials, the write access and list access
func Init(provider string) (*Storage, error) {
	s := new(Storage)
	s.Provider = provider

	config := stow.ConfigMap{}
	switch s.Provider {
	case "s3":
		config = stow.ConfigMap{
			s3.ConfigAccessKeyID: os.Getenv("AWS_ACCESS_KEY_ID"),
			s3.ConfigSecretKey:   os.Getenv("AWS_SECRET_ACCESS_KEY"),
			s3.ConfigRegion:      os.Getenv("AWS_DEFAULT_REGION"),
		}
	case "s3-compatible":
		config = stow.ConfigMap{
			s3.ConfigAccessKeyID: os.Getenv("AWS_ACCESS_KEY_ID"),
			s3.ConfigSecretKey:   os.Getenv("AWS_SECRET_ACCESS_KEY"),
			s3.ConfigEndpoint:    os.Getenv("S3_ENDPOINT"),
		}
	case "google":
		config = stow.ConfigMap{
			google.ConfigJSON:      os.Getenv("GOOGLE_CLOUD_KEYFILE_JSON"),
			google.ConfigProjectId: os.Getenv("GOOGLE_PROJECT"),
			google.ConfigScopes:    "read-write",
		}
	case "azure":
		config = stow.ConfigMap{
			azure.ConfigAccount: os.Getenv("AZURE_CONFIG_ID"),
			azure.ConfigKey:     os.Getenv("AZURE_CONFIG_KEY"),
		}
	case "swift":
		config = stow.ConfigMap{
			swift.ConfigUsername:      os.Getenv("OS_USERNAME"),
			swift.ConfigKey:           os.Getenv("OS_TOKEN"),
			swift.ConfigTenantName:    os.Getenv("OS_TENANT_NAME"),
			swift.ConfigTenantAuthURL: os.Getenv("OS_AUTH_URL"),
		}
	case "local":
		config = stow.ConfigMap{
			local.ConfigKeyPath: os.Getenv("PATH"),
		}
	default:
		return nil, fmt.Errorf("provider \"%s\" not supported", s.Provider)
	}
	location, err := stow.Dial(s.Provider, config)
	if err != nil {
		return nil, fmt.Errorf("cannot dial to %s: %v", s.Provider, err)
	}
	s.location = location
	return s, nil
}

// Download is the single interface to download something from Object Storage
func (s *Storage) Download() error {
return nil
}

// Upload is the single interface to upload something to Object Storage
func (s *Storage) Upload() error {
	return nil
	}
	
