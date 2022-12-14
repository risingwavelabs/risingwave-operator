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
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type FeatureStage string
type FeatureName string

const (
	EnableOpenKruiseFeature FeatureName = "enableOpenKruise"
)

const (
	Alpha FeatureStage = "Alpha"
	Beta  FeatureStage = "Beta"
)

var RisingWaveFeatureManager FeatureManager

var (
	supportedFeatureList = []Feature{
		{
			Name:          "enableOpenKruise",
			Description:   "This feature enables open kruise as an optional provider",
			DefaultEnable: false,
			Stage:         Beta,
		},
	}
)

type Feature struct {
	Name          FeatureName
	Description   string
	Enabled       bool
	DefaultEnable bool
	Stage         FeatureStage
}

// This method return a pointer to a deep copy of a feature struct.
func (f *Feature) DeepCopy() *Feature {
	return &Feature{
		Name:          f.Name,
		Description:   f.Description,
		Enabled:       f.Enabled,
		DefaultEnable: f.DefaultEnable,
		Stage:         f.Stage,
	}
}

// This method returns a pointer to an instance of a Feature struct.
func newFeature(name FeatureName, desc string, defaultEnable bool, stage FeatureStage) *Feature {
	return &Feature{
		Name:          name,
		Description:   desc,
		DefaultEnable: defaultEnable,
		Stage:         stage,
	}
}

type FeatureManager struct {
	// Feature map is an internal structure that stores a feature name and a pointer to a feature struct.
	featureMap map[FeatureName]*Feature
}

// Helper function that returns a pointer to an instance of the FeatureManager.
func newRisingWaveFeatureManager() *FeatureManager {
	return &FeatureManager{
		featureMap: make(map[FeatureName]*Feature),
	}
}

// This functions initializes the FeatureManager with the current supported Features.
func InitFeatureManager(featureGateString string) *FeatureManager {
	RisingWaveFeatureManager = *newRisingWaveFeatureManager()
	for _, supportedFeature := range supportedFeatureList {
		supportedFeature.Enabled = supportedFeature.DefaultEnable
		RisingWaveFeatureManager.addFeature(&supportedFeature)
	}
	RisingWaveFeatureManager.parseFromFeatureGateString(featureGateString)
	return &RisingWaveFeatureManager
}

// This method returns the Feature Manager Struct. Should not be modified after intialization.
func GetFeatureManager() *FeatureManager {
	return &RisingWaveFeatureManager
}

// This is a helper functions that adds a feature to the featureManager, will be used on init to init and add all
// features to the featureManager.
func (m *FeatureManager) addFeature(feature *Feature) {
	m.featureMap[feature.Name] = feature.DeepCopy()
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
	return errors.New(fmt.Sprintf("The following feature does not exist: %s", name))
}

// This method takes in a feature name and checks if it is enabled, returns a bool, nil if it exists,
// and a false,error if it does not exist in the featureManager.
func (m FeatureManager) IsFeatureEnabled(name FeatureName) bool {
	// check for existence of feature in map
	feature, featureExists := m.featureMap[name]
	if featureExists {
		// if feature has been actively set/unset, return enabled, else return deafult value.
		return feature.Enabled
	}
	return false
}

// This metod takes in a feature name and enables it if it exists, if it does not
// it returns an error.
func (m *FeatureManager) EnableFeature(name FeatureName) error {
	// check for existence of feature in map
	return m.setFeatureEnable(name, true)
}

// This metod takes in a feature name and disables it if it exists, if it does not
// it returns an error.
func (m *FeatureManager) DisableFeature(name FeatureName) error {
	// check for existence of feature in map
	return m.setFeatureEnable(name, false)
}

// This method returns the number of features in the featureManager.
func (m FeatureManager) getNumOfFeatures() int {
	return len(m.featureMap)
}

// This method lists all features, returns a copy of the list of feature structs.
func (m FeatureManager) ListFeatures() []Feature {
	var featureList = []Feature{}
	for _, feature := range m.featureMap {
		// make a deep copy of the feature
		curr_feature := *feature
		featureList = append(featureList, curr_feature)
	}
	return featureList
}

// This method lists all enabled features, returns a copy of the list of feature structs.
func (m FeatureManager) ListEnabledFeatures() []Feature {
	var featureList = []Feature{}
	for featureName, feature := range m.featureMap {
		if m.IsFeatureEnabled(featureName) {
			// make a copy of the feature
			curr_feature := *feature
			featureList = append(featureList, curr_feature)
		}
	}
	return featureList
}

// This method lists all disabled features, returns a copy of list of feature structs.
func (m FeatureManager) ListDisabledFeatures() []Feature {
	var featureList = []Feature{}
	for featureName, feature := range m.featureMap {
		if !m.IsFeatureEnabled(featureName) {
			// make a deep copy of the feature
			curr_feature := *feature
			featureList = append(featureList, curr_feature)
		}
	}
	return featureList
}

// This method takes in a feature name and return a copy of the feature struct with all its meta information.
func (m FeatureManager) GetFeature(name FeatureName) (Feature, error) {
	_, featureExists := m.featureMap[name]
	if !featureExists {
		return Feature{}, errors.New(fmt.Sprintf("The following feature does not exist: %s", name))
	}
	// make a deep copy of the feature, every other primitive field is copied implicitly.
	// pointer to bool has to be explicitly copied.
	curr_feature := *m.featureMap[name]
	return curr_feature, nil
}

// This method takes in a feature gate string that is given as a CLI argument,
// parses the features and updates the featureManager. e.g if command line argument is
// --feature-gates=enableOpenKruise=true,otherOption=false, it will set the feature enableOpenKruise
// as true if and only if it exists. if a feature is not supported, it is simply ignored.
func (m *FeatureManager) parseFromFeatureGateString(featureGateString string) error {
	if len(featureGateString) == 0 {
		return nil
	}
	featureGatesArgs := strings.Split(featureGateString, ",")
	for _, featureString := range featureGatesArgs {
		featureName, enabled, err := parseFeatureString(featureString)
		if err != nil {
			return err
		}
		fmt.Println(fmt.Sprintf("%s %t", featureName, enabled))
		_, featureExists := m.featureMap[featureName]
		if !featureExists {
			fmt.Println(fmt.Sprintf("Feature not supported: %s", featureName))
			continue
		}
		m.featureMap[featureName].Enabled = enabled
	}
	return nil
}

// This function parses a feature string into a featurename and a boolean and returns
// an error when a feature string cannot be parsed. e.g enableOpenKruise=true will returb
// (enableOpenKruise, true, nil).
func parseFeatureString(featureString string) (FeatureName, bool, error) {
	featureStringSplit := strings.Split(featureString, "=")
	if len(featureStringSplit) != 2 {
		return "", false, errors.New(fmt.Sprintf("Invalid feature syntax given: %s", featureString))
	}
	featureName := strings.TrimSpace(featureStringSplit[0])
	enabled, err := strconv.ParseBool(strings.TrimSpace(featureStringSplit[1]))
	if err != nil {
		return "", false, errors.New(fmt.Sprintf("Invalid boolean value given: %s", featureString))
	}
	return FeatureName(featureName), enabled, nil
}
