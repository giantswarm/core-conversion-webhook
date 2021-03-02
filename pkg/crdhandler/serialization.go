package crdhandler

import (
	"errors"
	"fmt"
	"strings"

	"github.com/munnerz/goautoneg"
	"k8s.io/apimachinery/pkg/runtime"
)

type mediaType struct {
	Type, SubType string
}

func (h Handler) getInputSerializer(contentType string) (runtime.Serializer, error) {
	parts := strings.SplitN(contentType, "/", 2)
	if len(parts) != 2 {
		return nil, errors.New(fmt.Sprintf("invalid content-type %#q", contentType))
	}
	return h.serializers[mediaType{parts[0], parts[1]}], nil
}

func (h Handler) getOutputSerializer(accept string) (runtime.Serializer, error) {
	if len(accept) == 0 {
		return h.serializers[mediaType{"application", "json"}], nil
	}

	clauses := goautoneg.ParseAccept(accept)
	for _, clause := range clauses {
		for k, v := range h.serializers {
			switch {
			case clause.Type == k.Type && clause.SubType == k.SubType,
				clause.Type == k.Type && clause.SubType == "*",
				clause.Type == "*" && clause.SubType == "*":
				return v, nil
			}
		}
	}

	return nil, errors.New("unknown type")
}
