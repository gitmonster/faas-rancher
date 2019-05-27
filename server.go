// Copyright (c) Alex Ellis 2017, Ken Fukuyama 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/gitmonster/faas-rancher/handlers"
	"github.com/gitmonster/faas-rancher/metastore"
	"github.com/gitmonster/faas-rancher/rancher"
	"github.com/juju/errors"
	"github.com/kelseyhightower/envconfig"
	bootstrap "github.com/openfaas/faas-provider"
	proxy "github.com/openfaas/faas-provider/proxy"
	bootTypes "github.com/openfaas/faas-provider/types"
	"github.com/sirupsen/logrus"
)

var (
	// CommitSHA gets overwritten by build process
	logger    = logrus.WithField("package", "main")
	settings  Settings
	CommitSHA = "n/a"
)

const (
	// Version is the current version
	Version = "0.13.0"
)

type Settings struct {
	Debug                  bool          `default:"false"`
	RancherCattleURL       string        `default:"" required:"true" split_words:"true"`
	RancherCattleAccessKey string        `default:"" required:"true" split_words:"true"`
	RancherCattleSecretKey string        `default:"" required:"true" split_words:"true"`
	FaasStackName          string        `default:"faas-functions" required:"true" split_words:"true"`
	FaasProxyTimeout       time.Duration `default:"10s" split_words:"true"`
	FaasReadTimeout        time.Duration `default:"8s" split_words:"true"`
	FaasWriteTimeout       time.Duration `default:"8s" split_words:"true"`
	FaasPort               int           `default:"8080" split_words:"true"`
}

func main() {
	logrus.SetOutput(os.Stdout)

	logger.Info("process settings")
	if err := envconfig.Process("", &settings); err != nil {
		logger.Fatal(errors.Annotate(err, "Process [settings"))
	}

	if settings.Debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	// creates the rancher client config
	config, err := rancher.NewClientConfig(
		settings.FaasStackName,
		settings.RancherCattleURL,
		settings.RancherCattleAccessKey,
		settings.RancherCattleSecretKey,
	)

	if err != nil {
		log.Fatal(errors.Annotate(err, "NewClientConfig"))
	}

	logger.Debug("created rancher client")
	rancherClient, err := rancher.NewClientForConfig(config)
	if err != nil {
		logger.Fatal(errors.Annotate(err, "NewClientForConfig"))
	}

	logger.Debug("open storage")
	if err := metastore.Open(); err != nil {
		logger.Fatal(errors.Annotate(err, "Open [metastore]"))
	}

	defer metastore.Close()

	resolver := NewFunctionURLResolver(8080)
	var bootstrapHandlers bootTypes.FaaSHandlers

	if settings.Debug {
		decorateDebug := func(name string, fn http.HandlerFunc) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				logger.Debugf("enter %s", name)
				defer logger.Debugf("leave %s", name)
				fn(w, r)
			}
		}

		bootstrapHandlers = bootTypes.FaaSHandlers{
			FunctionProxy:  decorateDebug("Proxy", proxy.NewHandlerFunc(settings.FaasProxyTimeout, resolver)),
			DeleteHandler:  decorateDebug("DeleteHandler", handlers.MakeDeleteHandler(rancherClient).ServeHTTP),
			DeployHandler:  decorateDebug("DeployHandler", handlers.MakeDeployHandler(rancherClient).ServeHTTP),
			FunctionReader: decorateDebug("FunctionReader", handlers.MakeFunctionReader(rancherClient).ServeHTTP),
			ReplicaReader:  decorateDebug("ReplicaReader", handlers.MakeReplicaReader(rancherClient).ServeHTTP),
			ReplicaUpdater: decorateDebug("ReplicaUpdater", handlers.MakeReplicaUpdater(rancherClient).ServeHTTP),
			UpdateHandler:  decorateDebug("UpdateHandler", handlers.MakeUpdateHandler(rancherClient).ServeHTTP),
			SecretHandler:  decorateDebug("SecretHandler", handlers.MakeSecretHandler(rancherClient)),
			InfoHandler:    decorateDebug("InfoHandler", handlers.MakeInfoHandler(Version, CommitSHA)),
			HealthHandler:  decorateDebug("HealthHandler", handlers.MakeHealthHandler()),
		}
	} else {
		bootstrapHandlers = bootTypes.FaaSHandlers{
			FunctionProxy:  proxy.NewHandlerFunc(settings.FaasProxyTimeout, resolver),
			DeleteHandler:  handlers.MakeDeleteHandler(rancherClient).ServeHTTP,
			DeployHandler:  handlers.MakeDeployHandler(rancherClient).ServeHTTP,
			FunctionReader: handlers.MakeFunctionReader(rancherClient).ServeHTTP,
			ReplicaReader:  handlers.MakeReplicaReader(rancherClient).ServeHTTP,
			ReplicaUpdater: handlers.MakeReplicaUpdater(rancherClient).ServeHTTP,
			UpdateHandler:  handlers.MakeUpdateHandler(rancherClient).ServeHTTP,
			SecretHandler:  handlers.MakeSecretHandler(rancherClient),
			InfoHandler:    handlers.MakeInfoHandler(Version, CommitSHA),
			HealthHandler:  handlers.MakeHealthHandler(),
		}
	}

	port := settings.FaasPort
	bootstrapConfig := bootTypes.FaaSConfig{
		ReadTimeout:  settings.FaasReadTimeout,
		WriteTimeout: settings.FaasWriteTimeout,
		TCPPort:      &port,
	}

	bootstrap.Serve(&bootstrapHandlers, &bootstrapConfig)
}

type FunctionURLResolver struct {
	watchdogPort int
}

func (p *FunctionURLResolver) Resolve(service string) (url.URL, error) {
	u, err := url.Parse(fmt.Sprintf("http://%s.%s:%d/",
		service,
		settings.FaasStackName,
		p.watchdogPort,
	))
	return *u, err
}

func NewFunctionURLResolver(watchdogPort int) *FunctionURLResolver {
	r := FunctionURLResolver{
		watchdogPort: watchdogPort,
	}

	return &r
}
