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
	"strconv"
	"strings"
	"unicode"
)

// FeatureStage is the stage of features, e.g., alpha, beta, GA. See Valid feature stages below.
type FeatureStage string

// FeatureName is an alias of the string.
type FeatureName string

// Valid feature names.
const (
	EnableOpenKruiseFeature     FeatureName = "EnableOpenKruise"
	EnableForceUpdate           FeatureName = "EnableForceUpdate"
	RandomSecretStorePrivateKey FeatureName = "RandomSecretStorePrivateKey"
)

// Valid feature stages.
const (
	Alpha FeatureStage = "Alpha"
	Beta  FeatureStage = "Beta"
)

var risingWaveFeatureManager *FeatureManager

var (
	// SupportedFeatureList is the global and constant supported feature list.
	SupportedFeatureList = []Feature{
		{
			Name:          EnableOpenKruiseFeature,
			Description:   "This feature enables open kruise as an optional provider",
			DefaultEnable: false,
			Stage:         Beta,
		},
		{
			Name:          EnableForceUpdate,
			Description:   "This feature enables force resolve version conflict due to operator update",
			DefaultEnable: true,
			Stage:         Beta,
		},
		{
			Name:          RandomSecretStorePrivateKey,
			Description:   "This feature enables the random generation of a secret store private key if it is not set",
			DefaultEnable: false,
			Enabled:       false,
			Stage:         Alpha,
		},
	}
)

// Feature defines a feature and its status.
type Feature struct {
	Name          FeatureName
	Description   string
	Enabled       bool
	DefaultEnable bool
	Stage         FeatureStage
}

// DeepCopy returns a pointer to a deep copy of a feature struct.
func (f *Feature) DeepCopy() *Feature {
	return &Feature{
		Name:          f.Name,
		Description:   f.Description,
		Enabled:       f.Enabled,
		DefaultEnable: f.DefaultEnable,
		Stage:         f.Stage,
	}
}

// FeatureManager is the manager of operator features.
type FeatureManager struct {
	// Feature map is an internal structure that stores a feature name and a pointer to a feature struct.
	featureMap map[FeatureName]*Feature
}

// NewRisingWaveFeatureManager is a helper function that returns a pointer to an instance of the FeatureManager.
func NewRisingWaveFeatureManager() *FeatureManager {
	return &FeatureManager{
		featureMap: make(map[FeatureName]*Feature),
	}
}

// InitFeatureManagerWithSupportedFeatures initializes the FeatureManager with the current supported Features.
func InitFeatureManagerWithSupportedFeatures(supportedFeatureList []Feature) *FeatureManager {
	risingWaveFeatureManager = NewRisingWaveFeatureManager()

	for _, supportedFeature := range supportedFeatureList {
		supportedFeature.Enabled = supportedFeature.DefaultEnable
		risingWaveFeatureManager.addFeature(&supportedFeature)
	}

	return risingWaveFeatureManager
}

// InitFeatureManager initializes the FeatureManager with the current supported Features and also parses the feature gate string.
func InitFeatureManager(supportedFeatureList []Feature, featureGateString string) *FeatureManager {
	risingWaveFeatureManager = NewRisingWaveFeatureManager()

	for _, supportedFeature := range supportedFeatureList {
		supportedFeature.Enabled = supportedFeature.DefaultEnable
		risingWaveFeatureManager.addFeature(&supportedFeature)
	}

	if risingWaveFeatureManager.ParseFromFeatureGateString(featureGateString) != nil {
		panic("Invalid value given to feature-gates argument")
	}

	return risingWaveFeatureManager
}

// GetFeatureManager returns the Feature Manager Struct. Should not be modified after initialization.
func GetFeatureManager() *FeatureManager {
	return risingWaveFeatureManager
}

// This is a helper functions that adds a feature to the featureManager, will be used on init to init and add all
// features to the featureManager.
func (m *FeatureManager) addFeature(feature *Feature) {
	m.featureMap[feature.Name] = feature.DeepCopy()
}

// IsFeatureExist returns true if the feature exists in the featureManager, else returns false.
func (m *FeatureManager) IsFeatureExist(featureName FeatureName) bool {
	_, exist := m.featureMap[featureName]

	return exist
}

// This is a helper function that helps to set a feature to a given boolean, if feature does not exist,
// returns an error. Used in EnableFeature and DisableFeature methods. Not visible to users.
func (m *FeatureManager) setFeatureEnable(name FeatureName, enable bool) error {
	// check for existence of feature in map
	_, featureExists := m.featureMap[name]
	if featureExists {
		m.featureMap[name].Enabled = enable

		return nil
	}

	return fmt.Errorf("the following feature does not exist: %s", name)
}

// IsFeatureEnabled takes in a feature name and checks if it is enabled, returns a bool, nil if it exists,
// and a false,error if it does not exist in the featureManager.
func (m *FeatureManager) IsFeatureEnabled(name FeatureName) bool {
	// check for existence of feature in map
	feature, featureExists := m.featureMap[name]

	return featureExists && feature.Enabled
}

// EnableFeature takes in a feature name and enables it if it exists, if it does not
// it returns an error.
func (m *FeatureManager) EnableFeature(name FeatureName) error {
	// check for existence of feature in map
	return m.setFeatureEnable(name, true)
}

// DisableFeature takes in a feature name and disables it if it exists, if it does not
// it returns an error.
func (m *FeatureManager) DisableFeature(name FeatureName) error {
	// check for existence of feature in map
	return m.setFeatureEnable(name, false)
}

// GetNumOfFeatures returns the number of features in the featureManager.
func (m *FeatureManager) GetNumOfFeatures() int {
	return len(m.featureMap)
}

// ListFeatures lists all features, returns a copy of the list of feature structs.
func (m *FeatureManager) ListFeatures() []Feature {
	featureList := make([]Feature, 0, len(m.featureMap))
	for _, feature := range m.featureMap {
		// make a deep copy of the feature
		featureList = append(featureList, *feature.DeepCopy())
	}

	return featureList
}

// ListEnabledFeatures lists all enabled features, returns a copy of the list of feature structs.
func (m *FeatureManager) ListEnabledFeatures() []Feature {
	var featureList []Feature

	for _, feature := range m.featureMap {
		if feature.Enabled {
			// make a copy of the feature
			featureList = append(featureList, *feature.DeepCopy())
		}
	}

	return featureList
}

// ListDisabledFeatures lists all disabled features, returns a copy of list of feature structs.
func (m *FeatureManager) ListDisabledFeatures() []Feature {
	var featureList []Feature

	for _, feature := range m.featureMap {
		if !feature.Enabled {
			// make a deep copy of the feature
			featureList = append(featureList, *feature.DeepCopy())
		}
	}

	return featureList
}

// GetFeature takes in a feature name and return a copy of the feature struct with all its meta information.
func (m *FeatureManager) GetFeature(name FeatureName) (Feature, error) {
	_, featureExists := m.featureMap[name]
	if !featureExists {
		return Feature{}, fmt.Errorf("the following feature does not exist: %s", name)
	}
	// make a deep copy of the feature, every other primitive field is copied implicitly.
	return *m.featureMap[name].DeepCopy(), nil
}

// ParseFromFeatureGateString takes in a feature gate string that is given as a CLI argument,
// parses the features and updates the featureManager. e.g if command line argument is
// --feature-gates=enableOpenKruise=true,otherOption=false, it will set the feature enableOpenKruise
// as true if and only if it exists. if a feature is not supported, it is simply ignored.
func (m *FeatureManager) ParseFromFeatureGateString(featureGateString string) error {
	if len(featureGateString) == 0 {
		return nil
	}

	featureGateString = strings.TrimSpace(featureGateString)
	if !unicode.IsLetter([]rune(featureGateString[0:1])[0]) {
		return fmt.Errorf("parsing error of feature gate string: %s", featureGateString)
	}

	featureGatesArgs := strings.Split(featureGateString, ",")
	for _, featureString := range featureGatesArgs {
		featureName, enabled, err := parseFeatureString(featureString)
		if err != nil {
			return err
		}

		_, featureExists := m.featureMap[featureName]
		if !featureExists {
			continue
		}

		m.featureMap[featureName].Enabled = enabled
	}

	return nil
}

// parseFeatureString parses a feature string into a FeatureName and a boolean and returns
// an error when a feature string cannot be parsed. e.g, enableOpenKruise=true will return
// (enableOpenKruise, true, nil).
func parseFeatureString(featureString string) (FeatureName, bool, error) {
	featureStringSplit := strings.Split(featureString, "=")
	if len(featureStringSplit) != 2 {
		return "", false, fmt.Errorf("invalid feature syntax given: %s", featureString)
	}

	featureName := strings.TrimSpace(featureStringSplit[0])
	if len(featureName) == 0 {
		return "", false, fmt.Errorf("invalid feature name given: %s", featureName)
	}

	enabled, err := strconv.ParseBool(strings.TrimSpace(featureStringSplit[1]))
	if err != nil {
		return "", false, fmt.Errorf("invalid feature status: %s, parse failed: %w", featureString, err)
	}

	return FeatureName(featureName), enabled, nil
}
