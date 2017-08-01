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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"text/template"

	"github.com/ghodss/yaml"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/nodeup"
	"k8s.io/kops/pkg/model/resources"
	"k8s.io/kops/upup/pkg/fi"
)

// BootstrapScript creates the bootstrap script
type BootstrapScript struct {
	NodeUpSource        string
	NodeUpSourceHash    string
	NodeUpConfigBuilder func(ig *kops.InstanceGroup) (*nodeup.Config, error)
}

func (b *BootstrapScript) ResourceNodeUp(ig *kops.InstanceGroup, cs *kops.ClusterSpec) (*fi.ResourceHolder, error) {
	if ig.Spec.Role == kops.InstanceGroupRoleBastion {
		// Bastions are just bare machines (currently), used as SSH jump-hosts
		return nil, nil
	}

	var igSpec string
	if cs.EnableClusterSpecInUserData {
		spec := make(map[string]interface{})

		spec["docker"] = cs.Docker
		spec["kubeProxy"] = cs.KubeProxy
		spec["kubelet"] = cs.Kubelet
		spec["cloudConfig"] = cs.CloudConfig

		if ig.IsMaster() {
			spec["kubeAPIServer"] = cs.KubeAPIServer
			spec["kubeControllerManager"] = cs.KubeControllerManager
			spec["kubeScheduler"] = cs.KubeScheduler
			spec["masterKubelet"] = cs.MasterKubelet
		}

		j, err := json.Marshal(spec)
		if err != nil {
			return nil, err
		}
		content, err := yaml.JSONToYAML(j)
		if err != nil {
			return nil, err
		}

		if cs.EnableClusterSpecHash {
			igSpec = base64.StdEncoding.EncodeToString(content)
		} else {
			igSpec = string(content)
		}
	}

	context := map[string]interface{}{
		"IncludeClusterSpec": cs.EnableClusterSpecInUserData,
		"ClusterSpecContent": igSpec,
	}

	functions := template.FuncMap{
		"NodeUpSource": func() string {
			return b.NodeUpSource
		},
		"NodeUpSourceHash": func() string {
			return b.NodeUpSourceHash
		},
		"KubeEnv": func() (string, error) {
			config, err := b.NodeUpConfigBuilder(ig)
			if err != nil {
				return "", err
			}

			data, err := kops.ToRawYaml(config)
			if err != nil {
				return "", err
			}

			return string(data), nil
		},

		// Pass in extra environment variables for user-defined S3 service
		"S3Env": func() string {
			if os.Getenv("S3_ENDPOINT") != "" {
				return fmt.Sprintf("export S3_ENDPOINT=%s\nexport S3_REGION=%s\nexport S3_ACCESS_KEY_ID=%s\nexport S3_SECRET_ACCESS_KEY=%s\n",
					os.Getenv("S3_ENDPOINT"),
					os.Getenv("S3_REGION"),
					os.Getenv("S3_ACCESS_KEY_ID"),
					os.Getenv("S3_SECRET_ACCESS_KEY"))
			}
			return ""
		},

		"AWS_REGION": func() string {
			if os.Getenv("AWS_REGION") != "" {
				return fmt.Sprintf("export AWS_REGION=%s\n",
					os.Getenv("AWS_REGION"))
			}
			return ""
		},
	}

	templateResource, err := NewTemplateResource("nodeup", resources.AWSNodeUpTemplate, functions, context)
	if err != nil {
		return nil, err
	}
	return fi.WrapResource(templateResource), nil
}
