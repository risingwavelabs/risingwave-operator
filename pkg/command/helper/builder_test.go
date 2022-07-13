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

package helper

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/api/meta/testrestmapper"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/resource"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest/fake"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"

	risingwavev1alpha1 "github.com/singularity-data/risingwave-operator/apis/risingwave/v1alpha1"
)

var serviceStr = `apiVersion: v1
kind: Service
metadata:
  name: test-service
spec:
  type: NodePort
  selector:
    app: frontend
  ports:
    - nodePort: 12345
      port: 4567
      targetPort: 4567
      name: test
`

func fakeClient() resource.FakeClientFunc {
	return func(version schema.GroupVersion) (resource.RESTClient, error) {
		return &fake.RESTClient{}, nil
	}
}

func newDefaultBuilder() *resource.Builder {
	s := runtime.NewScheme()
	clientgoscheme.AddToScheme(s)
	risingwavev1alpha1.AddToScheme(s)

	m := testrestmapper.TestOnlyStaticRESTMapper(s)

	return resource.NewFakeBuilder(
		fakeClient(),
		func() (meta.RESTMapper, error) {
			return m, nil
		},
		func() (restmapper.CategoryExpander, error) {
			return resource.FakeCategoryExpander, nil
		}).
		WithScheme(s, s.PrioritizedVersionsAllGroups()...)
}

func TestRisingWaveFile(t *testing.T) {
	b := newDefaultBuilder().
		NamespaceParam("default").
		DefaultNamespace().
		FilenameParam(false, &resource.FilenameOptions{Recursive: false, Filenames: []string{"test/test-rs.yaml"}})
	r := b.Do()
	if r.Err() != nil {
		t.Fatal(r.Err())
	}
	infos, _ := r.Infos()

	// test for risingwave with namespace
	info := infos[0]
	assert.Equal(t, info.Object.GetObjectKind().GroupVersionKind().Kind, "RisingWave")
	assert.Equal(t, len(info.ResourceVersion), 0)
	assert.Equal(t, info.Name, "test-risingwave-1")
	assert.Equal(t, info.Namespace, "test")

	// test for risingwave no namespace
	info = infos[1]
	assert.Equal(t, len(info.ResourceVersion), 0)
	assert.Equal(t, info.Name, "test-risingwave-2")
	assert.Equal(t, info.Namespace, "default")
}

func TestBuildWithIO(t *testing.T) {
	b := newDefaultBuilder().DefaultNamespace().NamespaceParam("default")
	var reader = bytes.NewReader([]byte(serviceStr))
	b.Stream(reader, "test-reader")
	r := b.Do()
	if r.Err() != nil {
		t.Fatal(r.Err())
	}
	infos, _ := r.Infos()
	// test for risingwave with namespace
	info := infos[0]
	assert.Equal(t, info.Object.GetObjectKind().GroupVersionKind().Kind, "Service")
	assert.Equal(t, len(info.ResourceVersion), 0)
	assert.Equal(t, info.Name, "test-service")
	assert.Equal(t, info.Namespace, "default")
}

func TestBuilder(t *testing.T) {
	home, e := os.LookupEnv("HOME")
	if !e {
		return
	}
	config, err := clientcmd.BuildConfigFromFlags("", fmt.Sprintf("%s/.kube/config", home))
	if err != nil {
		t.Log(err)
		return
	}

	s := runtime.NewScheme()
	clientgoscheme.AddToScheme(s)
	risingwavev1alpha1.AddToScheme(s)
	t.Log(config)

	i := &InnerRESTClientGetter{
		Config: config,
	}
	b := resource.NewBuilder(i).DefaultNamespace().NamespaceParam("default").Unstructured()

	var reader = bytes.NewReader([]byte(serviceStr))
	b.Stream(reader, "test-reader")
	r := b.Do()
	if r.Err() != nil {
		t.Fatal(r.Err())
	}
	infos, _ := r.Infos()
	// test for risingwave with namespace
	info := infos[0]
	assert.Equal(t, info.Object.GetObjectKind().GroupVersionKind().Kind, "Service")
	assert.Equal(t, len(info.ResourceVersion), 0)
}
