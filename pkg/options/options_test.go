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

package options

import (
	"io/ioutil"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ConfigRead(t *testing.T) {
	var filePath = "test_config.yaml"
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		t.Fatal(err)
	}
	var opt = &InnerRisingWaveOptions{}
	opt.unmarshal(data)
	assert.Equal(t, string(opt.Default.PullPolicy), "IfNotPresent")
	opt.install()
	assert.Equal(t, opt.MetaNode.Replicas, int32(2))
	assert.Equal(t, opt.MetaNode.Image["arm64"].Repository, "ghcr.io/singularity-data/risingwave-arm64")
	assert.Equal(t, opt.MetaNode.Image["arm64"].Tag, "arm64")
	assert.Equal(t, opt.MetaNode.Image["amd64"].Tag, "amd64")
	assert.Equal(t, opt.Frontend.Image["amd64"].Tag, "amd64-001")
}

func Test_ValueReplace(t *testing.T) {
	var value = "test"
	var str = `this is ${value} data`
	var str2 = `this is no value data`

	str = strings.Replace(str, "${value}", value, 1)
	assert.Equal(t, str, "this is test data")
	str2 = strings.Replace(str2, "${value}", value, 1)
	assert.Equal(t, str2, "this is no value data")
}
