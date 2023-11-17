module github.com/giantswarm/dns-operator-gcp

go 1.19

replace (
	// Fix non CVE vulnerability: sonatype-2020-1759 in aws/aws-sdk-go v1.8.39
	github.com/aws/aws-sdk-go => github.com/aws/aws-sdk-go v1.44.24

	// Fix multiple vulnerabilities caused by transitive dependency k8s.io/kubernetes@v1.13.0
	// This is caused by importing sigs.k8s.io/cluster-api-provider-gcp@v1.0.2.
	// The current main branch contains updated dependencies, but has not been released yet,
	// which means that this replace can be removed in with the next version.
	github.com/containerd/containerd => github.com/containerd/containerd v1.7.9

	// Fix vulnerability: CVE-2020-15114 in etcd v3.3.13+incompatible
	github.com/coreos/etcd => github.com/coreos/etcd v3.3.24+incompatible

	// Fix vulnerability: CVE-2020-26160 in dgrijalva/jwt-go v3.2.0
	// This package is archived and is replaced by golang-jwt/jwt
	github.com/dgrijalva/jwt-go => github.com/golang-jwt/jwt v3.2.2+incompatible

	// Fix minor vulnerability: CVE-2022-29162 in opencontainers/runc v1.1.1
	github.com/opencontainers/runc => github.com/opencontainers/runc v1.1.2

	// Fix non CVE vulnerability: sonatype-2019-0890 in pkg/sftp v1.10.1
	github.com/pkg/sftp => github.com/pkg/sftp v1.13.4

	// Explicitly use newest version of cluster-api, instead of one brought
	// from cluster-api-provider-gcp@v1.0.2
	sigs.k8s.io/cluster-api => sigs.k8s.io/cluster-api v1.1.3
	sigs.k8s.io/cluster-api/test => sigs.k8s.io/cluster-api/test v1.1.3
)

require (
	github.com/giantswarm/microerror v0.4.0
	github.com/go-logr/logr v1.2.4
	github.com/google/uuid v1.3.0
	github.com/maxbrunsfeld/counterfeiter/v6 v6.5.0
	github.com/miekg/dns v1.1.50
	github.com/onsi/ginkgo/v2 v2.6.1
	github.com/onsi/gomega v1.24.2
	go.uber.org/zap v1.21.0
	google.golang.org/api v0.126.0
	k8s.io/api v0.26.2
	k8s.io/apimachinery v0.26.2
	k8s.io/client-go v0.26.2
	sigs.k8s.io/cluster-api v1.1.3
	sigs.k8s.io/cluster-api-provider-gcp v1.0.2
	sigs.k8s.io/controller-runtime v0.12.1
)

require (
	cloud.google.com/go/compute v1.21.0 // indirect
	cloud.google.com/go/compute/metadata v0.2.3 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/blang/semver v3.5.1+incompatible // indirect
	github.com/blang/semver/v4 v4.0.0 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/emicklei/go-restful/v3 v3.10.1 // indirect
	github.com/evanphx/json-patch v5.6.0+incompatible // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/go-logr/zapr v1.2.3 // indirect
	github.com/go-openapi/jsonpointer v0.19.5 // indirect
	github.com/go-openapi/jsonreference v0.20.0 // indirect
	github.com/go-openapi/swag v0.21.1 // indirect
	github.com/gobuffalo/flect v0.2.5 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/gnostic v0.6.9 // indirect
	github.com/google/go-cmp v0.5.9 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/s2a-go v0.1.4 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.2.3 // indirect
	github.com/googleapis/gax-go/v2 v2.11.0 // indirect
	github.com/imdario/mergo v0.3.13 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/client_golang v1.14.0 // indirect
	github.com/prometheus/client_model v0.4.0 // indirect
	github.com/prometheus/common v0.37.0 // indirect
	github.com/prometheus/procfs v0.8.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.8.0 // indirect
	golang.org/x/crypto v0.14.0 // indirect
	golang.org/x/mod v0.11.0 // indirect
	golang.org/x/net v0.17.0 // indirect
	golang.org/x/oauth2 v0.10.0 // indirect
	golang.org/x/sys v0.13.0 // indirect
	golang.org/x/term v0.13.0 // indirect
	golang.org/x/text v0.13.0 // indirect
	golang.org/x/time v0.3.0 // indirect
	golang.org/x/tools v0.10.0 // indirect
	gomodules.xyz/jsonpatch/v2 v2.2.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20230711160842-782d3b101e98 // indirect
	google.golang.org/grpc v1.58.3 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/apiextensions-apiserver v0.24.1 // indirect
	k8s.io/component-base v0.26.2 // indirect
	k8s.io/klog/v2 v2.90.1 // indirect
	k8s.io/kube-openapi v0.0.0-20221012153701-172d655c2280 // indirect
	k8s.io/utils v0.0.0-20230220204549-a5ecb0141aa5 // indirect
	sigs.k8s.io/json v0.0.0-20220713155537-f223a00ba0e2 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.3 // indirect
	sigs.k8s.io/yaml v1.3.0 // indirect
)
