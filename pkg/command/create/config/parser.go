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

import (
	"github.com/BurntSushi/toml"
)

type resource struct {
	CPU    string
	Memory string
}
type baseConfig struct {
	Arch     string
	Replicas int
	Limit    resource
	Request  resource
}

type group struct {
	Replicas int
	Name     string
	Limit    resource
	Request  resource
}

type componentConfig struct {
	Groups []group
}

type frontendConfig struct {
	componentConfig
}
type innerConfig struct {
	Global    baseConfig
	Meta      componentConfig
	Compute   componentConfig
	Compactor componentConfig
	Frontend  frontendConfig
}

func parse(path string) (innerConfig, error) {
	var c innerConfig
	if _, err := toml.DecodeFile(path, &c); err != nil {
		return c, err
	}
	return c, nil
}
