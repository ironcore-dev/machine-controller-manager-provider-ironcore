module github.com/onmetal/machine-controller-manager-provider-onmetal

go 1.19

require (
	github.com/Masterminds/sprig v2.22.0+incompatible
	github.com/coreos/butane v0.16.0
	github.com/gardener/gardener v1.51.1
	github.com/gardener/machine-controller-manager v0.47.0
	github.com/google/addlicense v1.1.0
	github.com/imdario/mergo v0.3.12
	github.com/onmetal/controller-utils v0.6.5
	github.com/onmetal/onmetal-api v0.0.13-0.20221209171738-4a6e20c76693
	github.com/onsi/ginkgo/v2 v2.5.1
	github.com/onsi/gomega v1.24.0
	github.com/spf13/pflag v1.0.6-0.20210604193023-d5e0c0615ace
	go.uber.org/zap v1.24.0
	k8s.io/api v0.24.8
	k8s.io/apimachinery v0.26.0
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/component-base v0.24.8
	k8s.io/klog/v2 v2.80.1
	sigs.k8s.io/controller-runtime v0.12.3
	sigs.k8s.io/yaml v1.3.0
)

replace (
	github.com/gardener/machine-controller-manager => github.com/afritzler/machine-controller-manager v0.0.0-20221213081200-699074e5c9e6
	k8s.io/client-go => k8s.io/client-go v0.24.8
)

require (
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/PuerkitoBio/purell v1.1.1 // indirect
	github.com/PuerkitoBio/urlesc v0.0.0-20170810143723-de5bf2ad4578 // indirect
	github.com/aws/aws-sdk-go v1.30.28 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bmatcuk/doublestar/v4 v4.0.2 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/clarketm/json v1.17.1 // indirect
	github.com/coreos/go-json v0.0.0-20220810161552-7cce03887f34 // indirect
	github.com/coreos/go-semver v0.3.0 // indirect
	github.com/coreos/go-systemd/v22 v22.4.0 // indirect
	github.com/coreos/ignition/v2 v2.14.0 // indirect
	github.com/coreos/vcontext v0.0.0-20220810162454-88bd546c634c // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/emicklei/go-restful/v3 v3.8.0 // indirect
	github.com/evanphx/json-patch v4.12.0+incompatible // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/go-logr/zapr v1.2.3 // indirect
	github.com/go-openapi/jsonpointer v0.19.5 // indirect
	github.com/go-openapi/jsonreference v0.19.5 // indirect
	github.com/go-openapi/swag v0.19.14 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/gnostic v0.5.7-v3refs // indirect
	github.com/google/go-cmp v0.5.9 // indirect
	github.com/google/gofuzz v1.1.0 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/huandu/xstrings v1.3.2 // indirect
	github.com/inconshreveable/mousetrap v1.0.1 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/mailru/easyjson v0.7.6 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_golang v1.13.1 // indirect
	github.com/prometheus/client_model v0.3.0 // indirect
	github.com/prometheus/common v0.37.0 // indirect
	github.com/prometheus/procfs v0.8.0 // indirect
	github.com/spf13/cobra v1.6.1 // indirect
	github.com/stretchr/testify v1.8.1 // indirect
	github.com/vincent-petithory/dataurl v1.0.0 // indirect
	go.uber.org/atomic v1.10.0 // indirect
	go.uber.org/multierr v1.8.0 // indirect
	go4.org/netipx v0.0.0-20220812043211-3cc044ffd68d // indirect
	golang.org/x/crypto v0.1.0 // indirect
	golang.org/x/net v0.3.1-0.20221206200815-1e63c2f08a10 // indirect
	golang.org/x/oauth2 v0.0.0-20220411215720-9780585627b5 // indirect
	golang.org/x/sync v0.1.0 // indirect
	golang.org/x/sys v0.3.0 // indirect
	golang.org/x/term v0.3.0 // indirect
	golang.org/x/text v0.5.0 // indirect
	golang.org/x/time v0.0.0-20220609170525-579cf78fd858 // indirect
	gomodules.xyz/jsonpatch/v2 v2.2.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/protobuf v1.28.1 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/apiextensions-apiserver v0.24.8 // indirect
	k8s.io/apiserver v0.24.8 // indirect
	k8s.io/cluster-bootstrap v0.24.8 // indirect
	k8s.io/kube-aggregator v0.24.8 // indirect
	k8s.io/kube-openapi v0.0.0-20221012153701-172d655c2280 // indirect
	k8s.io/utils v0.0.0-20221107191617-1a15be271d1d // indirect
	sigs.k8s.io/json v0.0.0-20220713155537-f223a00ba0e2 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.3 // indirect
)
