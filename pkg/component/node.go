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

// Node represent the object that reflects what nodes need (implements ClusterComponent)
type Node struct {
}

// Backup of a node is Empty
func (n *Node) Backup(cfg *ClusterConfig) error {
	return nil
}

// Restore of a node is Empty
func (n *Node) Restore(cfg *ClusterConfig) error {
	return nil
}

// Configure basicall joins the nodes to the cluster, configures KUBELET_EXTRA_ARGS and restart kubelet and docker in case of necessity
func (n *Node) Configure(cfg *ClusterConfig) error {
	return nil
}
