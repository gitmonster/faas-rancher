// Copyright (c) Ken Fukuyama 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package rancher

import (
	"github.com/juju/errors"
	client "github.com/rancher/go-rancher/v2"
	"github.com/sirupsen/logrus"
)

var (
	logger = logrus.WithField("package", "rancher")
)

// BridgeClient is the interface for Rancher API
type BridgeClient interface {
	ListServices() ([]client.Service, error)
	FindServiceByName(name string) (*client.Service, error)
	CreateService(spec *client.Service) (*client.Service, error)
	DeleteService(spec *client.Service) error
	UpdateService(spec *client.Service, updates map[string]string) (*client.Service, error)
	UpgradeService(spec *client.Service, upgrade *client.ServiceUpgrade) (*client.Service, error)
	FinishUpgradeService(spec *client.Service) (*client.Service, error)
	CreateSecret(spec *client.Secret) (*client.Secret, error)
	ListSecrets(listOpts *client.ListOpts) (*client.SecretCollection, error)
	DeleteSecret(spec *client.Secret) error
	UpdateSecret(spec *client.Secret, update interface{}) (*client.Secret, error)
	CreateSecretReference(spec *client.SecretReference) (*client.SecretReference, error)
}

// Client is the REST client type
type Client struct {
	rancherClient    *client.RancherClient
	config           *Config
	functionsStackID string
}

// NewClientForConfig creates a new rancher REST client
func NewClientForConfig(config *Config) (BridgeClient, error) {
	c, newErr := client.NewRancherClient(&client.ClientOpts{
		Url:       config.CattleURL,
		AccessKey: config.CattleAccessKey,
		SecretKey: config.CattleSecretKey,
	})

	if newErr != nil {
		return nil, newErr
	}

	coll, listErr := c.Stack.List(&client.ListOpts{
		Filters: map[string]interface{}{
			"name": config.FunctionsStackName,
		},
	})

	if listErr != nil {
		return nil, errors.Annotate(listErr, "List")
	}

	var stack *client.Stack
	if len(coll.Data) == 0 {
		logger.Infof("stack named %s not found. creating...", config.FunctionsStackName)
		// create stack if not present
		reqStack := &client.Stack{
			Name: config.FunctionsStackName,
		}
		newStack, err := c.Stack.Create(reqStack)
		if err != nil {
			return nil, errors.Annotate(err, "Create")
		}
		logger.Info("stack creation complete")
		stack = newStack
	} else {
		stack = &coll.Data[0]
	}

	client := Client{
		rancherClient:    c,
		config:           config,
		functionsStackID: stack.Id,
	}

	return &client, nil

}

// ListServices lists rancher services inside the specified stack (set in config)
func (c *Client) ListServices() ([]client.Service, error) {
	services, err := c.rancherClient.Service.List(&client.ListOpts{
		Filters: map[string]interface{}{
			"stackId": c.functionsStackID,
		},
	})
	if err != nil {
		return nil, errors.Annotate(err, "List")
	}
	return services.Data, nil
}

// FindServiceByName finds a service based on its name
func (c *Client) FindServiceByName(name string) (*client.Service, error) {
	services, err := c.rancherClient.Service.List(&client.ListOpts{
		Filters: map[string]interface{}{
			"name": name,
		},
	})
	if err != nil {
		return nil, errors.Annotate(err, "List")
	}

	if len(services.Data) > 0 {
		return &services.Data[0], nil
	}

	return nil, nil
}

// CreateService creates a service inside rancher
func (c *Client) CreateService(spec *client.Service) (*client.Service, error) {
	spec.StackId = c.functionsStackID
	service, err := c.rancherClient.Service.Create(spec)
	if err != nil {
		return nil, errors.Annotate(err, "Create")
	}
	return service, nil
}

// DeleteService deletes the specified service in rancher
func (c *Client) DeleteService(spec *client.Service) error {
	err := c.rancherClient.Service.Delete(spec)
	if err != nil {
		return errors.Annotate(err, "Delete")
	}

	return nil
}

// UpdateService upgrades the specified service in rancher
func (c *Client) UpdateService(spec *client.Service, updates map[string]string) (*client.Service, error) {
	service, err := c.rancherClient.Service.Update(spec, updates)
	if err != nil {
		return nil, errors.Annotate(err, "Update")
	}
	return service, nil
}

// UpgradeService starts service upgrade of the specified service in rancher
func (c *Client) UpgradeService(spec *client.Service, upgrade *client.ServiceUpgrade) (*client.Service, error) {
	service, err := c.rancherClient.Service.ActionUpgrade(spec, upgrade)
	if err != nil {
		return nil, errors.Annotate(err, "ActionUpgrade")
	}
	return service, nil
}

// FinishUpgradeService finishes service upgrade of the specified service in rancher
func (c *Client) FinishUpgradeService(spec *client.Service) (*client.Service, error) {
	service, err := c.rancherClient.Service.ActionFinishupgrade(spec)
	if err != nil {
		return nil, errors.Annotate(err, "ActionFinishupgrade")
	}
	return service, nil
}

// CreateSecret creates a rancher secret
func (c *Client) CreateSecret(spec *client.Secret) (*client.Secret, error) {
	secret, err := c.rancherClient.Secret.Create(spec)
	if err != nil {
		return nil, errors.Annotate(err, "Create")
	}
	return secret, nil
}

// ListSecrets lists rancher secrets
func (c *Client) ListSecrets(listOpts *client.ListOpts) (*client.SecretCollection, error) {
	coll, err := c.rancherClient.Secret.List(listOpts)
	if err != nil {
		return nil, errors.Annotate(err, "List")
	}
	return coll, nil
}

// DeleteSecret deletes a rancher secret
func (c *Client) DeleteSecret(spec *client.Secret) error {
	if err := c.rancherClient.Secret.Delete(spec); err != nil {
		return errors.Annotate(err, "Delete")
	}

	return nil
}

// UpdateSecret updates a rancher secret
func (c *Client) UpdateSecret(spec *client.Secret, update interface{}) (*client.Secret, error) {
	secret, err := c.rancherClient.Secret.Update(spec, update)
	if err != nil {
		return nil, errors.Annotate(err, "Update")
	}
	return secret, nil
}

// UpdateSecret updates a rancher secret
func (c *Client) CreateSecretReference(spec *client.SecretReference) (*client.SecretReference, error) {
	secret, err := c.rancherClient.SecretReference.Create(spec)
	if err != nil {
		return nil, errors.Annotate(err, "Create")
	}
	return secret, nil
}
