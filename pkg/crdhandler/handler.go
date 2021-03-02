package crdhandler

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/apimachinery/pkg/types"

	"github.com/giantswarm/core-conversion-webhook/pkg/metrics"
)

type Converter interface {
	Convert(Object *unstructured.Unstructured, toVersion string) (*unstructured.Unstructured, error)
}

type Config struct {
	Logger    micrologger.Logger
	Converter Converter
	Scheme    *runtime.Scheme
}

type Handler struct {
	logger      micrologger.Logger
	converter   Converter
	scheme      *runtime.Scheme
	serializers map[mediaType]runtime.Serializer
}

func (h Handler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	ctx := request.Context()

	start := time.Now()
	defer metrics.DurationRequests.WithLabelValues("converting").Observe(float64(time.Since(start)) / float64(time.Second))

	metrics.TotalRequests.WithLabelValues("converting").Inc()

	contentType := request.Header.Get("Content-Type")
	serializer, err := h.getInputSerializer(contentType)
	if serializer == nil {
		h.logger.Errorf(ctx, err, fmt.Sprintf("invalid Content-Type header `%s`", contentType))
		metrics.InvalidRequests.WithLabelValues("mutating").Inc()
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	data, err := ioutil.ReadAll(request.Body)
	if err != nil {
		h.logger.Errorf(ctx, err, "unable to read request")
		metrics.InternalError.WithLabelValues("mutating").Inc()
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	review := v1.ConversionReview{}
	if _, _, err := serializer.Decode(data, nil, &review); err != nil {
		h.logger.Errorf(ctx, err, "unable to parse admission review request")
		metrics.InvalidRequests.WithLabelValues("mutating").Inc()
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	response, err := doConversion(review.Request, h.converter)
	if err != nil {
		h.writeErrorResponse(review.Request.UID, err)
	}

	h.logger.Debugf(request.Context(), "converted")
	metrics.SuccessfulRequests.WithLabelValues("converting").Inc()

	h.writeResponse(ctx, request, writer, response)
}

func New(config Config) (Handler, error) {
	serializers := map[mediaType]runtime.Serializer{
		{"application", "json"}: json.NewSerializerWithOptions(json.DefaultMetaFactory, config.Scheme, config.Scheme, json.SerializerOptions{
			Yaml:   false,
			Pretty: false,
			Strict: false,
		}),
		{"application", "yaml"}: json.NewSerializerWithOptions(json.DefaultMetaFactory, config.Scheme, config.Scheme, json.SerializerOptions{
			Yaml:   true,
			Pretty: false,
			Strict: false,
		}),
	}

	return Handler{
		logger:      config.Logger,
		converter:   config.Converter,
		serializers: serializers,
	}, nil
}

// doConversion converts the requested object given the conversion function and returns a conversion response.
// failures will be reported as Reason in the conversion response.
func doConversion(convertRequest *v1.ConversionRequest, converter Converter) (*v1.ConversionResponse, error) {
	var convertedObjects []runtime.RawExtension
	for _, obj := range convertRequest.Objects {
		cr := unstructured.Unstructured{}
		if err := cr.UnmarshalJSON(obj.Raw); err != nil {
			return nil, microerror.Mask(err)
		}
		convertedCR, err := converter.Convert(&cr, convertRequest.DesiredAPIVersion)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		convertedCR.SetAPIVersion(convertRequest.DesiredAPIVersion)
		convertedObjects = append(convertedObjects, runtime.RawExtension{Object: convertedCR})
	}
	return &v1.ConversionResponse{
		ConvertedObjects: convertedObjects,
		Result: metav1.Status{
			Status: metav1.StatusSuccess,
		},
	}, nil
}

func (h Handler) writeResponse(ctx context.Context, request *http.Request, writer http.ResponseWriter, response *v1.ConversionResponse) {
	accept := request.Header.Get("Accept")
	outSerializer, err := h.getOutputSerializer(accept)
	if err != nil {
		h.logger.Errorf(ctx, err, fmt.Sprintf("invalid Accept header `%s`", accept))
		metrics.InvalidRequests.WithLabelValues("mutating").Inc()
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	err = outSerializer.Encode(&v1.ConversionReview{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConversionReview",
			APIVersion: "apiextensions.k8s.io/v1",
		},
		Response: response,
	}, writer)
	if err != nil {
		h.logger.Errorf(ctx, err, "unable to serialize response")
		metrics.InternalError.WithLabelValues("mutating").Inc()
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h Handler) writeErrorResponse(uid types.UID, err error) *v1.ConversionResponse {
	return &v1.ConversionResponse{
		UID: uid,
		Result: metav1.Status{
			Status:  metav1.StatusFailure,
			Message: err.Error(),
		},
	}
}
