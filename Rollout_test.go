// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package main_test

import (
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/v3/pkg/kusttest"
	plugins_test "sigs.k8s.io/kustomize/v3/pkg/plugins/test"
)

const (
	target = `
apiVersion: apps/v1
metadata:
  name: myDeploy
kind: Deployment
spec:
  replica: 2
  template:
    metadata:
      labels:
        old-label: old-value
    spec:
      containers:
      - name: nginx
        image: nginx
`
	targetWithNamespace = `
apiVersion: apps/v1
metadata:
  name: myDeploy
  namespace: namespace1
kind: Deployment
spec:
  replica: 2
  template:
    metadata:
      labels:
        old-label: old-value
    spec:
      containers:
      - name: nginx
        image: nginx
`
	targetNoschema = `
apiVersion: argoproj.io/v1alpha1
kind: Rollout
metadata:
  name: my-foo
spec:
  template:
    metadata:
      labels:
        old-label: old-value
    spec:
      containers:
      - name: nginx
        image: nginx
`
)

func TestStrategicMergeTransformerNoSchema(t *testing.T) {
	tc := plugins_test.NewEnvForTest(t).Set()
	defer tc.Reset()

	tc.BuildGoPlugin(
		"argoproj.io", "v1alpha1", "Rollout")

	th := kusttest_test.NewKustTestPluginHarness(t, "/app")

	th.WriteF("/app/patch.yaml", `
apiVersion: argoproj.io/v1alpha1
kind: Rollout
metadata:
  name: my-foo
spec:
  template:
    metadata:
      labels:
        new-label: new-value
    spec:
      containers:
      - name: nginx
        env:
        - name: why
          value: so serious
`)
	rm := th.LoadAndRunTransformer(`
apiVersion: argoproj.io/v1alpha1
kind: Rollout
metadata:
  name: notImportantHere
paths:
- patch.yaml
`, targetNoschema)

	th.AssertActualEqualsExpected(rm, `
apiVersion: argoproj.io/v1alpha1
kind: Rollout
metadata:
  name: my-foo
spec:
  template:
    metadata:
      labels:
        new-label: new-value
        old-label: old-value
    spec:
      containers:
      - env:
        - name: why
          value: so serious
        image: nginx
        name: nginx
`)
}
