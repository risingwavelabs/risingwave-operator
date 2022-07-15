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

package config

const (
	Image = "ghcr.io/singularity-data/risingwave"
)

// TODO:

// Config contain the fields needed that creating a instance
// TODO: add more fields to create a instance flexibly.
type Config struct {
	Arch  string
	Image string
}

var DefaultConfig = Config{
	Arch:  "arm64",
	Image: Image,
}

// ApplyConfigFile
// TODO: support creating config by file.
func ApplyConfigFile(path string) (Config, error) {
	return DefaultConfig, nil
}
