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
	"time"

	"github.com/go-logr/logr"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/config"
)

type PromQuery interface {
	Query(context.Context, string) (string, error)
}

type Prometheus struct {
	prometheus v1.API
}

var _ PromQuery = &Prometheus{}

func NewPrometheus(address string, cfg config.HTTPClientConfig) (PromQuery, error) {
	httpClient, err := config.NewClientFromConfig(cfg, "prometheus")
	if err != nil {
		return nil, err
	}
	client, err := api.NewClient(api.Config{
		Address: address,
		Client:  httpClient,
	})
	api := v1.NewAPI(client)
	return &Prometheus{prometheus: api}, nil
}

func (p *Prometheus) Query(ctx context.Context, query string) (string, error) {
	log := logr.FromContextOrDiscard(ctx)
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	result, warnings, err := p.prometheus.Query(ctx, query, time.Now(), v1.WithTimeout(5*time.Second))
	if err != nil {
		log.Error(err, "Error querying prometheus", "query", query)
		return "", err
	}
	if len(warnings) > 0 {
		log.V(4).Info("Warning from prometheus", "warning", warnings)
	}

	return result.String(), nil
}
