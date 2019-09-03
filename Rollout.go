// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// go:generate go run sigs.k8s.io/kustomize/v3/cmd/pluginator
package main

import (
	"fmt"

	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"sigs.k8s.io/kustomize/v3/pkg/gvk"
	"sigs.k8s.io/kustomize/v3/pkg/ifc"
	"sigs.k8s.io/kustomize/v3/pkg/resmap"
	"sigs.k8s.io/kustomize/v3/pkg/resource"
	"sigs.k8s.io/kustomize/v3/pkg/types"
	"sigs.k8s.io/yaml"

	"github.com/dthomson25/argo-rollouts-kustomize/pkg/apis/rollouts/v1alpha1"
)

type plugin struct {
	ldr           ifc.Loader
	rf            *resmap.Factory
	loadedPatches []*resource.Resource
	Paths         []types.PatchStrategicMerge `json:"paths,omitempty" yaml:"paths,omitempty"`
	Patches       string                      `json:"patches,omitempty" yaml:"patches,omitempty"`
}

//noinspection GoUnusedGlobalVariable
var KustomizePlugin plugin

func (p *plugin) Config(ldr ifc.Loader, rf *resmap.Factory, config []byte) error {
	p.ldr = ldr
	p.rf = rf
	err := yaml.Unmarshal(config, p)
	if err != nil {
		return err
	}
	if len(p.Paths) == 0 && p.Patches == "" {
		return fmt.Errorf("empty file path and empty patch content")
	}
	if len(p.Paths) != 0 {
		for _, onePath := range p.Paths {
			res, err := p.rf.RF().SliceFromBytes([]byte(onePath))
			if err == nil {
				p.loadedPatches = append(p.loadedPatches, res...)
				continue
			}
			res, err = p.rf.RF().SliceFromPatches(ldr, []types.PatchStrategicMerge{onePath})
			if err != nil {
				return err
			}
			p.loadedPatches = append(p.loadedPatches, res...)
		}
	}
	if p.Patches != "" {
		res, err := p.rf.RF().SliceFromBytes([]byte(p.Patches))
		if err != nil {
			return err
		}
		p.loadedPatches = append(p.loadedPatches, res...)
	}

	if len(p.loadedPatches) == 0 {
		return fmt.Errorf(
			"patch appears to be empty; files=%v, Patch=%s", p.Paths, p.Patches)
	}
	return err
}

func (p *plugin) Transform(m resmap.ResMap) error {
	// patches, err := p.rf.MergePatches(p.loadedPatches)
	// if err != nil {
	// 	return err
	// }
	rolloutGvk := &gvk.Gvk{
		Group:   "argoproj.io",
		Version: "v1alpha1",
		Kind:    "Rollout",
	}
	for _, r := range m.Resources() {
		if r.OrgId().IsSelected(rolloutGvk) {
			orig, err := r.MarshalJSON()
			if err != nil {
				return err
			}
			patch, err := p.loadedPatches[0].MarshalJSON()
			// if string(patch) != "" {
			// 	return fmt.Errorf("PatchYaml: %s", patch)
			// }
			newYaml, err := strategicpatch.StrategicMergePatch(orig, patch, v1alpha1.Rollout{})
			if err != nil {
				return err
			}
			err = r.UnmarshalJSON(newYaml)
			if err != nil {
				return err
			}
		}
	}
	// for _, path := range p.FieldSpecs {
	// 	if !r.OrgId().IsSelected(&path.Gvk) {
	// 		continue
	// 	}
	// 	err := transformers.MutateField(
	// 		r.Map(), path.PathSlice(), false, p.mutateImage)
	// 	if err != nil {
	// 		return err
	// 	}
	// }
	// // Kept for backward compatibility
	// if err := p.findAndReplaceImage(r.Map()); err != nil && r.OrgId().Kind != `CustomResourceDefinition` {
	// 	return err
	// }
	// }
	// patches, err := p.rf.MergePatches(p.loadedPatches)
	// if err != nil {
	// 	return err
	// }
	// for _, patch := range patches.Resources() {
	// 	target, err := m.GetById(patch.OrgId())
	// 	if err != nil {
	// 		return err
	// 	}
	// 	err = target.Patch(patch.Kunstructured)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	// remove the resource from resmap
	// 	// when the patch is to $patch: delete that target
	// 	if len(target.Map()) == 0 {
	// 		err = m.Remove(target.CurId())
	// 		if err != nil {
	// 			return err
	// 		}
	// 	}
	// }
	return nil
}
