package handlers

import (
	"github.com/openfaas/faas/gateway/requests"
	client "github.com/rancher/go-rancher/v2"
)

func makeUpgradeSpec(request requests.CreateFunctionRequest) *client.ServiceUpgrade {
	envVars := make(map[string]interface{})
	for k, v := range request.EnvVars {
		envVars[k] = v
	}

	if len(request.EnvProcess) > 0 {
		envVars["fprocess"] = request.EnvProcess
	}

	// transfer request labels
	labels := make(map[string]interface{})
	if request.Labels != nil {
		for k, v := range *request.Labels {
			labels[k] = v
		}
	}

	labels[FaasFunctionLabel] = request.Service
	labels["io.rancher.container.pull_image"] = "always"

	launchConfig := &client.LaunchConfig{
		Environment: envVars,
		ImageUuid:   "docker:" + request.Image, // not sure if it's ok to just prefix with 'docker:'
		Labels:      labels,
	}

	// decodeAnnotations(&request, launchConfig)
	// spew.Dump(launchConfig)

	spec := &client.ServiceUpgrade{
		InServiceStrategy: &client.InServiceUpgradeStrategy{
			BatchSize:              1,
			StartFirst:             true,
			LaunchConfig:           launchConfig,
			SecondaryLaunchConfigs: []client.SecondaryLaunchConfig{},
		},
	}

	return spec
}

func makeServiceSpec(request requests.CreateFunctionRequest) *client.Service {
	envVars := make(map[string]interface{})
	for k, v := range request.EnvVars {
		envVars[k] = v
	}

	if len(request.EnvProcess) > 0 {
		envVars["fprocess"] = request.EnvProcess
	}

	// transfer request labels
	labels := make(map[string]interface{})
	if request.Labels != nil {
		for k, v := range *request.Labels {
			labels[k] = v
		}
	}

	labels[FaasFunctionLabel] = request.Service
	labels["io.rancher.container.pull_image"] = "always"

	launchConfig := &client.LaunchConfig{
		Environment: envVars,
		ImageUuid:   "docker:" + request.Image, // not sure if it's ok to just prefix with 'docker:'
		Labels:      labels,
	}

	// decodeAnnotations(&request, launchConfig)
	// spew.Dump(launchConfig)

	serviceSpec := &client.Service{
		Name:          request.Service,
		Scale:         1,
		StartOnCreate: true,
		LaunchConfig:  launchConfig,
	}

	return serviceSpec
}

// func decodeAnnotations(r *requests.CreateFunctionRequest, c *client.LaunchConfig) {
// 	if r.Annotations == nil {
// 		return
// 	}

// 	annot := *r.Annotations
// 	if link, ok := annot["link"]; ok {
// 		parts := strings.Split(link, ":")
// 		lnks := make(map[string]string)

// 		if len(parts) == 2 {
// 			lnks[parts[0]] = parts[1]
// 		} else {
// 			lnks[link] = link
// 		}
// 		c.Links = lnks
// 	}
// }
