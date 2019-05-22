// Copyright (c) 2017 Ken Fukuyama
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package handlers

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

var (
	logger = logrus.WithField("package", "handlers")
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
