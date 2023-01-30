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

package config

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Parse(t *testing.T) {
	c, e := parse("example.toml")
	if e != nil {
		t.Fatal(e)
	}
	assert.Equal(t, len(c.Frontend.Groups), 2)
}

func Test_ForNoPath(t *testing.T) {
	_, e := parse("fake.toml")
	if e != nil {
		assert.Equal(t, strings.Contains(e.Error(), "no such file or directory"), true)
	}
}
