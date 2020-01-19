package component

import (
	"bytes"
	"fmt"
	ioutil "io/ioutil"
	"log"
	"net/http"
	"os"
	"os/user"
	"path"
	"strconv"
	"time"

	"gopkg.in/yaml.v2"
)

const (
	SSHUserSpecs                  = "ssh-users.yml"
	SSHBucketDir                  = "ssh"
	SSHAuthorizedKeysFileName     = "authorized_keys"
	SSHAuthorizedKeysTempFileName = "authorized_keys_tmp"
)

type SSHComponent struct {
	ClusterComponentData
}

type SSHUsersFile struct {
	Users []UserSpec `yaml:"users"`
}

type UserSpec struct {
	Name     string `yaml:"name"`
	GithubID string `yaml:"github_id"`
}

//Backup is a nil function to match the interface
func (o SSHComponent) Backup() error {
	return nil
}

//Restore is a nil function to match the interface
func (o SSHComponent) Restore() error {
	return nil
}

func (o SSHComponent) getFiles() [][]string {
	return [][]string{
		[]string{SSHUserSpecs, SSHUserSpecs},
	}
}

var errorFound bool

// Configure setup for each file entry the github configured ssh keys in the authorized_keys file
func (o SSHComponent) Configure(overwrite bool) error {
	files := o.getFiles()
	err := o.DownloadFilesToDirectory(files, o.SSH.TempDir, SSHBucketDir, overwrite)
	if err != nil {
		log.Fatal("error downloading files ", err)
	}
	return sshPubKeys(o.SSH)
}

func sshPubKeys(config SSHConfig) error {
	//parse the ssh-user file
	sshYaml, err := unmarshalSSHUserYaml(config.TempDir, config)
	var errorFound bool
	authorizedKeys := &bytes.Buffer{}
	errorFound = false
	for _, user := range sshYaml.Users {
		if user.GithubID != "" {
			log.Printf("user github found: %s", user.GithubID)
			authorizedKeys, err = getPublicKeyFromGithub(user, authorizedKeys)
			if err != nil {
				log.Println("error found while getting github key for user %s", user.GithubID)
				errorFound = true
			}
		}
	}
	homeUserSsh, uid, gid := GetInfosFromUser(config.User)

	log.Printf("creating temporary authorizedKeys file %s", string(path.Join(homeUserSsh, SSHAuthorizedKeysTempFileName)))
	f, err := os.Create(path.Join(homeUserSsh, SSHAuthorizedKeysTempFileName))
	if err != nil {
		return err
	}
	//write the buffer into the temporary authorized_keys file
	_, err = f.Write([]byte(authorizedKeys.String()))
	if err != nil {
		return err
	}
	err = os.Chown(path.Join(homeUserSsh, SSHAuthorizedKeysTempFileName), uid, gid)
	if err != nil {
		log.Printf("error while changing ownership to file %s", string(path.Join(homeUserSsh, SSHAuthorizedKeysTempFileName)))
	}

	//Once finished, copy it to the the real authorized_keys file if everything went ok
	if errorFound {
		log.Fatal("conservative behaviour: error found, skipping the authorized_keys update")
	}
	log.Printf("everything is fine! Writing temp file %s to its final destination %s", string(path.Join(homeUserSsh, SSHAuthorizedKeysTempFileName)), string(path.Join(homeUserSsh, SSHAuthorizedKeysFileName)))
	err = os.Rename(path.Join(homeUserSsh, SSHAuthorizedKeysTempFileName), path.Join(homeUserSsh, SSHAuthorizedKeysFileName))
	if err != nil {
		log.Fatal("error while moving file to authorized_keys: ", err)
	}
	err = os.Chown(path.Join(homeUserSsh, SSHAuthorizedKeysFileName), uid, gid)
	if err != nil {
		log.Printf("error while changing ownership to file %s", string(path.Join(homeUserSsh, SSHAuthorizedKeysFileName)))
	}
	return nil
}

func GetInfosFromUser(username string) (sshPath string, uid int, gid int) {
	homeUserSsh := path.Join("/home", username, ".ssh")
	userInfo, err := user.Lookup(username)
	if err != nil {
		log.Fatal(err)
	}
	uid, err = strconv.Atoi(userInfo.Uid)
	if err != nil {
		log.Fatal(err)
	}
	gid, err = strconv.Atoi(userInfo.Uid)
	if err != nil {
		log.Fatal(err)
	}
	return homeUserSsh, uid, gid
}

//Init will upload to the configured bucket the ssh file users

func (o SSHComponent) Init(dir string) error {
	files := o.getFiles()
	return o.UploadFilesFromDirectoryWithForce(files, o.SSH.LocalDirConfigs, SSHBucketDir)
}

func unmarshalSSHUserYaml(dirPath string, config SSHConfig) (SSHUsersFile, error) {
	sshYaml := SSHUsersFile{}
	log.Printf("loading spec file %s", string(path.Join(dirPath, SSHUserSpecs)))
	fileRead, err := ioutil.ReadFile(string(path.Join(dirPath, SSHUserSpecs)))
	if err != nil {
		log.Fatal("no file found to open in "+path.Join(dirPath, SSHUserSpecs), err)
	}
	err = yaml.Unmarshal(fileRead, &sshYaml)
	if err != nil {
		log.Fatal("unable to unmarshal file "+path.Join(dirPath, SSHUserSpecs), err)
	}
	return sshYaml, nil
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

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
