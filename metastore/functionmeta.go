package metastore

import (
	"github.com/gitmonster/faas-rancher/helper"
	"github.com/openfaas/faas/gateway/requests"
)

// FunctionMeta hold function metadata for metastore
type FunctionMeta struct {
	Service     string                 `json:"service"`
	Image       string                 `json:"image"`
	EnvProcess  string                 `json:"envProcess"`
	EnvVars     map[string]interface{} `json:"envVars"`
	Constraints []string               `json:"constraints"`
	Secrets     []string               `json:"secrets"`
	Labels      map[string]interface{} `json:"labels"`
	Annotations map[string]interface{} `json:"annotations"`
}

func (p *FunctionMeta) CreateFrom(req *requests.CreateFunctionRequest) *FunctionMeta {
	p.Service = req.Service
	p.Image = req.Image
	p.EnvProcess = req.EnvProcess
	p.Constraints = req.Constraints
	p.Secrets = req.Secrets
	p.EnvVars = helper.ToRancherMap(&req.EnvVars)
	p.Labels = helper.ToRancherMap(req.Labels)
	p.Annotations = helper.ToRancherMap(req.Annotations)

	return p
}
