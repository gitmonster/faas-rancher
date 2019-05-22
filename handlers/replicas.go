// Copyright (c) Alex Ellis 2017, Ken Fukuyama 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package handlers

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/gitmonster/faas-rancher/rancher"
	"github.com/gitmonster/faas-rancher/types"
	"github.com/juju/errors"
	"github.com/openfaas/faas/gateway/requests"
)

// MakeReplicaUpdater updates desired count of replicas
func MakeReplicaUpdater(client rancher.BridgeClient) VarsHandler {
	return func(w http.ResponseWriter, r *http.Request, vars map[string]string) {

		log.Println("Update replicas")

		functionName := vars["name"]

		req := types.ScaleServiceRequest{}
		if r.Body != nil {
			defer r.Body.Close()
			bytesIn, _ := ioutil.ReadAll(r.Body)
			marshalErr := json.Unmarshal(bytesIn, &req)
			if marshalErr != nil {
				log.Println(errors.Annotate(marshalErr, "Unmarshal"))
				w.WriteHeader(http.StatusBadRequest)
				msg := "Cannot parse request. Please pass valid JSON."
				w.Write([]byte(msg))
				log.Println(msg, marshalErr)
				return
			}
		}

		service, findErr := client.FindServiceByName(functionName)
		if findErr != nil {
			log.Println(errors.Annotate(findErr, "FindServiceByName"))
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Unable to lookup function deployment " + functionName))
			return
		}

		updates := make(map[string]string)
		updates["scale"] = strconv.FormatInt(req.Replicas, 10)
		_, upgradeErr := client.UpdateService(service, updates)
		if upgradeErr != nil {
			log.Println(errors.Annotate(upgradeErr, "UpdateService"))
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Unable to update function deployment " + functionName))
			return
		}
	}
}

// MakeReplicaReader reads the amount of replicas for a deployment
func MakeReplicaReader(client rancher.BridgeClient) VarsHandler {
	return func(w http.ResponseWriter, r *http.Request, vars map[string]string) {
		functionName := vars["name"]
		functions, err := getServiceList(client)
		if err != nil {
			handleServerError(w, errors.Annotate(err, "getServiceList"))
			return
		}

		var found *requests.Function
		for _, function := range functions {
			if function.Name == functionName {
				found = &function
				break
			}
		}

		if found == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		buf, err := json.Marshal(found)
		if err != nil {
			handleServerError(w, errors.Annotate(err, "Marshal"))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(buf)
	}
}
