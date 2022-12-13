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

package utils

import (
	"errors"
	"fmt"
	"strings"

	"k8s.io/utils/pointer"
)

type FeatureStage string

const (
	Alpha FeatureStage = "Alpha"
	Beta  FeatureStage = "Beta"
)

type Feature struct {
	name          string
	description   string
	enabled       *bool
	defaultEnable bool
	stage         FeatureStage
}

func newFeature(name string, desc string, defaultEnable bool, stage FeatureStage) *Feature {
	return &Feature{
		name:          name,
		description:   desc,
		defaultEnable: defaultEnable,
		stage:         stage,
	}
}

type FeatureManager struct {
	feature_map map[string]*Feature
}

func NewFeatureManager() *FeatureManager {
	return &FeatureManager{
		feature_map: make(map[string]*Feature),
	}
}

func (m *FeatureManager) InitFeatureManager() *FeatureManager {
	m.addFeature("enableOpenKruise", "This feature provides open kruise as an optional provider", false, Alpha)
	return m
}

func (m *FeatureManager) addFeature(name string, desc string, defaultEnable bool, stage FeatureStage) {
	m.feature_map[name] = newFeature(name, desc, defaultEnable, stage)
}

// not visible to users
func (m *FeatureManager) setFeatureEnable(name string, enable bool) error {
	// check for existence of feature in map
	_, featureExists := m.feature_map[name]
	if featureExists {
		m.feature_map[name].enabled = pointer.Bool(enable)
		return nil
	}
	return errors.New(fmt.Sprintf("The following feature does not exist: %s", name))
}

func (m FeatureManager) IsFeatureEnabled(name string) (bool, error) {
	// check for existence of feature in map
	_, featureExists := m.feature_map[name]
	if featureExists {
		// if feature has been actively set/unset, return enabled, else return deafult value.
		return pointer.BoolDeref(m.feature_map[name].enabled, m.feature_map[name].defaultEnable), nil
	}
	return false, errors.New(fmt.Sprintf("The following feature does not exist: %s", name))
}

func (m *FeatureManager) EnableFeature(name string) error {
	// check for existence of feature in map
	return m.setFeatureEnable(name, true)
}

func (m *FeatureManager) DisableFeature(name string) error {
	// check for existence of feature in map
	return m.setFeatureEnable(name, false)
}

func (m FeatureManager) ListFeatures() []Feature {
	var featureList = []Feature{}
	for _, feature := range m.feature_map {
		// make a deep copy of the feature
		curr_feature := *feature
		// if pointer to enabled  has been set, make a copy and set a pointer to enabled.
		if curr_feature.enabled != nil {
			curr_feature.enabled = pointer.Bool(*curr_feature.enabled)
		}
		featureList = append(featureList, curr_feature)
	}
	return featureList
}

func (m FeatureManager) ListEnabledFeatures() []Feature {
	var featureList = []Feature{}
	for featureName, feature := range m.feature_map {
		isFeatureEnabled, _ := m.IsFeatureEnabled(featureName)
		if isFeatureEnabled {
			// make a copy of the feature
			curr_feature := *feature
			// if pointer to enabled  has been set, make a copy and set a pointer to enabled.
			if curr_feature.enabled != nil {
				curr_feature.enabled = pointer.Bool(*curr_feature.enabled)
			}
			featureList = append(featureList, curr_feature)
		}
	}
	return featureList
}

func (m FeatureManager) ListDisabledFeatures() []Feature {
	var featureList = []Feature{}
	for featureName, feature := range m.feature_map {
		isFeatureEnabled, _ := m.IsFeatureEnabled(featureName)
		if !isFeatureEnabled {
			// make a deep copy of the feature
			curr_feature := *feature
			// if pointer to enabled  has been set, make a copy and set a pointer to enabled.
			if curr_feature.enabled != nil {
				curr_feature.enabled = pointer.Bool(*curr_feature.enabled)
			}
			featureList = append(featureList, curr_feature)
		}
	}
	return featureList
}

func (m FeatureManager) GetFeature(name string) (*Feature, error) {
	_, featureExists := m.feature_map[name]
	if !featureExists {
		return &Feature{}, errors.New(fmt.Sprintf("The following feature does not exist: %s", name))
	}
	// make a deep copy of the feature
	curr_feature := *m.feature_map[name]
	if curr_feature.enabled != nil {
		curr_feature.enabled = pointer.Bool(*curr_feature.enabled)
	}
	return &curr_feature, nil

}

func (m *FeatureManager) ParseFromFeatureGateStringArgs(featureGateString string) error {
	featureGatesArgs := strings.Split(featureGateString, ",")
	for feature := range featureGatesArgs {
		fmt.Println(feature)
	}
}
