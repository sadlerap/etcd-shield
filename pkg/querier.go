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
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type Querier struct {
	state      StateManager
	prometheus PromQuery
	config     Config
}

func NewQuerier(prom PromQuery, state StateManager, config Config) *Querier {
	querier := Querier{
		prometheus: prom,
		state:      state,
		config:     config,
	}

	return &querier
}

var _ manager.Runnable = &Querier{}
var _ manager.LeaderElectionRunnable = &Querier{}

func (q *Querier) NeedLeaderElection() bool {
	// for now, only one reader/writer to prometheus
	return true
}

func (q *Querier) Start(ctx context.Context) error {
	l := logr.FromContextOrDiscard(ctx)
	ticker := time.NewTicker(q.config.WaitTime.Duration)
	for {
		select {
		case <-ticker.C:
			err := q.Process(ctx)
			if err != nil {
				l.Error(err, "failed to process state")
			}
		case <-ctx.Done():
			return nil
		}
	}
}

func (q *Querier) Process(ctx context.Context) error {
	var query string
	l := logr.FromContextOrDiscard(ctx)

	// step 1: read state and determine which query to run
	currentState, err := q.state.ReadConfig(ctx)
	if err != nil {
		return err
	}

	// step 2: check if we need to continue allowing or denying requests
	if currentState == true {
		query = q.config.DisableIngressQuery
		l.Info("checking to disable pipelinerun ingress")
	} else {
		query = q.config.EnableIngressQuery
		l.Info("checking to enable pipelinerun ingress")
	}
	result, err := q.prometheus.Query(ctx, query)
	if err != nil {
		return err
	}
	newState := result != "0"
	l.Info("pipelinerun ingress status", "status", newState)

	// step 3: update the webhooks
	err = q.state.WriteConfig(ctx, newState)
	if err != nil {
		return err
	}

	return nil
}
