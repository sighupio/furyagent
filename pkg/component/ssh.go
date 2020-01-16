package component

import (
	"bytes"
	"fmt"
	ioutil "io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"time"

	"gopkg.in/yaml.v2"
)

const (
	SSHFileName                   = "ssh-users.yml"
	SSHBucketDir                  = "ssh"
	SSHLocalDir                   = "secrets/ssh"
	SSHAuthorizedKeysFileName     = "authorized_keys"
	SSHAuthorizedKeysTempFileName = "authorized_keys_tmp"
)

type SSH struct {
	ClusterComponentData
}

type SSHUsersFile struct {
	Users []UserSpec `yaml:"users"`
}

type UserSpec struct {
	Name             string `yaml:"name"`
	GithubID         string `yaml:"github_id"`
	SSHPublicKeyFile string `yaml:"ssh_public_key_file"`
}

//Backup is a nil function to match the interface
func (o SSH) Backup() error {
	return nil
}

//Restore is a nil function to match the interface
func (o SSH) Restore() error {
	return nil
}

func (o SSH) getFile() [][]string {
	if o.SSH.DefaultSShPubKeyFile != "" {
		return [][]string{
			[]string{SSHFileName, SSHFileName},
			[]string{o.SSH.DefaultSShPubKeyFile, o.SSH.DefaultSShPubKeyFile},
		}
	}
	return [][]string{
		[]string{SSHFileName, SSHFileName},
	}
}

var errorFound bool

// Configure setup for each file entry the github configured ssh keys in the authorized_keys file
func (o SSH) Configure(overwrite bool) error {
	files := o.getFile()
	err := o.DownloadFilesToDirectory(files, o.SSH.TempDir, SSHBucketDir, overwrite)
	if err != nil {
		log.Fatal("error downloading files ", err)
	}

	sshYaml := SSHUsersFile{}
	fileRead, err := ioutil.ReadFile(string(path.Join(o.SSH.TempDir, SSHFileName)))
	if err != nil {
		log.Fatal("no file found to open in "+path.Join(o.SSH.TempDir, SSHFileName), err)
	}
	err = yaml.Unmarshal(fileRead, &sshYaml)
	if err != nil {
		log.Fatal("unable to unmarshal file "+path.Join(o.SSH.TempDir, SSHFileName), err)
	}
	//parse the ssh-user file

	var errorFound bool
	authorizedKeys := &bytes.Buffer{}
	errorFound = false
	for _, user := range sshYaml.Users {
		if user.GithubID != "" {
			authorizedKeys, err = getPublicKeyFromGithub(user, authorizedKeys)
			if err != nil {
				errorFound = true
			}
		}
		if user.SSHPublicKeyFile != "" {
			authorizedKeys, err = getPublicKeyFromFile(user, authorizedKeys)

		}
	}
	if o.SSH.DefaultSShPubKeyFile != "" {
		//write the default_ssh_key in the buffer too
		fileContent, err := ioutil.ReadFile(string(path.Join(o.SSH.TempDir, o.SSH.DefaultSShPubKeyFile)))
		if err != nil {
			log.Fatal(err)
		}
		authorizedKeys.WriteString("#### default_ssh pub_key\n")
		authorizedKeys.WriteString(string(fileContent))
	}

	f, err := os.Create(path.Join(o.SSH.UserDir, SSHAuthorizedKeysTempFileName))
	if err != nil {
		return err
	}
	//write the buffer into the temporary authorized_keys file
	_, err = f.Write([]byte(authorizedKeys.String()))
	if err != nil {
		return err
	}

	//Once finished, copy it to the the real authorized_keys file if everything went ok
	if errorFound {
		log.Fatal("conservative behaviour: error found, skipping the authorized_keys update")
	}

	err = os.Rename(path.Join(o.SSH.UserDir, SSHAuthorizedKeysTempFileName), path.Join(o.SSH.UserDir, SSHAuthorizedKeysFileName))
	if err != nil {
		log.Fatal("error while moving file to authorized_keys: ", err)
	}
	return nil
}

//Init will upload to the configured bucket the ssh file users

func (o SSH) Init(dir string) error {
	files := o.getFile()
	return o.UploadFilesFromDirectory(files, SSHLocalDir, SSHBucketDir)
}

func getPublicKeyFromGithub(userspec UserSpec, authorizedKeys *bytes.Buffer) (*bytes.Buffer, error) {
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	resp, err := client.Get("https://github.com/" + userspec.GithubID + ".keys")
	if err != nil {
		log.Fatal("http protocol error", err)
	}
	if resp.StatusCode == 200 {
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		log.Println("writing ssh keys for user " + userspec.Name)
		authorizedKeys.WriteString("#### " + userspec.GithubID + "\n")
		authorizedKeys.WriteString(buf.String())
	} else {
		log.Println("user " + userspec.GithubID + " not found!")
		authorizedKeys.WriteString("#### " + userspec.Name + "\n")
		authorizedKeys.WriteString("#### no keys for " + userspec.GithubID + "\n")
		return authorizedKeys, fmt.Errorf("error while getting github user")
	}
	return authorizedKeys, nil
}

func getPublicKeyFromFile(userspec UserSpec, authorizedKeys *bytes.Buffer) (*bytes.Buffer, error) {
	fileContent, err := ioutil.ReadFile(string(userspec.SSHPublicKeyFile))
	if err != nil {
		return authorizedKeys, err
	}
	authorizedKeys.WriteString("####" + userspec.Name + "\n")
	authorizedKeys.WriteString(string(fileContent))
	return authorizedKeys, nil
}
