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
	"log"
	"os/exec"
	"path"
	"time"

	backoff "github.com/cenkalti/backoff/v4"
)

const (
	JoinFile          = "join.sh"
	BucketPath        = "join"
	LocalJoinFilePath = "."
)

// Node represent the object that reflects what nodes need (implements ClusterComponent)
type Node struct {
	ClusterComponentData
	OverWrite bool
}

// Backup of a node is Empty
func (n *Node) Backup() error {
	return nil
}

// Restore of a node is Empty
func (n *Node) Restore() error {
	return nil
}

func (n *Node) getFiles() [][]string {
	return [][]string{
		[]string{JoinFile, JoinFile},
	}
}

// Configure basically joins the nodes to the cluster
func (n *Node) Configure(overwrite bool) error {
	n.OverWrite = overwrite

	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 20 * time.Minute
	b.MaxInterval = 5 * time.Second
	backoff.RetryNotify(n.executeCommand, b, func(err error, t time.Duration) {
		log.Printf("Failed join attempt: %v -> will retry in %s", err, t)
	})
	return nil
}

func (n *Node) executeCommand() error {
	files := n.getFiles()
	err := n.DownloadFilesToDirectory(files, LocalJoinFilePath, BucketPath, n.OverWrite)
	if err != nil {
		return err
	}
	cmd := exec.Command("bash", path.Join(LocalJoinFilePath, JoinFile))
	var output bytes.Buffer
	cmd.Stdout = &output
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}
