// Copyright (c) Alex Ellis 2017, Ken Fukuyama 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package main

import (
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/gitmonster/faas-rancher/handlers"
	"github.com/gitmonster/faas-rancher/rancher"
	"github.com/juju/errors"
	bootstrap "github.com/openfaas/faas-provider"
	bootTypes "github.com/openfaas/faas-provider/types"
	"github.com/sirupsen/logrus"
)

var (
	// CommitSHA gets overwritten by build process
	logger    = logrus.WithField("package", "main")
	CommitSHA = "n/a"
)

const (
	// TimeoutSeconds used for http client
	TimeoutSeconds = 2
	// Version is the current version
	Version = "0.13.0"
)

func main() {
	if os.Getenv("DEBUG") == "true" {
		logrus.SetLevel(logrus.DebugLevel)
	}

	functionStackName := os.Getenv("FUNCTION_STACK_NAME")
	cattleURL := os.Getenv("CATTLE_URL")
	cattleAccessKey := os.Getenv("CATTLE_ACCESS_KEY")
	cattleSecretKey := os.Getenv("CATTLE_SECRET_KEY")

	// creates the rancher client config
	config, err := rancher.NewClientConfig(
		functionStackName,
		cattleURL,
		cattleAccessKey,
		cattleSecretKey)

	if err != nil {
		log.Fatal(errors.Annotate(err, "NewClientConfig"))
	}

	// create the rancher REST client
	rancherClient, err := rancher.NewClientForConfig(config)
	if err != nil {
		logger.Fatal(errors.Annotate(err, "NewClientForConfig"))
	}

	logger.Info("Created Rancher Client")

	proxyClient := http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   3 * time.Second,
				KeepAlive: 0,
			}).DialContext,
			MaxIdleConns:          1,
			DisableKeepAlives:     true,
			IdleConnTimeout:       120 * time.Millisecond,
			ExpectContinueTimeout: 1500 * time.Millisecond,
		},
	}

	wrapHandlerFunc := func(name string, fn http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			logger.Debugf("enter %s", name)
			defer logger.Debugf("leave %s", name)
			fn(w, r)
		}
	}

	bootstrapHandlers := bootTypes.FaaSHandlers{
		FunctionProxy:  wrapHandlerFunc("Proxy", handlers.MakeProxy(&proxyClient, config.FunctionsStackName).ServeHTTP),
		DeleteHandler:  wrapHandlerFunc("DeleteHandler", handlers.MakeDeleteHandler(rancherClient).ServeHTTP),
		DeployHandler:  wrapHandlerFunc("DeployHandler", handlers.MakeDeployHandler(rancherClient).ServeHTTP),
		FunctionReader: wrapHandlerFunc("FunctionReader", handlers.MakeFunctionReader(rancherClient).ServeHTTP),
		ReplicaReader:  wrapHandlerFunc("ReplicaReader", handlers.MakeReplicaReader(rancherClient).ServeHTTP),
		ReplicaUpdater: wrapHandlerFunc("ReplicaUpdater", handlers.MakeReplicaUpdater(rancherClient).ServeHTTP),
		UpdateHandler:  wrapHandlerFunc("UpdateHandler", handlers.MakeUpdateHandler(rancherClient).ServeHTTP),
		HealthHandler:  wrapHandlerFunc("HealthHandler", handlers.MakeHealthHandler()),
		InfoHandler:    wrapHandlerFunc("InfoHandler", handlers.MakeInfoHandler(Version, CommitSHA)),
		SecretHandler:  wrapHandlerFunc("SecretHandler", handlers.MakeSecretHandler()),
	}

	// Todo: AE - parse port and parse timeout from env-vars
	var port int
	port = 8080
	bootstrapConfig := bootTypes.FaaSConfig{
		ReadTimeout:  time.Second * 8,
		WriteTimeout: time.Second * 8,
		TCPPort:      &port,
	}

	bootstrap.Serve(&bootstrapHandlers, &bootstrapConfig)
}
