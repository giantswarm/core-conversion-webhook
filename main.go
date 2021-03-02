package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/dyson/certman"
	"github.com/giantswarm/microerror"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/cluster-api/api/v1alpha3"
	"sigs.k8s.io/cluster-api/api/v1alpha4"

	"github.com/giantswarm/core-conversion-webhook/config"
	"github.com/giantswarm/core-conversion-webhook/pkg/crdconverter"
	"github.com/giantswarm/core-conversion-webhook/pkg/crdhandler"
)

func main() {
	err := mainErr()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, microerror.Pretty(err, true))
	}
}

func mainErr() error {
	webhookConfig, err := config.Parse()
	if err != nil {
		return microerror.Mask(err)
	}

	scheme := runtime.NewScheme()
	{
		err = v1alpha3.AddToScheme(scheme)
		if err != nil {
			return microerror.Mask(err)
		}
		err = v1alpha4.AddToScheme(scheme)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	var crdHandler http.Handler
	{
		crdConverter, err := crdconverter.New(crdconverter.Config{
			Logger: webhookConfig.Logger,
		})
		if err != nil {
			return microerror.Mask(err)
		}

		crdHandler, err = crdhandler.New(crdhandler.Config{
			Logger:    webhookConfig.Logger,
			Converter: crdConverter,
			Scheme:    scheme,
		})
		if err != nil {
			return microerror.Mask(err)
		}
	}

	// Here we register our endpoints.
	handler := http.NewServeMux()
	handler.Handle("/crdconvert", crdHandler)

	handler.HandleFunc("/healthz", healthCheck)

	metrics := http.NewServeMux()
	metrics.Handle("/metrics", promhttp.Handler())

	go serveMetrics(webhookConfig, metrics)
	serveTLS(webhookConfig, handler)

	return nil
}

func healthCheck(writer http.ResponseWriter, _ *http.Request) {
	writer.WriteHeader(http.StatusOK)
	_, err := writer.Write([]byte("ok"))
	if err != nil {
		panic(microerror.JSON(err))
	}
}

func serveTLS(config config.Config, handler http.Handler) {
	cm, err := certman.New(config.CertFile, config.KeyFile)
	if err != nil {
		panic(microerror.JSON(err))
	}
	if err := cm.Watch(); err != nil {
		panic(microerror.JSON(err))
	}

	server := &http.Server{
		Addr:    config.Address,
		Handler: handler,
		TLSConfig: &tls.Config{
			GetCertificate: cm.GetCertificate,
		},
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM)
	go func() {
		<-sig
		err := server.Shutdown(context.Background())
		if err != nil {
			panic(microerror.JSON(err))
		}
	}()

	err = server.ListenAndServeTLS("", "")
	if err != nil {
		if err != http.ErrServerClosed {
			panic(microerror.JSON(err))
		}
	}
}

func serveMetrics(config config.Config, handler http.Handler) {
	server := &http.Server{
		Addr:    config.MetricsAddress,
		Handler: handler,
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM)
	go func() {
		<-sig
		err := server.Shutdown(context.Background())
		if err != nil {
			panic(microerror.JSON(err))
		}
	}()

	err := server.ListenAndServe()
	if err != nil {
		if err != http.ErrServerClosed {
			panic(microerror.JSON(err))
		}
	}
}
