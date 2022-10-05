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

import "testing"

func TestGetDistances(t *testing.T) {
	testCases := []struct {
		name              string
		nodeDir           string
		expectedErr       bool
		expectedDistances string
		nodeId            int
	}{
		{
			name:              "reading proper distance file",
			nodeDir:           "testdata",
			expectedErr:       false,
			expectedDistances: "10 11 12 13",
			nodeId:            0,
		},
		{
			name:              "no distance file",
			nodeDir:           "testdata",
			expectedErr:       true,
			expectedDistances: "",
			nodeId:            99,
		},
	}

	for _, tcase := range testCases {
		nodeDir = tcase.nodeDir
		sysFs := NewRealSysFs()

		distances, err := sysFs.GetDistances(tcase.nodeId)
		if !tcase.expectedErr && err != nil {
			t.Fatalf("Expected err to equal nil, not %v", err)
		} else if tcase.expectedErr && err == nil {
			t.Fatalf("Expected err to equal %v, not nil", tcase.expectedErr)
		}

		if !tcase.expectedErr && distances != tcase.expectedDistances {
			t.Fatalf("Expected distances to equal %s, not %s", tcase.expectedDistances, distances)
		}
	}
}
