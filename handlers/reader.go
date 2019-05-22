// Copyright (c) Alex Ellis 2017, Ken Fukuyama 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gitmonster/faas-rancher/rancher"
	"github.com/juju/errors"
)

// MakeFunctionReader handler for reading functions deployed in the cluster as deployments.
func MakeFunctionReader(client rancher.BridgeClient) VarsHandler {
	return func(w http.ResponseWriter, r *http.Request, vars map[string]string) {

		functions, err := getServiceList(client)
		if err != nil {
			handleServerError(w, errors.Annotate(err, "getServiceList"))
			return
		}

		functionBytes, marshalErr := json.Marshal(functions)
		if marshalErr != nil {
			handleServerError(w, errors.Annotate(marshalErr, "Marshal"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(functionBytes)
	}
}
