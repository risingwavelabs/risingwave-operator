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
	"fmt"
	"testing"
)

func Test_CommonGetVersionFromImage(t *testing.T) {
	testcases := map[string]struct {
		image   string
		version string
	}{
		"empty": {
			image:   "",
			version: "",
		},
		"image-version": {
			image:   "ghcr.io/risingwavelabs/risingwave:v0.1.16",
			version: "v0.1.16",
		},
		"image-port-default": {
			image:   "ghcr.io:10043/risingwavelabs/risingwave",
			version: "latest",
		},
		"host-port-repo-tag": {
			image:   "host:port/path/to/repo:tag",
			version: "tag",
		},
		"host-port-repo": {
			image:   "host:port/path/to/repo",
			version: "latest",
		},
		"image-default": {
			image:   "ghcr.io/risingwavelabs/risingwave",
			version: "latest",
		},
		"image-latest": {
			image:   "ghcr.io/risingwavelabs/risingwave:latest",
			version: "latest",
		},
		"image-simplified": {
			image:   "risingwave",
			version: "latest",
		},
		"image-simplified-with-tag": {
			image:   "risingwave:tag",
			version: "tag",
		},
	}
	for name := range testcases {
		obj := testcases[name]
		t.Run(name, func(t *testing.T) {
			if GetVersionFromImage(obj.image) != obj.version {
				fmt.Printf("Test %s: Input:%s, Output: %s, Expect:%s\n", name, obj.image, GetVersionFromImage(obj.image), obj.version)
				t.Fail()
			}
		})
	}
}
