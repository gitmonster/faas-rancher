// Copyright (c) Alex Ellis 2017, Ken Fukuyama 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"regexp"

	"github.com/gitmonster/faas-rancher/rancher"
	"github.com/juju/errors"
	"github.com/openfaas/faas/gateway/requests"
	client "github.com/rancher/go-rancher/v2"
)

// ValidateDeployRequest validates that the service name is valid for Kubernetes
func ValidateDeployRequest(request *requests.CreateFunctionRequest) error {
	var validDNS = regexp.MustCompile(`^[a-zA-Z\-]+$`)
	matched := validDNS.MatchString(request.Service)
	if matched {
		return nil
	}

	return errors.Errorf("(%s) must be a valid DNS entry for service name", request.Service)
}

// MakeDeployHandler creates a handler to create new functions in the cluster
func MakeDeployHandler(client rancher.BridgeClient) VarsHandler {
	return func(w http.ResponseWriter, r *http.Request, vars map[string]string) {

		defer r.Body.Close()

		body, _ := ioutil.ReadAll(r.Body)

		request := requests.CreateFunctionRequest{}
		err := json.Unmarshal(body, &request)
		if err != nil {
			handleBadRequest(w, errors.Annotate(err, "Unmarshal"))
			return
		}

		if err := ValidateDeployRequest(&request); err != nil {
			handleBadRequest(w, errors.Annotate(err, "ValidateDeployRequest"))
			return
		}

		serviceSpec := makeServiceSpec(request)

		_, err = client.CreateService(serviceSpec)
		if err != nil {
			handleServerError(w, errors.Annotate(err, "CreateService"))
			return
		}

		logger.Infof("Created service %s", request.Service)
		logger.Debug(string(body))
		w.WriteHeader(http.StatusAccepted)
	}
}

func makeServiceSpec(request requests.CreateFunctionRequest) *client.Service {
	envVars := make(map[string]interface{})
	for k, v := range request.EnvVars {
		envVars[k] = v
	}

	if len(request.EnvProcess) > 0 {
		envVars["fprocess"] = request.EnvProcess
	}

	// transfer request labels
	labels := make(map[string]interface{})
	if request.Labels != nil {
		for k, v := range *request.Labels {
			labels[k] = v
		}
	}

	labels[FaasFunctionLabel] = request.Service
	labels["io.rancher.container.pull_image"] = "always"

	launchConfig := &client.LaunchConfig{
		Environment: envVars,
		ImageUuid:   "docker:" + request.Image, // not sure if it's ok to just prefix with 'docker:'
		Labels:      labels,
	}

	serviceSpec := &client.Service{
		Name:          request.Service,
		Scale:         1,
		StartOnCreate: true,
		LaunchConfig:  launchConfig,
	}

	return serviceSpec
}
