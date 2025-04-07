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

package etcd_shield_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/yaml"

	etcd_shield "github.com/konflux-ci/etcd-shield/pkg"
)

var _ = Describe("Pkg/Config", func() {
	var yamlConfig string

	BeforeEach(func() {
		cfg := etcd_shield.Config{
			DestName:      "etcd-shield-state",
			DestNamespace: "etcd-shield",
			Prometheus: etcd_shield.PrometheusConfig{
				AlertName: "foo",
				Address:   "prometheus.prometheus.svc:8080",
			},
			WaitTime: etcd_shield.NewDuration(15 * time.Second),
		}
		config, err := yaml.Marshal(&cfg)
		Expect(err).NotTo(HaveOccurred())

		yamlConfig = string(config)
	})

	It("Should deserialize a yaml config", func() {
		var config etcd_shield.Config
		err := yaml.Unmarshal([]byte(yamlConfig), &config)
		Expect(err).NotTo(HaveOccurred())

		Expect(config.DestName).To(Equal("etcd-shield-state"))
		Expect(config.DestNamespace).To(Equal("etcd-shield"))
		Expect(config.Prometheus.AlertName).To(Equal("foo"))
		Expect(config.Prometheus.Address).To(Equal("prometheus.prometheus.svc:8080"))
		Expect(config.WaitTime).To(Equal(etcd_shield.NewDuration(15 * time.Second)))
	})

	It("Should deserialize from a configmap", func() {
		configMap := v1.ConfigMap{
			Data: map[string]string{
				"config": yamlConfig,
			},
		}
		configMap.SetName("config")
		configMap.SetNamespace("etcd-shield")

		fakeClient := fake.NewFakeClient(&configMap)
		config, err := etcd_shield.GetConfig(context.Background(), fakeClient, types.NamespacedName{Namespace: configMap.Namespace, Name: configMap.Name})
		Expect(err).NotTo(HaveOccurred())
		Expect(config).NotTo(BeNil())

		Expect(config.DestName).To(Equal("etcd-shield-state"))
		Expect(config.DestNamespace).To(Equal("etcd-shield"))
		Expect(config.Prometheus.AlertName).To(Equal("foo"))
		Expect(config.Prometheus.Address).To(Equal("prometheus.prometheus.svc:8080"))
		Expect(config.WaitTime).To(Equal(etcd_shield.NewDuration(15 * time.Second)))
	})
})
