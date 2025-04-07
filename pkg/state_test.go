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

	etcd_shield "github.com/konflux-ci/etcd-shield/pkg"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("Pkg/State/Read", func() {
	var client client.Client

	BeforeEach(func() {
		client = fake.NewClientBuilder().Build()
	})

	It("Should read a true state from the configmap", func(ctx context.Context) {
		configmap := v1.ConfigMap{
			Data: map[string]string{etcd_shield.CONFIG_KEY: "1"},
		}
		configmap.SetName("state")
		configmap.SetNamespace("etcd-shield")
		err := client.Create(ctx, &configmap)
		Expect(err).NotTo(HaveOccurred())

		state := etcd_shield.NewState(client, types.NamespacedName{Name: configmap.Name, Namespace: configmap.Namespace})
		allowed, err := state.ReadConfig(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(allowed).To(BeTrue())
	})

	It("Should read a false state from the configmap", func(ctx context.Context) {
		configmap := v1.ConfigMap{
			Data: map[string]string{etcd_shield.CONFIG_KEY: "0"},
		}
		configmap.SetName("state")
		configmap.SetNamespace("etcd-shield")
		err := client.Create(ctx, &configmap)
		Expect(err).NotTo(HaveOccurred())

		state := etcd_shield.NewState(client, types.NamespacedName{Name: configmap.Name, Namespace: configmap.Namespace})
		allowed, err := state.ReadConfig(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(allowed).To(BeFalse())
	})

	It("Should assume a default of true if the configmap isn't found", func(ctx context.Context) {
		state := etcd_shield.NewState(client, types.NamespacedName{Name: "state", Namespace: "etcd-shield"})
		allowed, err := state.ReadConfig(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(allowed).To(BeTrue())
	})
})

var _ = Describe("Pkg/State/Write", func() {
	var client client.Client
	var ref types.NamespacedName

	BeforeEach(func() {
		client = fake.NewClientBuilder().Build()
		ref = types.NamespacedName{Name: "state", Namespace: "etcd-shield"}
	})
	exist := func() {
		configmap := v1.ConfigMap{
			Data: map[string]string{},
		}
		configmap.SetName(ref.Name)
		configmap.SetNamespace(ref.Namespace)
		err := client.Create(context.Background(), &configmap)
		Expect(err).NotTo(HaveOccurred())
	}

	DescribeTable("write state", func(setup func(), allowed bool) {
		setup()
		state := etcd_shield.NewState(client, ref)
		err := state.WriteConfig(context.Background(), allowed)
		Expect(err).NotTo(HaveOccurred())

		read_state, err := state.ReadConfig(context.Background())
		Expect(err).NotTo(HaveOccurred())
		Expect(read_state).To(BeEquivalentTo(allowed))
	},
		Entry("existing configmap", exist, true),
		Entry("existing configmap", exist, false),
		Entry("non-existing configmap", func() {}, true),
		Entry("non-existing configmap", func() {}, false),
	)
})
