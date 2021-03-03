package crdconverter

import (
	"errors"
	"fmt"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/cluster-api/api/v1alpha3"
	"sigs.k8s.io/cluster-api/api/v1alpha4"
)

const (
	ClusterAPIGroup = "cluster.x-k8s.io"
	KindCluster     = "Cluster"
	VersionV1Alpha3 = "v1alpha3"
	VersionV1Alpha4 = "v1alpha4"
)

var (
	capiV1Alpha3 = schema.GroupVersion{
		Group:   ClusterAPIGroup,
		Version: VersionV1Alpha3,
	}
	capiV1Alpha4 = schema.GroupVersion{
		Group:   ClusterAPIGroup,
		Version: VersionV1Alpha4,
	}
	clusterV1Alpha3 = schema.GroupVersionKind{
		Group:   capiV1Alpha3.Group,
		Version: capiV1Alpha3.Version,
		Kind:    KindCluster,
	}
	clusterV1Alpha4 = schema.GroupVersionKind{
		Group:   capiV1Alpha4.Group,
		Version: capiV1Alpha4.Version,
		Kind:    KindCluster,
	}
)

type Config struct {
	Logger micrologger.Logger
}

type Converter struct {
	micrologger.Logger
}

func New(config Config) (Converter, error) {
	return Converter{
		Logger: config.Logger,
	}, nil
}

func (c Converter) Convert(object *unstructured.Unstructured, toVersion string) (*unstructured.Unstructured, error) {
	fromVersion := object.GetAPIVersion()

	if toVersion == fromVersion {
		return nil, errors.New(fmt.Sprintf("conversion from a version to itself should not call the webhook: %s", toVersion))
	}

	var result runtime.Object
	fromGVK := object.GetObjectKind().GroupVersionKind()
	fromContent := object.UnstructuredContent()

	switch {
	case fromGVK == clusterV1Alpha4 && toVersion == capiV1Alpha3.String():
		var fromCluster v1alpha4.Cluster
		err := runtime.DefaultUnstructuredConverter.FromUnstructured(fromContent, &fromCluster)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		toCluster := &v1alpha3.Cluster{}
		err = toCluster.ConvertFrom(&fromCluster)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		result = toCluster
	case fromGVK == clusterV1Alpha3 && toVersion == capiV1Alpha4.String():
		var fromCluster v1alpha3.Cluster
		err := runtime.DefaultUnstructuredConverter.FromUnstructured(fromContent, &fromCluster)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		toCluster := &v1alpha4.Cluster{}
		err = fromCluster.ConvertTo(toCluster)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		result = toCluster
	default:
		return nil, errors.New(fmt.Sprintf("unexpected conversion version %q", fromVersion))
	}

	resultData, err := runtime.DefaultUnstructuredConverter.ToUnstructured(result)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	resultUnstructured := &unstructured.Unstructured{
		Object: resultData,
	}

	return resultUnstructured, nil
}
