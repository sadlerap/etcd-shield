// Copyright 2025 Red Hat Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package etcd_shield

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/prometheus/common/config"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

type Config struct {
	// DestName is the name of the ConfigMap to write our state to
	DestName string `json:"destName"`

	// DestNamespace is the namespace of the ConfigMap to write our state to
	DestNamespace string `json:"destNamespace"`

	// DisableIngressQuery is the prometheus query to run to determine if PipelineRun
	// ingress is no longer allowed.  Should be mutually exclusive with
	// ResetQuery.
	DisableIngressQuery string `json:"disableIngressQuery"`

	// EnableIngressQuery is the prometheus query to run to determine if `PipelineRun`
	// ingress will be allowed.  Should be mutually exclusive with SetQuery.
	EnableIngressQuery string `json:"enableIngressQuery"`

	Prometheus PrometheusConfig `json:"prometheus"`

	// WaitTime is how long we'll wait before checking prometheus again.
	WaitTime Duration `json:"waitTime"`
}

type PrometheusConfig struct {
	// Address to make prometheus queries to
	Address string `json:"address"`

	// Config details the connection information to the prometheus server
	Config config.HTTPClientConfig `json:"config"`
}

func GetConfig(ctx context.Context, client client.Client, ref types.NamespacedName) (*Config, error) {
	l := logr.FromContextOrDiscard(ctx)
	l = l.WithValues("name", ref.Name, "namespace", ref.Namespace)
	config := v1.ConfigMap{}

	l.Info("Retrieving config")

	err := client.Get(ctx, ref, &config)
	if err != nil {
		l.Error(err, "Failed to retrieve config")
		return nil, err
	}

	cfg := Config{}
	err = yaml.Unmarshal([]byte(config.Data["config"]), &cfg)
	if err != nil {
		l.Error(err, "Failed to deserialize config")
		return nil, err
	}

	return &cfg, nil
}
