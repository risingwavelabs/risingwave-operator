module github.com/singularity-data/risingwave-operator

go 1.16

require (
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.13.0
	github.com/stretchr/testify v1.7.0
	github.com/thoas/go-funk v0.9.2
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
	gotest.tools/v3 v3.0.3
	k8s.io/api v0.21.2
	k8s.io/apimachinery v0.21.2
	k8s.io/cli-runtime v0.21.2
	k8s.io/client-go v0.21.2
	sigs.k8s.io/controller-runtime v0.9.2
	sigs.k8s.io/yaml v1.2.0 // indirect
)
