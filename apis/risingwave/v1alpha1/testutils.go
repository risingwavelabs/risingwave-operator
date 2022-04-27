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

package v1alpha1

import (
	"fmt"
	"os"

	"k8s.io/client-go/tools/clientcmd"
)

// CheckConfigFile
// if $HOME/.kube/config exist
// return ture.
func CheckConfigFile() bool {
	home, e := os.LookupEnv("HOME")
	if !e {
		return false
	}
	var path = fmt.Sprintf("%s/.kube/config", home)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	config, err := clientcmd.BuildConfigFromFlags("", path)
	if err != nil {
		return false
	}
	if config != nil {
		return true
	}
	return false
}

var needTest = CheckConfigFile()
