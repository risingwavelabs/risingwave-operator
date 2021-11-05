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
	"fmt"
	risingwavev1alpha1 "github.com/singularity-data/risingwave-operator/apis/risingwave/v1alpha1"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"testing"
)

func initParser() error {
	home, e := os.LookupEnv("HOME")
	if !e {
		return fmt.Errorf("no HOME env")
	}
	config, err := clientcmd.BuildConfigFromFlags("", fmt.Sprintf("%s/.kube/config", home))
	if err != nil {
		return err
	}
	i := &InnerRESTClientGetter{
		Config: config,
	}
	s := runtime.NewScheme()
	clientgoscheme.AddToScheme(s)
	risingwavev1alpha1.AddToScheme(s)
	NewParser(i)
	return nil
}

func TestParseFile(t *testing.T) {
	err := initParser()
	if err != nil {
		t.Log(err)
		return
	}
	var path = "test/test-template.yaml"
	opt := map[string]interface{}{
		"Name":      "test",
		"Namespace": "test-ns",
	}

	objs, err := ParseFile(path, opt)
	if err != nil {
		t.Fatal(err)
	}

	// test for risingwave with namespace
	obj := objs[0]
	assert.Equal(t, obj.GetObjectKind().GroupVersionKind().Kind, "ConfigMap")
	m := meta.NewAccessor()
	name, _ := m.Name(obj)
	assert.Equal(t, name, "test")
	ns, _ := m.Namespace(obj)
	assert.Equal(t, ns, "test-ns")
}
