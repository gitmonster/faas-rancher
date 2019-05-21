// Copyright (c) Alex Ellis 2017, Ken Fukuyama 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/gitmonster/faas-rancher/rancher"
	"github.com/juju/errors"
	"github.com/openfaas/faas/gateway/requests"
)

// MakeDeleteHandler delete a function
func MakeDeleteHandler(client rancher.BridgeClient) VarsHandler {
	return func(w http.ResponseWriter, r *http.Request, vars map[string]string) {

		defer r.Body.Close()
		body, _ := ioutil.ReadAll(r.Body)

		request := requests.DeleteFunctionRequest{}
		err := json.Unmarshal(body, &request)
		if err != nil {
			handleServerError(w, errors.Annotate(err, "Unmarshal"))
			return
		}

		if len(request.FunctionName) == 0 {
			handleBadRequest(w, errors.New("FunctionName is empty"))
			return
		}

		// This makes sure we don't delete non-labelled deployments
		service, findErr := client.FindServiceByName(request.FunctionName)
		if findErr != nil {
			handleServerError(w, errors.Annotate(findErr, "FindServiceByName"))
			return
		} else if service == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		delErr := client.DeleteService(service)
		if delErr != nil {
			handleBadRequest(w, errors.Annotate(delErr, "DeleteService"))
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
