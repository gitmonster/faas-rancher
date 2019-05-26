package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/gitmonster/faas-rancher/rancher"
	"github.com/juju/errors"
	"github.com/openfaas/faas-cli/schema"
	rancherClient "github.com/rancher/go-rancher/v2"
)

// MakeSecretHandler makes a handler for Create/List/Delete/Update of
//secrets in the Rancher API
func MakeSecretHandler(client rancher.BridgeClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		defer r.Body.Close()
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			handleBadRequest(w, errors.Annotate(err, "ReadAll"))
			return
		}

		if len(body) == 0 {
			handleList(client, w)
			return
		}

		secret := schema.Secret{}
		if err := json.Unmarshal(body, &secret); err != nil {
			handleBadRequest(w, errors.Annotate(err, "Unmarshal"))
			return
		}

		if secret.Name != "" && secret.Value != "" {
			if err := ensureSecret(client, &secret); err != nil {
				handleBadRequest(w, errors.Annotate(err, "ensureSecret"))
				return
			}
		} else if secret.Name != "" {
			if err := deleteSecret(client, &secret); err != nil {
				handleBadRequest(w, errors.Annotate(err, "deleteSecret"))
				return
			}
		}

		w.WriteHeader(http.StatusAccepted)
	}
}

func handleList(client rancher.BridgeClient, w http.ResponseWriter) {
	coll, err := client.ListSecrets(nil)
	if err != nil {
		handleServerError(w, errors.Annotate(err, "ListSecrets"))
		return
	}

	var results []schema.Secret
	for _, s := range coll.Data {
		results = append(results, schema.Secret{
			Name:  s.Name,
			Value: s.Value,
		})
	}

	buf, err := json.Marshal(results)
	if err != nil {
		handleBadRequest(w, errors.Annotate(err, "Marshal"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(buf)
}

func deleteSecret(client rancher.BridgeClient, secret *schema.Secret) error {
	old, err := lookupSecret(client, secret)
	if err != nil {
		return errors.Annotate(err, "lookupSecret")
	}

	if old != nil {
		if err := client.DeleteSecret(old); err != nil {
			return errors.Annotate(err, "DeleteSecret")
		}
		return nil
	}

	return errors.Errorf("no secret with name %q available", secret.Name)
}

func ensureSecret(client rancher.BridgeClient, secret *schema.Secret) error {
	old, err := lookupSecret(client, secret)
	if err != nil {
		return errors.Annotate(err, "lookupSecret")
	}

	sec := rancherClient.Secret{
		Name:  secret.Name,
		Value: secret.Value,
	}

	if old != nil {
		_, err := client.UpdateSecret(old, sec.Value)
		if err != nil {
			return errors.Annotate(err, "UpdateSecret")
		}

		return nil
	}

	_, err = client.CreateSecret(&sec)
	if err != nil {
		return errors.Annotate(err, "CreateSecret")
	}

	return nil
}

func lookupSecret(client rancher.BridgeClient, secret *schema.Secret) (*rancherClient.Secret, error) {
	coll, err := client.ListSecrets(nil)
	if err != nil {
		return nil, errors.Annotate(err, "ListSecrets")
	}

	for _, s := range coll.Data {
		if secret.Name == s.Name {
			return &s, nil
		}
	}
	return nil, nil
}
