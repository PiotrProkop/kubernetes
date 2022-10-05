/*
Copyright 2019 The Kubernetes Authors.

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
	"reflect"
	"testing"

	cadvisorapi "github.com/google/cadvisor/info/v1"

	"k8s.io/kubernetes/pkg/kubelet/cm/topologymanager/bitmask"
	"k8s.io/kubernetes/pkg/kubelet/cm/topologymanager/numatopology/fakesysfs"
)

func TestNUMATopology(t *testing.T) {
	tcases := []struct {
		name             string
		topology         []cadvisorapi.Node
		expectedTopology Topology
		expectedErr      error
	}{
		{
			name: "positive test 1 node",
			topology: []cadvisorapi.Node{
				{
					Id: 0,
				},
			},
			expectedTopology: Topology{
				numaNodes: []int{0},
				distances: [][]uint64{
					{
						10,
						11,
						12,
						12,
					},
				},
			},
		},
		{
			name: "positive test 2 nodes",
			topology: []cadvisorapi.Node{
				{
					Id: 0,
				},
				{
					Id: 1,
				},
			},
			expectedTopology: Topology{
				numaNodes: []int{0, 1},
				distances: [][]uint64{
					{
						10,
						11,
						12,
						12,
					},
					{
						11,
						10,
						12,
						12,
					},
				},
			},
		},
		{
			name: "positive test 3 nodes",
			topology: []cadvisorapi.Node{
				{
					Id: 0,
				},
				{
					Id: 1,
				},
				{
					Id: 2,
				},
			},
			expectedTopology: Topology{
				numaNodes: []int{0, 1, 2},
				distances: [][]uint64{
					{
						10,
						11,
						12,
						12,
					},
					{
						11,
						10,
						12,
						12,
					},
					{
						12,
						12,
						10,
						11,
					},
				},
			},
		},
		{
			name: "positive test 4 nodes",
			topology: []cadvisorapi.Node{
				{
					Id: 0,
				},
				{
					Id: 1,
				},
				{
					Id: 2,
				},
				{
					Id: 3,
				},
			},
			expectedTopology: Topology{
				numaNodes: []int{0, 1, 2, 3},
				distances: [][]uint64{
					{
						10,
						11,
						12,
						12,
					},
					{
						11,
						10,
						12,
						12,
					},
					{
						12,
						12,
						10,
						11,
					},
					{
						12,
						12,
						11,
						10,
					},
				},
			},
		},
		{
			name: "negative test 1 node",
			topology: []cadvisorapi.Node{
				{
					Id: 0,
				},
			},
			expectedTopology: Topology{},
			expectedErr:      fmt.Errorf("no distance file found"),
		},
	}

	fakeSysFs := fakesysfs.FakeSysFs{}
	fakeSysFs.SetDistances(0, "10 11 12 12", nil)
	fakeSysFs.SetDistances(1, "11 10 12 12", nil)
	fakeSysFs.SetDistances(2, "12 12 10 11", nil)
	fakeSysFs.SetDistances(3, "12 12 11 10", nil)

	for _, tcase := range tcases {
		if tcase.expectedErr != nil {
			fakeSysFs.SetDistances(99, "", tcase.expectedErr)
		}

		topology, err := New(tcase.topology, &fakeSysFs)
		if tcase.expectedErr == nil && err != nil {
			t.Fatalf("Expected err to equal nil, not %v", err)
		} else if tcase.expectedErr != nil && err == nil {
			t.Fatalf("Expected err to equal %v, not nil", tcase.expectedErr)
		} else if tcase.expectedErr != nil && err != tcase.expectedErr {
			t.Fatalf("Expected err to equal %v, not %v", tcase.expectedErr, err)
		}

		if !reflect.DeepEqual(topology, tcase.expectedTopology) {
			t.Fatalf("Expected topology to equal %v, not %v", tcase.expectedTopology, topology)
		}

	}
}

func TestCalculateAvgDistanceFor(t *testing.T) {
	tcases := []struct {
		name        string
		bm          []int
		distance    [][]uint64
		expectedAvg float64
	}{
		{
			name: "1 NUMA node",
			bm: []int{
				0,
			},
			distance: [][]uint64{
				{
					10,
				},
			},
			expectedAvg: 10,
		},
		{
			name: "2 NUMA node, 1 set in bitmask",
			bm: []int{
				0,
			},
			distance: [][]uint64{
				{
					10,
					11,
				},
				{
					11,
					10,
				},
			},
			expectedAvg: 10,
		},
		{
			name: "2 NUMA node, 2 set in bitmask",
			bm: []int{
				0,
				1,
			},
			distance: [][]uint64{
				{
					10,
					11,
				},
				{
					11,
					10,
				},
			},
			expectedAvg: 10.5,
		},
		{
			name: "4 NUMA node, 2 set in bitmask",
			bm: []int{
				0,
				2,
			},
			distance: [][]uint64{
				{
					10,
					11,
					12,
					12,
				},
				{
					11,
					10,
					12,
					12,
				},
				{
					12,
					12,
					10,
					11,
				},
				{
					12,
					12,
					11,
					10,
				},
			},
			expectedAvg: 11,
		},
		{
			name: "4 NUMA node, 3 set in bitmask",
			bm: []int{
				0,
				2,
				3,
			},
			distance: [][]uint64{
				{
					10,
					11,
					12,
					12,
				},
				{
					11,
					10,
					12,
					12,
				},
				{
					12,
					12,
					10,
					11,
				},
				{
					12,
					12,
					11,
					10,
				},
			},
			expectedAvg: 11.11111111111111,
		},
	}

	for _, tcase := range tcases {
		bm, err := bitmask.NewBitMask(tcase.bm...)
		if err != nil {
			t.Errorf("no error expected got %v", err)
		}

		topo := Topology{
			numaNodes: tcase.bm,
			distances: tcase.distance,
		}

		result := topo.CalculateAvgDistanceFor(bm)
		if result != tcase.expectedAvg {
			t.Errorf("Expected result to equal %g, not %g", tcase.expectedAvg, result)
		}
	}

}
