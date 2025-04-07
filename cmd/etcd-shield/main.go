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

package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/go-logr/logr"
	shield "github.com/konflux-ci/etcd-shield/pkg"
	tektonv1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func configRef() types.NamespacedName {
	name := os.Getenv("CONFIG_NAME")
	namespace := os.Getenv("CONFIG_NAMESPACE")
	return types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}
}

func SetupStateWithManager(manager manager.Manager, configRef types.NamespacedName) error {
	client := manager.GetClient()
	cfg, err := shield.GetConfig(context.Background(), client, configRef)
	if err != nil {
		return fmt.Errorf("failed to fetch config: %s", err)
	}

	prom, err := shield.NewPrometheus(cfg.Prometheus.Address, cfg.Prometheus.Config)
	if err != nil {
		return fmt.Errorf("failed to setup prometheus connection: %s", err)
	}

	state := shield.NewState(client, types.NamespacedName{
		Namespace: cfg.DestNamespace,
		Name:      cfg.DestName,
	})

	querier := shield.NewQuerier(prom, state, *cfg)

	manager.Add(querier)

	return nil
}

func main() {
	var enableLeaderElection bool
	var probeAddr string
	flag.BoolVar(&enableLeaderElection, "leader-elect", false, "Enable leader election.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")

	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)

	flag.Parse()
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	ctx := logr.NewContext(context.Background(), ctrl.Log)

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Logger:                 ctrl.Log,
		BaseContext:            func() context.Context { return ctx },
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "etcd-shield.konflux-ci.dev",
		HealthProbeBindAddress: probeAddr,
	})
	if err != nil {
		ctrl.Log.Error(err, "failed to create manager")
		os.Exit(1)
	}

	if err := SetupStateWithManager(mgr, configRef()); err != nil {
		ctrl.Log.Error(err, "failed to setup state with manager")
		os.Exit(1)
	}

	err = ctrl.NewWebhookManagedBy(mgr).
		For(&tektonv1.PipelineRun{}).
		WithValidator(&shield.Webhook{}).
		Complete()
	if err != nil {
		ctrl.Log.Error(err, "unable to setup pipelinerun webhooks")
		os.Exit(1)
	}

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		ctrl.Log.Error(err, "unable to setup healthz check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		ctrl.Log.Error(err, "unable to setup readyz check")
		os.Exit(1)
	}

	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		ctrl.Log.Error(err, "failed to run manager")
		os.Exit(1)
	}
}
