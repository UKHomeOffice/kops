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
	"fmt"
	"path/filepath"

	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
)

// s is a helper that builds a *string from a string value
func s(v string) *string {
	return fi.String(v)
}

// i64 is a helper that builds a *int64 from an int64 value
func i64(v int64) *int64 {
	return fi.Int64(v)
}

// buildCertificateRequest retrieves the certificate from a keystore
func buildCertificateRequest(c *fi.ModelBuilderContext, b *NodeupModelContext, name, path string) error {
	cert, err := b.KeyStore.Cert(name)
	if err != nil {
		return err
	}

	serialized, err := cert.AsString()
	if err != nil {
		return err
	}

	location := filepath.Join(b.PathSrvKubernetes(), fmt.Sprintf("%s.pem", name))
	if path != "" {
		location = path
	}

	c.AddTask(&nodetasks.File{
		Path:     location,
		Contents: fi.NewStringResource(serialized),
		Type:     nodetasks.FileType_File,
	})

	return nil
}

// buildPrivateKeyRequest retrieves a private key from the store
func buildPrivateKeyRequest(c *fi.ModelBuilderContext, b *NodeupModelContext, name, path string) error {
	k, err := b.KeyStore.PrivateKey(name)
	if err != nil {
		return err
	}

	serialized, err := k.AsString()
	if err != nil {
		return err
	}

	location := filepath.Join(b.PathSrvKubernetes(), fmt.Sprintf("%s-key.pem", name))
	if path != "" {
		location = path
	}

	c.AddTask(&nodetasks.File{
		Path:     location,
		Contents: fi.NewStringResource(serialized),
		Type:     nodetasks.FileType_File,
	})

	return nil
}
