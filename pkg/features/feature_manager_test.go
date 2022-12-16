/*
 * Copyright 2022 Singularity Data
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package features

import (
	"fmt"
	"reflect"
	"testing"
)

func newRisingWaveSupportedFeatureListForTest() []Feature {
	return []Feature{
		{
			Name:          "feature-1",
			Description:   "feature-1 desc",
			DefaultEnable: false,
			Enabled:       false,
			Stage:         Beta,
		},
		{
			Name:          "feature-2",
			Description:   "feature-2 desc",
			DefaultEnable: true,
			Enabled:       true,
			Stage:         Alpha,
		},
		{
			Name:          "feature-3",
			Description:   "feature-3 desc",
			DefaultEnable: true,
			Enabled:       true,
			Stage:         Beta,
		},
	}
}

func getNonExistentFeatureName() FeatureName {
	return FeatureName("feature-4")
}
func isFeatureEqual(f1 Feature, f2 Feature) bool {
	return reflect.DeepEqual(f1, f2)
}

func newRisingWaveFeatureManagerForTest(featureGates string) *FeatureManager {
	return InitFeatureManager(newRisingWaveSupportedFeatureListForTest(), featureGates)
}

func TestIsFeatureEnabled(t *testing.T) {
	// Get a fake risingwave feature manager that was initialized with the supported FeatureList.
	fakeRisingWaveFeatureManager := newRisingWaveFeatureManagerForTest("")

	// Test for features in supported list with default true/false value and for features that do not exist.
	testcases := map[string]struct {
		featureName FeatureName
		expected    bool
	}{
		"enabled-feature": {
			featureName: FeatureName("feature-2"),
			expected:    true,
		},
		"disabled-feature": {
			featureName: FeatureName("feature-1"),
			expected:    false,
		},
		"feature-not-exist": {
			featureName: getNonExistentFeatureName(),
			expected:    false,
		},
	}
	for _, tc := range testcases {
		if fakeRisingWaveFeatureManager.IsFeatureEnabled(tc.featureName) != tc.expected {
			t.Fatal("Feature Enabled/Disabled wrongly")
		}
	}
}

func TestEnableFeature(t *testing.T) {
	// Get a fake risingwave feature manager that was initialized with the supported FeatureList.
	fakeRisingWaveFeatureManager := newRisingWaveFeatureManagerForTest("")

	// feature-1 is by default disabled, so enabling it would suffice disabled -> enabled case.
	// feature-2 is by default enabled, so enabling it would suffice for enabled -> enabled case.
	testcases := map[string]struct {
		featureName FeatureName
		exist       bool
		expected    bool
	}{
		"enable-enabled-feature": {
			featureName: FeatureName("feature-2"),
			exist:       true,
		},
		"enable-disabled-feature": {
			featureName: FeatureName("feature-1"),
			exist:       true,
		},
		"feature-not-exist": {
			featureName: getNonExistentFeatureName(),
			exist:       false,
		},
	}

	for _, tc := range testcases {
		err := fakeRisingWaveFeatureManager.EnableFeature(tc.featureName)
		if tc.exist && err != nil {
			t.Fatal("Feature exists but enabling it is throwing an error.")
		}
		if tc.exist && !fakeRisingWaveFeatureManager.IsFeatureEnabled(tc.featureName) {
			t.Fatal("Feature was not enabled.")
		}
	}
}

func TestDisableFeature(t *testing.T) {
	// Get a fake risingwave feature manager that was initialized with the supported FeatureList.
	fakeRisingWaveFeatureManager := newRisingWaveFeatureManagerForTest("")

	// feature-1 is by default disabled, so enabling it would suffice disabled -> disabled case.
	// feature-2 is by default enabled, so enabling it would suffice for enabled -> disabled case.
	testcases := map[string]struct {
		featureName FeatureName
		exist       bool
		expected    bool
	}{
		"disable-enabled-feature": {
			featureName: FeatureName("feature-2"),
			exist:       true,
			expected:    true,
		},
		"enable-disabled-feature": {
			featureName: FeatureName("feature-1"),
			exist:       true,
			expected:    false,
		},
		"feature-not-exist": {
			featureName: getNonExistentFeatureName(),
			exist:       false,
			expected:    false,
		},
	}

	for _, tc := range testcases {
		err := fakeRisingWaveFeatureManager.DisableFeature(tc.featureName)
		if tc.exist && err != nil {
			t.Fatal("Feature exists but disabling it is throwing an error.")
		}
		if tc.exist && fakeRisingWaveFeatureManager.IsFeatureEnabled(tc.featureName) {
			t.Fatal("Feature was not disabled.")
		}
	}

}

func TestGetNumOfFeatures(t *testing.T) {
	// Get a fake risingwave feature manager that was initialized with the supported FeatureList.
	fakeRisingWaveFeatureManager := newRisingWaveFeatureManagerForTest("")

	// Get the actual supported Feature List to test for comparison.
	supportedFeatureList := newRisingWaveSupportedFeatureListForTest()

	if len(supportedFeatureList) != fakeRisingWaveFeatureManager.getNumOfFeatures() {
		t.Fatal("Number of features do not math")
	}
}

func TestListFeatures(t *testing.T) {
	// Get a fake risingwave feature manager that was initialized with the supported FeatureList.
	fakeRisingWaveFeatureManager := newRisingWaveFeatureManagerForTest("")

	// Get the actual supported Feature List to test for comparison.
	supportedFeatureList := newRisingWaveSupportedFeatureListForTest()

	// Make a map of the feature name to index of feature in supported feature list for fast lookup.
	featureListIndexMap := make(map[FeatureName]int)
	for idx, feature := range supportedFeatureList {
		featureListIndexMap[feature.Name] = idx
	}

	// Test for comparison between features
	for _, feature := range fakeRisingWaveFeatureManager.ListFeatures() {
		actualIdx := featureListIndexMap[feature.Name]
		if !isFeatureEqual(feature, supportedFeatureList[actualIdx]) {
			t.Fatal("All features in feature list do not match")
		}
	}
}

func TestListEnabledFeatures(t *testing.T) {
	// Get a fake risingwave feature manager that was initialized with the supported FeatureList.
	fakeRisingWaveFeatureManager := newRisingWaveFeatureManagerForTest("")

	// Get the actual supported Feature List to test for comparison.
	supportedFeatureList := newRisingWaveSupportedFeatureListForTest()

	// Make a map of the feature name to index of feature in supported feature list for fast lookup.
	featureListIndexMap := make(map[FeatureName]int)
	for idx, feature := range supportedFeatureList {
		featureListIndexMap[feature.Name] = idx
	}

	// We test if all features are enabled and also check the previous state in the supportedFeatureList,
	// to check if it was also enabled.
	for _, feature := range fakeRisingWaveFeatureManager.ListEnabledFeatures() {
		if !fakeRisingWaveFeatureManager.IsFeatureEnabled(feature.Name) {
			t.Fatal("Disabled feature present in enabled feature list.")
		}
		if !supportedFeatureList[featureListIndexMap[feature.Name]].Enabled {
			t.Fatal("Feature was previously not enabled.")
		}
	}
}

func TestListDisabledFeatures(t *testing.T) {
	// Get a fake risingwave feature manager that was initialized with the supported FeatureList.
	fakeRisingWaveFeatureManager := newRisingWaveFeatureManagerForTest("")

	// Get the actual supported Feature List to test for comparison.
	supportedFeatureList := newRisingWaveSupportedFeatureListForTest()

	// Make a map of the feature name to index of feature in supported feature list for fast lookup.
	featureListIndexMap := make(map[FeatureName]int)
	for idx, feature := range supportedFeatureList {
		featureListIndexMap[feature.Name] = idx
	}

	// We test if all features are disabled and also check the previous state in the supportedFeatureList,
	// to check if it was also dsiabled.
	for _, feature := range fakeRisingWaveFeatureManager.ListDisabledFeatures() {
		if fakeRisingWaveFeatureManager.IsFeatureEnabled(feature.Name) {
			t.Fatal("Enabled feature present in disabled feature list.")
		}
		if supportedFeatureList[featureListIndexMap[feature.Name]].Enabled {
			t.Fatal("Feature was previously not disabled.")
		}
	}

}

func TestGetFeature(t *testing.T) {
	// Get a fake risingwave feature manager that was initialized with the supported FeatureList.
	fakeRisingWaveFeatureManager := newRisingWaveFeatureManagerForTest("")

	// Get the actual supported Feature List to test for comparison.
	supportedFeatureList := newRisingWaveSupportedFeatureListForTest()

	// Make a map of the feature name to index of feature in supported feature list for fast lookup.
	featureListIndexMap := make(map[FeatureName]int)
	for idx, feature := range supportedFeatureList {
		featureListIndexMap[feature.Name] = idx
	}

	for _, supportedFeature := range supportedFeatureList {
		feature, err := fakeRisingWaveFeatureManager.GetFeature(supportedFeature.Name)
		if err != nil {
			t.Fatal("There was an error in retrieving a supported feature.")
		}
		if !isFeatureEqual(feature, supportedFeature) {
			t.Fatal("Retrieved feature is not equal to expected feature.")
		}
	}

	// We have to test for if a fature does not exist as well.
	_, err := fakeRisingWaveFeatureManager.GetFeature(getNonExistentFeatureName())
	if err == nil {
		t.Fatal("Error not thrown for non existent feature request.")
	}

}

func TestParseFeatureString(t *testing.T) {
	testcases := map[string]struct {
		featureString   string
		expectedName    FeatureName
		expectedEnabled bool
		isError         bool
	}{
		"Valid-feature-string-enable-true": {
			featureString:   "openKruise=true",
			expectedName:    FeatureName("openKruise"),
			expectedEnabled: true,
			isError:         false,
		},
		"Valid-feature-string-disable-false": {
			featureString:   "openKruise=false",
			expectedName:    FeatureName("openKruise"),
			expectedEnabled: false,
			isError:         false,
		},
		"Valid-feature-string-enable-1": {
			featureString:   "openKruise=1",
			expectedName:    FeatureName("openKruise"),
			expectedEnabled: true,
			isError:         false,
		},
		"Valid-feature-string-enable-0": {
			featureString:   "openKruise=0",
			expectedName:    FeatureName("openKruise"),
			expectedEnabled: false,
			isError:         false,
		},
		"Invalid-feature-string-with-2-equals": {
			featureString:   "openKruise=true=",
			expectedName:    FeatureName(""),
			expectedEnabled: false,
			isError:         true,
		},
		"Invalid-feature-string-with-multiple-equals": {
			featureString:   "==openKruise=true=",
			expectedName:    FeatureName(""),
			expectedEnabled: false,
			isError:         true,
		},
		"Invalid-feature-string-with-no-featur-name": {
			featureString:   "=true",
			expectedName:    FeatureName(""),
			expectedEnabled: false,
			isError:         true,
		},
		"Invalid-feature-string-with-no-equals": {
			featureString:   "openKruisetrue",
			expectedName:    FeatureName(""),
			expectedEnabled: false,
			isError:         true,
		},
		"Valid-feature-string-with-spaces": {
			featureString:   "   openKruise=true   ",
			expectedName:    FeatureName("openKruise"),
			expectedEnabled: true,
			isError:         false,
		},
		"Invalid-feature-string-with-invalid-boolean-value": {
			featureString:   "openKruise=yes",
			expectedName:    FeatureName(""),
			expectedEnabled: false,
			isError:         true,
		},
	}

	for _, tc := range testcases {
		featureName, enabled, err := parseFeatureString(tc.featureString)

		// Check for if error should not be thrown but en err is being returned and is not nil.
		if !tc.isError && err != nil {
			t.Fatal("Error was not thrown as intended.")
		}

		// If err is not thrown, we check for equality of feature name and its parsed value.
		if err == nil {
			if featureName != tc.expectedName {
				t.Fatal("Feature name was incorrectly parsed.")
			}
			if enabled != tc.expectedEnabled {
				t.Fatal("Enable value was incorrectly parsed.")
			}
		}
	}
}

func TestParseFromFeatureGateString(t *testing.T) {
	errorTestCases := map[string]string{
		"Invalid-commas":       "enableOpenKruise=true,,",
		"Invalid-equals":       "enableOpenKruise==true,",
		"Empty-enable-value":   "enableOpenKruise=",
		"Invalid-enable-value": "enableOpenKruise=yes,otherFeature=true",
		"Invalid-comma-start":  ",enableOpenKruise=true",
	}
	for tc_name, tc := range errorTestCases {
		fakeRisingWaveFeatureManager := InitFeatureManager(SupportedFeatureList, "")
		if fakeRisingWaveFeatureManager.parseFromFeatureGateString(tc) == nil {
			t.Fatal(fmt.Sprintf("Parse error not thrown for testcase: %s", tc_name))
		}
	}
}
