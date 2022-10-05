/*
Copyright 2022 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package numatopology

import (
	"fmt"
	"strconv"
	"strings"

	cadvisorapi "github.com/google/cadvisor/info/v1"

	"k8s.io/kubernetes/pkg/kubelet/cm/topologymanager/bitmask"
)

type Topology struct {
	numaNodes []int
	distances [][]uint64
}

func New(topology []cadvisorapi.Node, sysFs SysFs) (Topology, error) {
	var numaNodes []int
	distances := make([][]uint64, len(topology))
	for _, node := range topology {
		numaNodes = append(numaNodes, node.Id)

		// uncomment when this commit is released https://github.com/google/cadvisor/commit/24dd1de08a72cfee661f6178454db995900c0fee
		// distances[node.Id] = append(distances[node.Id], node.Distances...)

		// for now we need to retrieve this information in Kubelet
		nodeDistance, err := getDistancesForNode(sysFs, node.Id)
		if err != nil {
			return Topology{}, err
		}

		distances[node.Id] = nodeDistance

	}

	return Topology{
		numaNodes: numaNodes,
		distances: distances,
	}, nil
}

func getDistancesForNode(sysFs SysFs, nodeId int) ([]uint64, error) {
	rawDistances, err := sysFs.GetDistances(nodeId)
	if err != nil {
		return nil, err
	}

	distances := []uint64{}
	for _, distance := range strings.Split(rawDistances, " ") {
		distanceUint, err := strconv.ParseUint(distance, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("cannot convert %s to int", distance)
		}
		distances = append(distances, distanceUint)
	}
	return distances, nil
}

func (n Topology) GetNumaNodes() []int {
	return n.numaNodes
}

func (n Topology) NumaNodesCount() int {
	return len(n.numaNodes)
}

func (n Topology) CalculateAvgDistanceFor(bm bitmask.BitMask) float64 {
	var count float64 = 0
	var sum float64 = 0
	for _, node := range bm.GetBits() {
		for _, distance := range bm.GetBits() {
			sum += float64(n.distances[node][distance])
			count++
		}
	}

	return sum / count
}
