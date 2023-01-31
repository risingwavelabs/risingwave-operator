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

package utils

import (
	"context"
	"fmt"
	"strings"

	"github.com/samber/lo"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetCustomResourceDefinition is the helper function to get the CRD for the given group kind.
func GetCustomResourceDefinition(ctx context.Context, client client.Reader, gk metav1.GroupKind) (*apiextensionsv1.CustomResourceDefinition, error) {
	var crd apiextensionsv1.CustomResourceDefinition
	err := client.Get(ctx, types.NamespacedName{Name: fmt.Sprintf("%ss.%s", strings.ToLower(gk.Kind), gk.Group)}, &crd)
	if err != nil {
		return nil, err
	}
	return &crd, nil
}

// IsVersionServingInCustomResourceDefinition returns true when the CRD serves the given version.
func IsVersionServingInCustomResourceDefinition(crd *apiextensionsv1.CustomResourceDefinition, version string) bool {
	if crd == nil || version == "" {
		return false
	}
	return lo.ContainsBy(crd.Spec.Versions, func(v apiextensionsv1.CustomResourceDefinitionVersion) bool {
		return v.Name == version && v.Served && !v.Deprecated
	})
}
