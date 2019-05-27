// Copyright (c) Alex Ellis 2017, Ken Fukuyama 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/gitmonster/faas-rancher/metastore"
	"github.com/gitmonster/faas-rancher/rancher"
	"github.com/juju/errors"
	"github.com/openfaas/faas/gateway/requests"
)

// MakeDeleteHandler delete a function
func MakeDeleteHandler(client rancher.BridgeClient) VarsHandler {
	return func(w http.ResponseWriter, r *http.Request, vars map[string]string) {

		defer r.Body.Close()
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			handleBadRequest(w, errors.Annotate(err, "ReadAll"))
			return
		}

		request := requests.DeleteFunctionRequest{}
		if err := json.Unmarshal(body, &request); err != nil {
			handleServerError(w, errors.Annotate(err, "Unmarshal"))
			return
		}

		if len(request.FunctionName) == 0 {
			handleBadRequest(w, errors.New("FunctionName is empty"))
			return
		}

		// This makes sure we don't delete non-labelled deployments
		service, err := client.FindServiceByName(request.FunctionName)
		if err != nil {
			handleServerError(w, errors.Annotate(err, "FindServiceByName"))
			return
		}

		if service == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if err := client.DeleteService(service); err != nil {
			handleServerError(w, errors.Annotate(err, "DeleteService"))
			return
		}

		meta := &metastore.FunctionMeta{
			Service: service.Name,
			Image:   service.LaunchConfig.ImageUuid,
		}

		if err := metastore.Delete(meta); err != nil {
			if err != metastore.ErrEntityNotFound {
				handleServerError(w, errors.Annotate(err, "Delete [metastore]"))
			}
		}

		w.WriteHeader(http.StatusOK)
	}
}
