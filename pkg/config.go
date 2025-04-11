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
	"os"

	"github.com/go-logr/logr"
	"github.com/prometheus/common/config"
	"sigs.k8s.io/yaml"
)

type Config struct {
	// DestName is the name of the ConfigMap to write our state to
	DestName string `json:"destName"`

	// DestNamespace is the namespace of the ConfigMap to write our state to
	DestNamespace string `json:"destNamespace"`

	// Prometheus specifies how to talk to a service that speaks the prometheus HTTP API.
	Prometheus PrometheusConfig `json:"prometheus"`

	// WaitTime is how long we'll wait before checking prometheus again.
	WaitTime Duration `json:"waitTime"`
}

type PrometheusConfig struct {
	// Address to make prometheus queries to
	Address string `json:"address"`

	// EnableIngressQuery is the prometheus query to run to determine if `PipelineRun`
	// ingress will be allowed.  Should be mutually exclusive with SetQuery.
	AlertName string `json:"alertName"`

	// Config details the connection information to the prometheus server
	Config config.HTTPClientConfig `json:"config"`
}

func GetConfig(l logr.Logger, path string) (*Config, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		l.Error(err, "failed to read config", "path", path)
		return nil, err
	}

	cfg := Config{}
	err = yaml.Unmarshal(contents, &cfg)
	if err != nil {
		l.Error(err, "failed to deserialize config", "path", path)
		return nil, err
	}

	return &cfg, nil
}
