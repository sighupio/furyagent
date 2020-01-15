package component

import (
	"bufio"
	"bytes"
	"log"
	"net/http"
	"os"
	"path"
)

const (
	SSHFileName                   = "ssh-users"
	SSHFilePath                   = "ssh"
	SSHTempPath                   = "/tmp"
	SSHAuthorizedKeysFileName     = "authorized_keys"
	SSHAuthorizedKeysTempFileName = "authorized_keys_tmp"
	SSHUserHomePath               = "/home/ubuntu/.ssh"
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
	return [][]string{
		[]string{SSHFileName, SSHFileName},
	}
}

var errorFound bool

// Configure setup for each file entry the github configured ssh keys in the authorized_keys file
func (o SSH) Configure(overwrite bool) error {
	files := o.getFile()
	return o.DownloadFilesToDirectory(files, o.SSH.SSHDir, SSHTempPath, overwrite)

	file, err := os.Open(path.Join(SSHTempPath, SSHFileName))
	if err != nil {
		log.Fatal("no file found to open in "+path.Join(SSHTempPath, SSHFileName), err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)

	authorizedKeys := bytes.Buffer{}
	errorFound = false
	for scanner.Scan() {
		resp, err := http.Get("https://github.com/" + scanner.Text() + ".keys")
		if err != nil || resp.StatusCode != 404 {
			buf := new(bytes.Buffer)
			buf.ReadFrom(resp.Body)
			log.Println("writing ssh keys for user " + scanner.Text())
			authorizedKeys.WriteString("#### " + scanner.Text() + "\n")
			authorizedKeys.WriteString(buf.String())
			errorFound = true
		} else {
			log.Println("user " + scanner.Text() + " not found!")
			authorizedKeys.WriteString("#### " + scanner.Text() + "\n")
			authorizedKeys.WriteString("#### no keys for " + scanner.Text())
		}
	}
	f, err := os.Create(path.Join(SSHUserHomePath, SSHAuthorizedKeysTempFileName))
	if err != nil {
		return err
	}
	_, err = f.Write([]byte(authorizedKeys.String()))
	if err != nil {
		return err
	}

	//Once finished, copy it to the the real authorized_keys file if everything went ok
	if errorFound {
		log.Fatal("error found during fetch of ssh keys, so i won't write it to authorized_keys")
	}

	err = os.Rename(path.Join(SSHUserHomePath, SSHAuthorizedKeysTempFileName), path.Join(SSHUserHomePath, SSHAuthorizedKeysFileName))
	if err != nil {
		log.Fatal("error while moving file to authorized_keys: ", err)
		return nil
	}
	return nil
}

//Init will upload to the configured bucket the ssh file users
func (o SSH) Init(dir string) error {
	return o.UploadFile(path.Join(SSHFilePath, SSHFileName), SSHFilePath)
}
