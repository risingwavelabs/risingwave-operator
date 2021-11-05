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

package rendor

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestTemplate(t *testing.T) {
	var path = "test/test-template.yaml"
	opt := map[string]interface{}{
		"Name":      "test",
		"Namespace": "test-ns",
	}
	b, err := Template(path, opt)
	if err != nil {
		t.Fatal(err)
	}
	objStr := string(b)
	assert.Equal(t, strings.Contains(objStr, "name: test"), true)
	assert.Equal(t, strings.Contains(objStr, "namespace: test-ns"), true)
}
