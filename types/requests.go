// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package types

type ScaleServiceRequest struct {
	ServiceName string `json:"serviceName"`
	Replicas    int64  `json:"replicas"`
}
