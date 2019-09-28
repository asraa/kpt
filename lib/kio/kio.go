// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package kio contains low-level libraries for reading, modifying and writing
// Resource Configuration and packages.
package kio

import (
	"lib.kpt.dev/yaml"
)

// Reader reads ResourceNodes. Analogous to io.Reader.
type Reader interface {
	Read() ([]*yaml.RNode, error)
}

// ResourceNodeSlice is a collection of ResourceNodes.
// While ResourceNodeSlice has no inherent constraints on ordering or uniqueness, specific
// Readers, Filters or Writers may have constraints.
type ResourceNodeSlice []*yaml.RNode

var _ Reader = ResourceNodeSlice{}

func (o ResourceNodeSlice) Read() ([]*yaml.RNode, error) {
	return o, nil
}

// Writer writes ResourceNodes. Analogous to io.Writer.
type Writer interface {
	Write([]*yaml.RNode) error
}

type WriterFunc func([]*yaml.RNode) error

func (fn WriterFunc) Write(o []*yaml.RNode) error {
	return fn(o)
}

// Filter modifies a collection of Resource Configuration by returning the modified slice.
// When possible, Filters should be serializable to yaml so that they can be described
// declaratively as data.
//
// Analogous to http://www.linfo.org/filters.html
type Filter interface {
	Filter([]*yaml.RNode) ([]*yaml.RNode, error)
}

// FilterFunc can be used to implement Filter by defining a function.
type FilterFunc func([]*yaml.RNode) ([]*yaml.RNode, error)

func (fn FilterFunc) Filter(o []*yaml.RNode) ([]*yaml.RNode, error) {
	return fn(o)
}

// Pipeline reads Resource Configuration from a set of Inputs, applies some
// transformations, and writes the results to a set of Outputs.
//
// Analogous to http://www.linfo.org/pipes.html
type Pipeline struct {
	// Inputs provide sources for Resource Configuration to be read.
	Inputs []Reader `yaml:"inputs,omitempty"`

	// Filters are transformations applied to the Resource Configuration.
	// They are applied in the order they are specified.
	// Analogous to http://www.linfo.org/filters.html
	Filters []Filter `yaml:"filters,omitempty"`

	// Outputs are where the transformed Resource Configuration is written.
	Outputs []Writer `yaml:"outputs,omitempty"`
}

// Execute implements the Pipeline pipeline.
func (p Pipeline) Execute() error {
	var result []*yaml.RNode

	// read from the inputs
	for _, i := range p.Inputs {
		nodes, err := i.Read()
		if err != nil {
			return err
		}
		result = append(result, nodes...)
	}
	if len(result) == 0 {
		// no inputs to operate on
		return nil
	}

	// apply operations
	var err error
	for i := range p.Filters {
		op := p.Filters[i]
		result, err = op.Filter(result)
		if len(result) == 0 || err != nil {
			return err
		}
	}

	// write to the outputs
	for _, o := range p.Outputs {
		if err := o.Write(result); err != nil {
			return err
		}
	}
	return nil
}