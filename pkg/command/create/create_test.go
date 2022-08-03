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

package create

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/singularity-data/risingwave-operator/pkg/command/create/config"
)

func Test_CreateInstance(t *testing.T) {
	var o = Options{
		name:       "test",
		namespace:  "test-ns",
		configFile: "config/example.toml",
	}
	c, err := config.ApplyConfigFile(o.configFile, o.arch)
	if err != nil {
		t.Fatal(err)
	}
	o.config = c
	rw, err := o.createInstance()
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, rw.Name, "test")
	assert.Equal(t, rw.Spec.Components.Compactor.Groups[0].Name, "compactor-group-1")
}
