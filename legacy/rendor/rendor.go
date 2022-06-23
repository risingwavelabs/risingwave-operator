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
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"
	"text/template"
)

func Template(filename string, obj interface{}) ([]byte, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return parseString(string(data), obj)
}

func parseString(temStr string, obj interface{}) ([]byte, error) {
	var buf bytes.Buffer
	tmpl, err := template.New("template").Funcs(template.FuncMap{"spaces": spaces}).Parse(temStr)
	if err != nil {
		return nil, fmt.Errorf("error when parsing template, %w", err)
	}
	err = tmpl.Execute(&buf, obj)
	if err != nil {
		return nil, fmt.Errorf("error when executing template, %w", err)
	}
	return buf.Bytes(), nil
}

func spaces(n int, v string) string {
	pad := strings.Repeat(" ", n)
	return pad + strings.Replace(v, "\n", "\n"+pad, -1)
}
