package component

import (
	"bufio"
	"bytes"
	ioutil "io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"time"
)

const (
	SSHFileName                   = "ssh-users"
	SSHBucketDir                  = "ssh"
	SSHLocalDir                   = "secrets/ssh"
	SSHAuthorizedKeysFileName     = "authorized_keys"
	SSHAuthorizedKeysTempFileName = "authorized_keys_tmp"
)

type SSH struct {
	ClusterComponentData
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

	file, err := os.Open(path.Join(o.SSH.TempDir, SSHFileName))
	if err != nil {
		log.Fatal("no file found to open in "+path.Join(o.SSH.TempDir, SSHFileName), err)
	}
	defer file.Close()

	// parse the ssh-user file
	scanner := bufio.NewScanner(file)

	authorizedKeys := bytes.Buffer{}
	errorFound = false
	for scanner.Scan() {
		client := http.Client{
			Timeout: 5 * time.Second,
		}
		resp, err := client.Get("https://github.com/" + scanner.Text() + ".keys")
		if err != nil {
			log.Fatal("http protocol error", err)
		}
		if resp.StatusCode == 200 {
			buf := new(bytes.Buffer)
			buf.ReadFrom(resp.Body)
			log.Println("writing ssh keys for user " + scanner.Text())
			authorizedKeys.WriteString("#### " + scanner.Text() + "\n")
			authorizedKeys.WriteString(buf.String())
		} else {
			log.Println("user " + scanner.Text() + " not found!")
			authorizedKeys.WriteString("#### " + scanner.Text() + "\n")
			authorizedKeys.WriteString("#### no keys for " + scanner.Text() + "\n")
			errorFound = true
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
