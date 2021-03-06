// Copyright (c) Alex Ellis 2017, Ken Fukuyama 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"regexp"

	"github.com/gitmonster/faas-rancher/metastore"
	"github.com/gitmonster/faas-rancher/rancher"
	"github.com/juju/errors"
	"github.com/openfaas/faas-provider/types"
)

// ValidateDeployRequest validates that the service name is valid for Kubernetes
func ValidateDeployRequest(request *types.FunctionDeployment) error {
	var validDNS = regexp.MustCompile(`^[a-zA-Z\-]+$`)
	matched := validDNS.MatchString(request.Service)
	if matched {
		return nil
	}

	return errors.Errorf("%q must be a valid DNS entry for service name", request.Service)
}

// MakeDeployHandler creates a handler to create new functions in the cluster
func MakeDeployHandler(client rancher.BridgeClient) VarsHandler {
	return func(w http.ResponseWriter, r *http.Request, vars map[string]string) {

		defer r.Body.Close()
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			handleBadRequest(w, errors.Annotate(err, "ReadAll"))
			return
		}

		request := types.FunctionDeployment{}
		if err := json.Unmarshal(body, &request); err != nil {
			handleBadRequest(w, errors.Annotate(err, "Unmarshal"))
			return
		}

		if err := ValidateDeployRequest(&request); err != nil {
			handleBadRequest(w, errors.Annotate(err, "ValidateDeployRequest"))
			return
		}

		serviceSpec, err := makeServiceSpec(client, request)
		if err != nil {
			handleServerError(w, errors.Annotate(err, "makeServiceSpec"))
			return
		}

		_, err = client.CreateService(serviceSpec)
		if err != nil {
			handleServerError(w, errors.Annotate(err, "CreateService"))
			return
		}

		meta := metastore.FunctionMeta{}
		if err := metastore.Update(meta.CreateFrom(&request)); err != nil {
			handleServerError(w, errors.Annotate(err, "Update [metastore]"))
			return
		}

		logger.Debugf("Service %q created", request.Service)
		w.WriteHeader(http.StatusAccepted)
	}
}
