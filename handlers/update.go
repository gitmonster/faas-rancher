// Copyright (c) Alex Ellis 2017, Ken Fukuyama 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gitmonster/faas-rancher/metastore"
	"github.com/gitmonster/faas-rancher/rancher"
	"github.com/juju/errors"
	"github.com/openfaas/faas/gateway/requests"
)

// MakeUpdateHandler creates a handler to create new functions in the cluster
func MakeUpdateHandler(client rancher.BridgeClient) VarsHandler {
	return func(w http.ResponseWriter, r *http.Request, vars map[string]string) {

		defer r.Body.Close()

		body, _ := ioutil.ReadAll(r.Body)

		request := requests.CreateFunctionRequest{}
		err := json.Unmarshal(body, &request)
		if err != nil {
			handleBadRequest(w, errors.Annotate(err, "Unmarshal"))
			return
		}

		if len(request.Service) == 0 {
			handleBadRequest(w, errors.New("service is empty"))
			return
		}

		serviceSpec, findErr := client.FindServiceByName(request.Service)
		if findErr != nil {
			handleServerError(w, errors.Annotate(findErr, "FindServiceByName"))
			return
		} else if serviceSpec == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		logger.Debug(serviceSpec.State)

		if serviceSpec.State != "active" {
			handleServerError(w, errors.New("service to upgrade not in active state"))
			return
		}

		upgradeSpec := makeUpgradeSpec(request)
		_, err = client.UpgradeService(serviceSpec, upgradeSpec)
		if err != nil {
			handleServerError(w, errors.Annotate(err, "UpgradeService"))
			return
		}

		meta := metastore.FunctionMeta{}
		if err := metastore.Update(meta.CreateFrom(&request)); err != nil {
			handleServerError(w, errors.Annotate(err, "Write [metastore]"))
			return
		}

		go func() {
			logger.Info("Waiting for upgrade to finish")
			for pollCounter := 20; pollCounter > 0; pollCounter-- {
				pollResult, pollErr := client.FindServiceByName(request.Service)
				logger.Debug(pollResult.State)
				if pollErr != nil {
					logger.Error(errors.Annotate(pollErr, "FindServiceByName"))
					continue
				}
				time.Sleep(1 * time.Second)

				if pollResult.State == "upgraded" {
					logger.Debug("Finishing upgrade")
					_, err = client.FinishUpgradeService(pollResult)
					if err != nil {
						logger.Error(errors.Annotate(err, "FinishUpgradeService"))
						return
					}
					logger.Info("Upgrade finished")
					return
				}
			}
			logger.Warn("Poll timeout!")
		}()

		logger.Infof("Updated service - %s", request.Service)
		w.WriteHeader(http.StatusAccepted)
	}
}
