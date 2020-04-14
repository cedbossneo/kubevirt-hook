module github.com/cedbossneo/kubevirt-custom-hook

go 1.14

require (
	github.com/clbanning/mxj v1.8.4
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.4.0
	google.golang.org/grpc v1.23.0
	kubevirt.io/client-go v0.0.0-00010101000000-000000000000
	kubevirt.io/kubevirt v0.28.0
)

replace (
	github.com/go-kit/kit => github.com/go-kit/kit v0.3.0
	github.com/openshift/api => github.com/openshift/api v0.0.0-20191219222812-2987a591a72c
	github.com/openshift/client-go => github.com/openshift/client-go v0.0.0-20191125132246-f6563a70e19a
	gopkg.in/yaml.v2 => gopkg.in/yaml.v2 v2.2.4

	k8s.io/api => k8s.io/api v0.16.4
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.16.4
	k8s.io/apimachinery => k8s.io/apimachinery v0.16.4
	k8s.io/client-go => k8s.io/client-go v0.16.4
	k8s.io/klog => k8s.io/klog v0.4.0
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20191107075043-30be4d16710a
	kubevirt.io/client-go => kubevirt.io/client-go v0.28.0

	kubevirt.io/containerized-data-importer => kubevirt.io/containerized-data-importer v1.10.6
)
