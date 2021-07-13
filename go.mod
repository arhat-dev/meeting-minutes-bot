module arhat.dev/meeting-minutes-bot

go 1.16

require (
	arhat.dev/pkg v0.5.5
	github.com/Masterminds/sprig/v3 v3.2.2
	github.com/andybalholm/brotli v1.0.2 // indirect
	github.com/chromedp/chromedp v0.7.4
	github.com/google/uuid v1.2.0 // indirect
	github.com/h2non/filetype v1.1.1
	github.com/huandu/xstrings v1.3.2 // indirect
	github.com/imdario/mergo v0.3.12 // indirect
	github.com/json-iterator/go v1.1.11 // indirect
	github.com/klauspost/compress v1.12.2 // indirect
	github.com/klauspost/cpuid/v2 v2.0.6 // indirect
	github.com/minio/md5-simd v1.1.2 // indirect
	github.com/minio/minio-go/v7 v7.0.10
	github.com/minio/sha256-simd v1.0.0 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/rs/xid v1.3.0 // indirect
	github.com/spf13/cobra v1.1.3
	github.com/spf13/pflag v1.0.5
	github.com/valyala/fasthttp v1.24.0 // indirect
	gitlab.com/toby3d/telegraph v1.2.1
	go.uber.org/multierr v1.5.0
	golang.org/x/crypto v0.0.0-20210513164829-c07d793c2f9a // indirect
	golang.org/x/net v0.0.0-20210521195947-fe42d452be8f // indirect
	gopkg.in/ini.v1 v1.62.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
)

replace github.com/itchyny/gojq => github.com/itchyny/gojq v0.12.3

replace (
	k8s.io/api => github.com/kubernetes/api v0.19.7
	k8s.io/apiextensions-apiserver => github.com/kubernetes/apiextensions-apiserver v0.19.7
	k8s.io/apimachinery => github.com/kubernetes/apimachinery v0.19.7
	k8s.io/apiserver => github.com/kubernetes/apiserver v0.19.7
	k8s.io/cli-runtime => github.com/kubernetes/cli-runtime v0.19.7
	k8s.io/client-go => github.com/kubernetes/client-go v0.19.7
	k8s.io/cloud-provider => github.com/kubernetes/cloud-provider v0.19.7
	k8s.io/cluster-bootstrap => github.com/kubernetes/cluster-bootstrap v0.19.7
	k8s.io/code-generator => github.com/kubernetes/code-generator v0.19.7
	k8s.io/component-base => github.com/kubernetes/component-base v0.19.7
	k8s.io/cri-api => github.com/kubernetes/cri-api v0.19.7
	k8s.io/csi-translation-lib => github.com/kubernetes/csi-translation-lib v0.19.7
	k8s.io/klog => github.com/kubernetes/klog v1.0.0
	k8s.io/klog/v2 => github.com/kubernetes/klog/v2 v2.4.0
	k8s.io/kube-aggregator => github.com/kubernetes/kube-aggregator v0.19.7
	k8s.io/kube-controller-manager => github.com/kubernetes/kube-controller-manager v0.19.7
	k8s.io/kube-proxy => github.com/kubernetes/kube-proxy v0.19.7
	k8s.io/kube-scheduler => github.com/kubernetes/kube-scheduler v0.19.7
	k8s.io/kubectl => github.com/kubernetes/kubectl v0.19.7
	k8s.io/kubelet => github.com/kubernetes/kubelet v0.19.7
	k8s.io/kubernetes => github.com/kubernetes/kubernetes v1.19.7
	k8s.io/legacy-cloud-providers => github.com/kubernetes/legacy-cloud-providers v0.19.7
	k8s.io/metrics => github.com/kubernetes/metrics v0.19.7
	k8s.io/sample-apiserver => github.com/kubernetes/sample-apiserver v0.19.7
	k8s.io/utils => github.com/kubernetes/utils v0.0.0-20201110183641-67b214c5f920
	vbom.ml/util => github.com/fvbommel/util v0.0.2
)
