/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package model

import (
	"io/ioutil"
	"testing"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/nodeup"
)

func TestBootstrapUserData(t *testing.T) {
	cs := []struct {
		EnableClusterSpec bool
		HashClusterSpec   bool
		Role              kops.InstanceGroupRole
		ExpectedFilePath  string
	}{
		{
			EnableClusterSpec: false,
			HashClusterSpec:   false,
			Role:              "Master",
			ExpectedFilePath:  "tests/data/bootstrapscript_0.txt",
		},
		{
			EnableClusterSpec: true,
			HashClusterSpec:   false,
			Role:              "Master",
			ExpectedFilePath:  "tests/data/bootstrapscript_1.txt",
		},
		{
			EnableClusterSpec: true,
			HashClusterSpec:   false,
			Role:              "Node",
			ExpectedFilePath:  "tests/data/bootstrapscript_2.txt",
		},
		{
			EnableClusterSpec: true,
			HashClusterSpec:   true,
			Role:              "Master",
			ExpectedFilePath:  "tests/data/bootstrapscript_3.txt",
		},
	}

	for i, x := range cs {
		spec := makeTestCluster(x.EnableClusterSpec, x.HashClusterSpec).Spec
		group := makeTestInstanceGroup(x.Role)

		renderNodeUpConfig := func(ig *kops.InstanceGroup) (*nodeup.Config, error) {
			return &nodeup.Config{}, nil
		}

		bs := &BootstrapScript{
			NodeUpSource:        "NUSource",
			NodeUpSourceHash:    "NUSHash",
			NodeUpConfigBuilder: renderNodeUpConfig,
		}

		res, err := bs.ResourceNodeUp(group, &spec)
		if err != nil {
			t.Errorf("case %d failed to create nodeup resource. error: %s", i, err)
			continue
		}

		actual, err := res.AsString()
		if err != nil {
			t.Errorf("case %d failed to render nodeup resource. error: %s", i, err)
			continue
		}

		expectedBytes, err := ioutil.ReadFile(x.ExpectedFilePath)
		if err != nil {
			t.Fatalf("unexpected error reading ExpectedFilePath %q: %v", x.ExpectedFilePath, err)
		}

		if actual != string(expectedBytes) {
			t.Errorf("case %d, expected: %s. got: %s", i, string(expectedBytes), actual)
		}
	}
}

func makeTestCluster(enableClusterSpec bool, hashClusterSpec bool) *kops.Cluster {
	return &kops.Cluster{
		Spec: kops.ClusterSpec{
			EnableClusterSpecInUserData: enableClusterSpec,
			EnableClusterSpecHash:       hashClusterSpec,
			CloudProvider:               "aws",
			KubernetesVersion:           "1.7.0",
			Subnets: []kops.ClusterSubnetSpec{
				{Name: "test", Zone: "eu-west-1a"},
			},
			NonMasqueradeCIDR: "10.100.0.0/16",
			EtcdClusters: []*kops.EtcdClusterSpec{
				{
					Name: "main",
					Members: []*kops.EtcdMemberSpec{
						{
							Name:          "test",
							InstanceGroup: s("master-1"),
						},
					},
				},
			},
			NetworkCIDR: "10.79.0.0/24",
			CloudConfig: &kops.CloudConfiguration{
				NodeTags: s("something"),
			},
			Docker: &kops.DockerConfig{
				LogLevel: s("INFO"),
			},
			KubeAPIServer: &kops.KubeAPIServerConfig{
				Image: "CoreOS",
			},
			KubeControllerManager: &kops.KubeControllerManagerConfig{
				CloudProvider: "aws",
			},
			KubeProxy: &kops.KubeProxyConfig{
				CPURequest: "30m",
				FeatureGates: map[string]string{
					"AdvancedAuditing": "true",
				},
			},
			KubeScheduler: &kops.KubeSchedulerConfig{
				Image: "SomeImage",
			},
			Kubelet: &kops.KubeletConfigSpec{
				KubeconfigPath: "/etc/kubernetes/config.txt",
			},
			MasterKubelet: &kops.KubeletConfigSpec{
				KubeconfigPath: "/etc/kubernetes/config.cfg",
			},
		},
	}
}

func makeTestInstanceGroup(role kops.InstanceGroupRole) *kops.InstanceGroup {
	return &kops.InstanceGroup{
		Spec: kops.InstanceGroupSpec{
			Role: role,
		},
	}
}
