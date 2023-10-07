/*
 * Copyright 2023 RisingWave Labs
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
	"strings"
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
	return "feature-4"
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

	if len(supportedFeatureList) != fakeRisingWaveFeatureManager.GetNumOfFeatures() {
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
	// to check if it was also disabled.
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

	// We have to test for if a feature does not exist as well.
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
		"Invalid-feature-string-with-no-feature-name": {
			featureString:   "=true",
			expectedName:    FeatureName(""),
			expectedEnabled: false,
			isError:         true,
		},
		"Invalid-feature-string-with-no-equals": {
			// spellchecker: disable
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

	// Check for cases where parsing would fail first.
	errorTestCases := map[string]string{
		"Invalid-commas":       "enableOpenKruise=true,,",
		"Invalid-equals":       "enableOpenKruise==true,",
		"Empty-enable-value":   "enableOpenKruise=",
		"Invalid-enable-value": "enableOpenKruise=yes,otherFeature=true",
		"Invalid-comma-start":  ",enableOpenKruise=true",
	}
	for name, tc := range errorTestCases {
		fakeRisingWaveFeatureManager := NewRisingWaveFeatureManager()
		if fakeRisingWaveFeatureManager.ParseFromFeatureGateString(tc) == nil {
			t.Fatalf("Parse error not thrown for testcase: %s", name)
		}
	}

	var fakeRisingWaveFeatureManager *FeatureManager

	// Check for when feature gate string only contains features that are supported,
	// we test by enabling all of them, in our featureGate string and check if all are enabled in list of features.
	AllFeatureEnabledString := "feature-1=true,feature-2=TRUE,feature-3=1"
	fakeRisingWaveFeatureManager = InitFeatureManagerWithSupportedFeatures(newRisingWaveSupportedFeatureListForTest())
	if err := fakeRisingWaveFeatureManager.ParseFromFeatureGateString(AllFeatureEnabledString); err != nil {
		t.Fatal(err)
	}
	if fakeRisingWaveFeatureManager.GetNumOfFeatures() != len(strings.Split(AllFeatureEnabledString, ",")) {
		t.Fatal("Error in parsing the correct number of features")
	}
	for _, feature := range fakeRisingWaveFeatureManager.ListFeatures() {
		if !fakeRisingWaveFeatureManager.IsFeatureEnabled(feature.Name) {
			t.Fatal("Parsing has failed, not all features were enabled")
		}
	}

	// Check for when feature gate string only contains features that are supported,
	// we test by disabling all of them, in our featureGate string and check if all are disabled in list of features.
	AllFeatureDisabledString := "feature-1=false,feature-2=FALSE,feature-3=0"
	fakeRisingWaveFeatureManager = InitFeatureManagerWithSupportedFeatures(newRisingWaveSupportedFeatureListForTest())
	if err := fakeRisingWaveFeatureManager.ParseFromFeatureGateString(AllFeatureDisabledString); err != nil {
		t.Fatal(err)
	}
	if fakeRisingWaveFeatureManager.GetNumOfFeatures() != len(strings.Split(AllFeatureDisabledString, ",")) {
		t.Fatal("Error in parsing the correct number of features")
	}
	for _, feature := range fakeRisingWaveFeatureManager.ListFeatures() {
		if fakeRisingWaveFeatureManager.IsFeatureEnabled(feature.Name) {
			t.Fatal("Parsing has failed, not all features were disabled")
		}
	}

	// Check for when feature gate string also contains features that are not supported.
	// we test by enabling all of them, in our featureGate string and check if all are enabled in list of features.
	AllFeatureEnabledStringWithUnsupportedFeatures := fmt.Sprintf("feature-1=true,feature-2=TRUE,feature-3=1,%s=True", getNonExistentFeatureName())
	fakeRisingWaveFeatureManager = InitFeatureManagerWithSupportedFeatures(newRisingWaveSupportedFeatureListForTest())
	if err := fakeRisingWaveFeatureManager.ParseFromFeatureGateString(AllFeatureEnabledStringWithUnsupportedFeatures); err != nil {
		t.Fatal(err)
	}
	for _, feature := range fakeRisingWaveFeatureManager.ListFeatures() {
		if !fakeRisingWaveFeatureManager.IsFeatureEnabled(feature.Name) {
			t.Fatal("Parsing has failed, not all features were enabled")
		}
	}
}
