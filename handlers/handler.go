// Copyright (c) 2017 Ken Fukuyama
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package handlers

import (
	"net/http"

	"github.com/gitmonster/faas-rancher/helper"
	"github.com/gitmonster/faas-rancher/metastore"
	"github.com/gitmonster/faas-rancher/rancher"
	"github.com/gorilla/mux"
	"github.com/juju/errors"
	"github.com/openfaas/faas-cli/schema"
	"github.com/openfaas/faas/gateway/requests"
	rancherClient "github.com/rancher/go-rancher/v2"
	"github.com/sirupsen/logrus"
)

var (
	logger = logrus.WithField("package", "handlers")
)

const (
	// FaasFunctionLabel is the label set to faas function containers
	FaasFunctionLabel = "faas_function"
)

// VarsHandler a wrapper type for mux.Vars
type VarsHandler func(w http.ResponseWriter, r *http.Request, vars map[string]string)

// ServeHTTP a wrapper function for mux.Vars
func (vh VarsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	vh(w, r, vars)
}

func handleServerError(w http.ResponseWriter, err error) {
	logger.Error(err)
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

func handleBadRequest(w http.ResponseWriter, err error) {
	logger.Error(err)
	http.Error(w, err.Error(), http.StatusBadRequest)
}

func getServiceList(client rancher.BridgeClient) ([]requests.Function, error) {
	functions := []requests.Function{}

	services, err := client.ListServices()
	if err != nil {
		return nil, errors.Annotate(err, "ListServices")
	}

	for _, service := range services {
		if service.State != "active" {
			// ignore inactive services
			continue
		}

		if _, ok := service.LaunchConfig.Labels[FaasFunctionLabel]; ok {
			meta := &metastore.FunctionMeta{
				Service: service.Name,
				Image:   service.LaunchConfig.ImageUuid,
			}

			if err := metastore.Read(meta); err != nil {
				if err != metastore.ErrEntityNotFound {
					return nil, errors.Annotate(err, "Read [metastore]")
				}
			}

			// restore meta from rancher service
			if err == metastore.ErrEntityNotFound {
				meta.Service = service.Name
				meta.Image = service.LaunchConfig.ImageUuid
				meta.Labels = service.LaunchConfig.Labels
				meta.Annotations = make(map[string]interface{})

				if envProcess, ok := service.LaunchConfig.Environment["fprocess"]; ok {
					if envProcess, ok := envProcess.(string); ok {
						meta.EnvProcess = envProcess
					}
				}

				if err := metastore.Update(meta); err != nil {
					return nil, errors.Annotate(err, "Update [metastore]")
				}
			}

			// filter to faas function services
			replicas := uint64(service.Scale)
			function := requests.Function{
				Name:              meta.Service,
				Replicas:          replicas,
				AvailableReplicas: replicas,
				Image:             meta.Image,
				EnvProcess:        meta.EnvProcess,
				Labels:            helper.ToFaasMap(meta.Labels),
				Annotations:       helper.ToFaasMap(meta.Annotations),
				InvocationCount:   0,
			}

			functions = append(functions, function)
		}
	}

	return functions, nil
}

func makeUpgradeSpec(

	client rancher.BridgeClient,
	request requests.CreateFunctionRequest,

) (*rancherClient.ServiceUpgrade, error) {
	lc, err := launchConfigFromReq(client, request)
	if err != nil {
		return nil, errors.Annotate(err, "launchConfigFromReq")
	}

	spec := &rancherClient.ServiceUpgrade{
		InServiceStrategy: &rancherClient.InServiceUpgradeStrategy{
			BatchSize:              1,
			StartFirst:             true,
			LaunchConfig:           lc,
			SecondaryLaunchConfigs: []rancherClient.SecondaryLaunchConfig{},
		},
	}

	return spec, nil
}

func makeServiceSpec(

	client rancher.BridgeClient,
	request requests.CreateFunctionRequest,

) (*rancherClient.Service, error) {
	lc, err := launchConfigFromReq(client, request)
	if err != nil {
		return nil, errors.Annotate(err, "launchConfigFromReq")
	}

	serviceSpec := &rancherClient.Service{
		Name:          request.Service,
		Scale:         1,
		StartOnCreate: true,
		LaunchConfig:  lc,
	}

	return serviceSpec, nil
}

func launchConfigFromReq(

	client rancher.BridgeClient,
	request requests.CreateFunctionRequest,

) (*rancherClient.LaunchConfig, error) {

	envVars := make(map[string]interface{})
	for k, v := range request.EnvVars {
		envVars[k] = v
	}

	if len(request.EnvProcess) > 0 {
		envVars["fprocess"] = request.EnvProcess
	}

	labels := helper.ToRancherMap(request.Labels)
	labels[FaasFunctionLabel] = request.Service
	labels["io.rancher.container.pull_image"] = "always"

	lc := &rancherClient.LaunchConfig{
		Environment: envVars,
		ImageUuid:   "docker:" + request.Image, // not sure if it's ok to just prefix with 'docker:'
		Labels:      labels,
	}

	for _, name := range request.Secrets {
		s := schema.Secret{
			Name: name,
		}

		sec, err := lookupSecret(client, &s)
		if err != nil {
			return nil, errors.Annotate(err, "lookupSecret")
		}

		ref := rancherClient.SecretReference{
			Name:     sec.Name,
			SecretId: sec.Id,
		}

		lc.Secrets = append(lc.Secrets, ref)
	}

	return lc, nil
}
