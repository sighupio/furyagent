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
	b.MaxElapsedTime = n.Node.RetryMaxMin * time.Minute
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
	cmd := exec.Command("bash", path.Join(LocalJoinFilePath, JoinFile))
	var output bytes.Buffer
	cmd.Stdout = &output
	err = cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

//Init is for interface compliance, now is empty
func (n Node) Init(s string) error {
	return nil
}
