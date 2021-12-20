module github.com/giantswarm/core-conversion-webhook

go 1.16

require (
	github.com/dyson/certman v0.2.1
	github.com/giantswarm/microerror v0.4.0
	github.com/giantswarm/micrologger v0.5.0
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822
	github.com/prometheus/client_golang v1.9.0
	golang.org/x/sys v0.0.0-20210119212857-b64e53b001e4 // indirect
	golang.org/x/tools v0.0.0-20200706234117-b22de6825cf7 // indirect
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	k8s.io/apiextensions-apiserver v0.20.2
	k8s.io/apimachinery v0.20.2
	sigs.k8s.io/cluster-api v0.3.13
)

replace sigs.k8s.io/cluster-api => sigs.k8s.io/cluster-api v0.3.11-0.20210302171319-f7351b165992
