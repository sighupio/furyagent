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

package component

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	backoff "github.com/cenkalti/backoff/v4"
)

const (
	JoinFile                 string        = "join.sh"
	BucketPath               string        = "join"
	LocalJoinFilePath        string        = "."
	DefaultJointTimeoutValue time.Duration = 30
)

// Node represent the object that reflects what nodes need (implements ClusterComponent)
type Node struct {
	ClusterComponentData
}

// Backup of a node is Empty
func (n Node) Backup() error {
	return nil
}

// Restore of a node is Empty
func (n Node) Restore() error {
	return nil
}

func (n Node) getFiles() [][]string {
	return [][]string{
		[]string{JoinFile, JoinFile},
	}
}

type BackoffNode struct {
	Node
	OverWrite bool
}

// Configure basically joins the nodes to the cluster
func (n Node) Configure(overwrite bool) error {
	bn := newBackoffNode(n, overwrite)
	b := backoff.NewExponentialBackOff()
	joinTimeout := getJoinTimeout(n.Node.joinTimeout)
	b.MaxElapsedTime = joinTimeout * time.Minute
	b.MaxInterval = 5 * time.Second
	notify := func(err error, t time.Duration) {
		log.Printf("Failed join attempt: %v -> will retry in %s", err, t)
	}
	err := backoff.RetryNotify(bn.executeCommand, b, notify)
	if err != nil {
		log.Fatalf("join command exit abnormally after %v of retry with error: %v", b.MaxElapsedTime, err)
	}
	return nil
}

func getJoinTimeout(joinTimeout time.Duration) time.Duration {
	if int(joinTimeout) == 0 {
		// set default value to 30 min if no values are passed from yaml
		return DefaultJointTimeoutValue
	} else {
		return joinTimeout
	}
}

func newBackoffNode(node Node, overwrite bool) *BackoffNode {
	return &BackoffNode{
		Node:      node,
		OverWrite: overwrite,
	}
}

// executeCommand must be a function of type Operation.v4 for backoff
func (b BackoffNode) executeCommand() error {
	files := b.Node.getFiles()
	err := b.Node.DownloadFilesToDirectory(files, LocalJoinFilePath, BucketPath, b.OverWrite)
	if err != nil {
		return err
	}

	err = addNodeName(path.Join(LocalJoinFilePath, JoinFile))
	if err != nil {
		return err
	}

	cmd := exec.Command("bash", path.Join(LocalJoinFilePath, JoinFile))
	output := new(bytes.Buffer)
	cmd.Stdout = output
	cmd.Stderr = output
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("error: %v, output: %s", err, output.String())
	}
	return nil
}

//Init is for interface compliance, now is empty
func (n Node) Init(s string) error {
	return nil
}

func addNodeName(file string) error {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}
	fqdn, _ := getHostnameFqdn()
	nodename := fmt.Sprintf(" --node-name=%s", fqdn)
	newcontent := content
	if !strings.Contains(string(content), nodename) {
		newcontent = append(bytes.Trim(content, "\n"), nodename...)
	}
	newfile, err := os.OpenFile(file, os.O_WRONLY, 0644)
	_, err = newfile.Write(newcontent)
	newfile.Close()
	content, err = ioutil.ReadFile(file)
	return nil
}

func getHostnameFqdn() (string, error) {
	cmd := exec.Command("/bin/hostname", "-f")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("Error when get_hostname_fqdn: %v", err)
	}
	fqdn := out.String()
	fqdn = fqdn[:len(fqdn)-1] // removing EOL

	return fqdn, nil
}
