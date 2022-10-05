/*
Copyright 2021 The Kubernetes Authors.

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

package topologymanager

import (
	"fmt"
	"strings"
	"testing"

	"k8s.io/apimachinery/pkg/util/sets"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	"k8s.io/component-base/featuregate"
	featuregatetesting "k8s.io/component-base/featuregate/testing"
	pkgfeatures "k8s.io/kubernetes/pkg/features"
)

var fancyBetaOption = "fancy-new-option"

type optionAvailTest struct {
	option            string
	featureGate       featuregate.Feature
	featureGateEnable bool
	expectedAvailable bool
}

func TestNewTopologyManagerOptions(t *testing.T) {
	testCases := []struct {
		description     string
		policyOptions   map[string]string
		featureGate     featuregate.Feature
		expectedErr     error
		expectedOptions TopologyManagerOptions
	}{
		{
			description: "return TopologyManagerOptions with PreferClosestNUMA set to true",
			featureGate: pkgfeatures.TopologyManagerPolicyAlphaOptions,
			expectedOptions: TopologyManagerOptions{
				PreferClosestNUMA: true,
			},
			policyOptions: map[string]string{
				PreferClosestNUMA: "true",
			},
		},
		{
			description: "return empty TopologyManagerOptions",
		},
		{
			description: "fail to parse options",
			featureGate: pkgfeatures.TopologyManagerPolicyAlphaOptions,
			policyOptions: map[string]string{
				PreferClosestNUMA: "not a boolean",
			},
			expectedErr: fmt.Errorf("bad value for option"),
		},
		{
			description: "test beta options success",
			featureGate: pkgfeatures.TopologyManagerPolicyBetaOptions,
			policyOptions: map[string]string{
				fancyBetaOption: "true",
			},
		},
		{
			description: "test beta options success",
			policyOptions: map[string]string{
				fancyBetaOption: "true",
			},
			expectedErr: fmt.Errorf("Topology Manager Policy Beta-level Options not enabled,"),
		},
	}

	betaOptions = sets.NewString(fancyBetaOption)

	for _, tcase := range testCases {
		t.Run(tcase.description, func(t *testing.T) {
			if tcase.featureGate != "" {
				defer featuregatetesting.SetFeatureGateDuringTest(t, utilfeature.DefaultFeatureGate, tcase.featureGate, true)()
			}
			opts, err := NewTopologyManagerOptions(tcase.policyOptions)
			if tcase.expectedErr != nil {
				if !strings.Contains(err.Error(), tcase.expectedErr.Error()) {
					t.Errorf("Unexpected error message. Have: %s wants %s", err.Error(), tcase.expectedErr.Error())
				}
			}

			if opts != tcase.expectedOptions {
				t.Errorf("Expected TopologyManagerOptions to equal %v, not %v", tcase.expectedOptions, opts)

			}
		})
	}
}

func TestPolicyDefaultsAvailable(t *testing.T) {
	testCases := []optionAvailTest{
		{
			option:            "this-option-does-not-exist",
			expectedAvailable: false,
		},
		{
			option:            PreferClosestNUMA,
			expectedAvailable: false,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.option, func(t *testing.T) {
			err := CheckPolicyOptionAvailable(testCase.option)
			isEnabled := (err == nil)
			if isEnabled != testCase.expectedAvailable {
				t.Errorf("option %q available got=%v expected=%v", testCase.option, isEnabled, testCase.expectedAvailable)
			}
		})
	}
}

func TestPolicyOptionsAvailable(t *testing.T) {
	testCases := []optionAvailTest{
		{
			option:            "this-option-does-not-exist",
			featureGate:       pkgfeatures.TopologyManagerPolicyBetaOptions,
			featureGateEnable: false,
			expectedAvailable: false,
		},
		{
			option:            "this-option-does-not-exist",
			featureGate:       pkgfeatures.TopologyManagerPolicyBetaOptions,
			featureGateEnable: true,
			expectedAvailable: false,
		},
		{
			option:            PreferClosestNUMA,
			featureGate:       pkgfeatures.TopologyManagerPolicyAlphaOptions,
			featureGateEnable: true,
			expectedAvailable: true,
		},
		{
			option:            PreferClosestNUMA,
			featureGate:       pkgfeatures.TopologyManagerPolicyBetaOptions,
			featureGateEnable: true,
			expectedAvailable: false,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.option, func(t *testing.T) {
			defer featuregatetesting.SetFeatureGateDuringTest(t, utilfeature.DefaultFeatureGate, testCase.featureGate, testCase.featureGateEnable)()
			err := CheckPolicyOptionAvailable(testCase.option)
			isEnabled := (err == nil)
			if isEnabled != testCase.expectedAvailable {
				t.Errorf("option %q available got=%v expected=%v", testCase.option, isEnabled, testCase.expectedAvailable)
			}
		})
	}
}
