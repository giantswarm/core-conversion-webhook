module github.com/giantswarm/core-conversion-webhook

go 1.16

require (
	github.com/dyson/certman v0.2.1
	github.com/giantswarm/microerror v0.3.0
	github.com/giantswarm/micrologger v0.5.0
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822
	github.com/prometheus/client_golang v1.11.0
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	k8s.io/apiextensions-apiserver v0.21.1
	k8s.io/apimachinery v0.21.1
	sigs.k8s.io/cluster-api v0.3.13
	sigs.k8s.io/controller-runtime v0.9.0 // indirect
)

replace sigs.k8s.io/cluster-api => sigs.k8s.io/cluster-api v0.3.11-0.20210302171319-f7351b165992
