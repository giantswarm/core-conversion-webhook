package crdconverter

import (
	"errors"
	"fmt"
	"strings"

	"github.com/giantswarm/micrologger"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
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
	convertedObject := object.DeepCopy()
	fromVersion := object.GetAPIVersion()

	if toVersion == fromVersion {
		return nil, errors.New(fmt.Sprintf("conversion from a version to itself should not call the webhook: %s", toVersion))
	}

	switch object.GetAPIVersion() {
	case "stable.example.com/v1":
		switch toVersion {
		case "stable.example.com/v2":
			hostPort, ok := convertedObject.Object["hostPort"]
			if ok {
				delete(convertedObject.Object, "hostPort")
				parts := strings.Split(hostPort.(string), ":")
				if len(parts) != 2 {
					return nil, errors.New(fmt.Sprintf("invalid hostPort value `%v`", hostPort))
				}
				convertedObject.Object["host"] = parts[0]
				convertedObject.Object["port"] = parts[1]
			}
		default:
			return nil, errors.New(fmt.Sprintf("unexpected conversion version %q", toVersion))
		}
	case "stable.example.com/v2":
		switch toVersion {
		case "stable.example.com/v1":
			host, hasHost := convertedObject.Object["host"]
			port, hasPort := convertedObject.Object["port"]
			if hasHost || hasPort {
				if !hasHost {
					host = ""
				}
				if !hasPort {
					port = ""
				}
				convertedObject.Object["hostPort"] = fmt.Sprintf("%s:%s", host, port)
				delete(convertedObject.Object, "host")
				delete(convertedObject.Object, "port")
			}
		default:
			return nil, errors.New(fmt.Sprintf("unexpected conversion version %q", toVersion))
		}
	default:
		return nil, errors.New(fmt.Sprintf("unexpected conversion version %q", fromVersion))
	}
	return convertedObject, nil
}
