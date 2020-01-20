package component

import (
	"bytes"
	"fmt"
	ioutil "io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

const (
	SSHUserSpecs                  = "ssh-users.yml"
	SSHBucketDir                  = "ssh"
	SSHAuthorizedKeysFileName     = "authorized_keys"
	SSHAuthorizedKeysTempFileName = "authorized_keys_tmp"
	SSHSudoerDir                  = "/etc/sudoers.d"
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

type SystemUser struct {
	Name string
	Home string
	Gid  int
	Uid  int
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
	var sysUser *SystemUser
	sysUser, err = createUser(config.User)
	if err != nil {
		log.Fatal("error while creating user", err)
	}

	homeUserSSH := path.Join(sysUser.Home, ".ssh")

	log.Printf("creating temporary authorizedKeys file %s", string(path.Join(homeUserSSH, SSHAuthorizedKeysTempFileName)))
	f, err := os.Create(path.Join(homeUserSSH, SSHAuthorizedKeysTempFileName))
	if err != nil {
		return err
	}
	//write the buffer into the temporary authorized_keys file
	_, err = f.Write([]byte(authorizedKeys.String()))
	if err != nil {
		return err
	}
	err = os.Chown(path.Join(homeUserSSH, SSHAuthorizedKeysTempFileName), sysUser.Uid, sysUser.Gid)
	if err != nil {
		log.Printf("error while changing ownership to file %s", string(path.Join(homeUserSSH, SSHAuthorizedKeysTempFileName)))
	}

	//Once finished, copy it to the the real authorized_keys file if everything went ok
	if errorFound {
		log.Fatal("conservative behaviour: error found, skipping the authorized_keys update")
	}
	log.Printf("everything is fine! Writing temp file %s to its final destination %s", string(path.Join(homeUserSSH, SSHAuthorizedKeysTempFileName)), string(path.Join(homeUserSSH, SSHAuthorizedKeysFileName)))
	err = os.Rename(path.Join(homeUserSSH, SSHAuthorizedKeysTempFileName), path.Join(homeUserSSH, SSHAuthorizedKeysFileName))
	if err != nil {
		log.Fatal("error while moving file to authorized_keys: ", err)
	}
	err = os.Chown(path.Join(homeUserSSH, SSHAuthorizedKeysFileName), sysUser.Uid, sysUser.Gid)
	if err != nil {
		log.Printf("error while changing ownership to file %s", string(path.Join(homeUserSSH, SSHAuthorizedKeysFileName)))
	}
	return nil
}

func GetUidGid(username string) (uid, gid int) {
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
	return uid, gid
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

func getOS() string {
	b, err := ioutil.ReadFile("/etc/os-release")
	if err != nil {
		log.Fatal(err)
	}
	s := strings.Split(string(b), "\n")
	version := ""
	for _, line := range s {
		if bits := strings.Split(line, `=`); len(bits) > 0 {
			if bits[0] == "ID" {
				version = strings.Replace(bits[1], `"`, ``, -1)
			}
		}
	}

	if version != "" {
		return version
	}
	return ""
}

func getAdduserCommand(home, username string) []string {

	switch strings.ToLower(getOS()) {
	case "debian", "ubuntu":
		log.Printf("os identified is %s: ", strings.ToLower(getOS()))
		cmd := []string{"adduser", "--home", home, "--disabled-password", username}
		return cmd
	case "centos", "redhat":
		log.Printf("os identified is %s: ", strings.ToLower(getOS()))
		cmd := []string{"adduser", "-m", "--home-dir", home, username}
		return cmd
	default:
		log.Fatalf("the os %s is not handled", getOS())
	}
	return []string{}
}

func createUser(username string) (*SystemUser, error) {
	userSpec := new(SystemUser)
	var userAlreadyCreated bool
	userAlreadyCreated = false
	if _, err := user.Lookup(username); err == nil {
		log.Printf("the user %s already exists", username)
		userAlreadyCreated = true
	}
	home := path.Join("/home", username)
	log.Println("the user %s is missing, creating it", username)
	if !userAlreadyCreated {
		cmdList := getAdduserCommand(home, username)
		cmd := exec.Command(cmdList[0], cmdList[1:]...)
		log.Println("executing command: ", cmd.String())
		var output bytes.Buffer
		cmd.Stdout = &output
		err := cmd.Run()
		log.Println(output.String())
		if err != nil {
			log.Fatal("error while executing command adduser")
		}
	}
	// create sudoer file for user
	sudoerFile := fmt.Sprintf("99_%s", username)
	sudoerPathname := path.Join(SSHSudoerDir, sudoerFile)
	if !fileExists(sudoerPathname) {
		log.Printf("the sudoer file %s is missing, creating it", sudoerPathname)
		sudoerContent := []byte(fmt.Sprintf("%s ALL=(ALL) NOPASSWD:ALL", username))
		err := ioutil.WriteFile(sudoerPathname, sudoerContent, 0644)
		if err != nil {
			log.Fatal("error while  writing sudoer configuration")
		}
	}
	// create sshdir
	homeUserSSH := path.Join(home, ".ssh")
	uid, gid := GetUidGid(username)
	sshDir := path.Join(homeUserSSH)
	ok, _ := exists(sshDir)
	if !ok {
		log.Printf("the %s is missing, creating it", sshDir)
		os.Mkdir(sshDir, 0755)
		os.Chown(sshDir, uid, gid)
	}

	userSpec.Name = username
	userSpec.Home = home
	userSpec.Gid = gid
	userSpec.Uid = uid

	return userSpec, nil
}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
