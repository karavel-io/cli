// Copyright 2021 The Karavel Project
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

package plan

import (
	"context"
	"fmt"
	"sync"

	"github.com/karavel-io/cli/internal/helmw"
	"github.com/karavel-io/cli/pkg/config"
	"github.com/karavel-io/cli/pkg/logger"
)

type Plan struct {
	components     map[string]*Component
	seenComponents map[string]string
	log            logger.Logger
}

func NewFromConfig(ctx context.Context, log logger.Logger, cfg *config.Config) (*Plan, error) {
	p := New(log)

	var wg sync.WaitGroup
	ch := make(chan error)
	components := make(chan Component, 10)
	done := make(chan bool)
	for _, c := range cfg.Components {
		wg.Add(1)
		go func(cc config.Component) {
			defer wg.Done()

			chartName := cc.Name
			if cc.ComponentName != "" {
				chartName = cc.ComponentName
			}

			log.Debugf("Loading component '%s'", chartName)
			meta, err := helmw.GetChartManifest(ctx, chartName, cc.Version, cc.Unstable)
			if err != nil {
				ch <- fmt.Errorf("failed to load component '%s': %w", chartName, err)
				return
			}
			comp, err := NewComponentFromChartMetadata(meta, cc.Unstable)
			if err != nil {
				ch <- fmt.Errorf("failed to instantiate component configuration: %w", err)
				return
			}
			if cc.ComponentName != "" {
				comp.name = cc.Name
			}

			comp.namespace = cc.Namespace
			comp.jsonParams = cc.JsonParams

			components <- comp

			log.Debugf("Loaded component %s", comp.DebugLabel())
		}(c)
	}

	go func() {
		wg.Wait()
		done <- true
	}()

	for {
		select {
		case err := <-ch:
			return nil, err
		case comp := <-components:
			err := p.AddComponent(comp)
			if err != nil {
				return nil, fmt.Errorf("failed to build plan from config: %w", err)
			}
		case <-done:
			return &p, nil
		}
	}
}

func New(log logger.Logger) Plan {
	return Plan{
		components:     map[string]*Component{},
		seenComponents: map[string]string{},
		log:            log,
	}
}

func (p *Plan) Components() []*Component {
	var cc []*Component
	for _, c := range p.components {
		cc = append(cc, c)
	}

	return cc
}

func (p *Plan) GetComponent(name string) *Component {
	return p.components[name]
}

func (p *Plan) AddComponent(c Component) error {
	if p.components[c.name] != nil {
		return fmt.Errorf("duplicate component '%s' found", c.name)
	}

	if other := p.seenComponents[c.ComponentName()]; c.singleton && other != "" {
		withAlias := ""
		if name := c.NameOverride(); name != "" {
			withAlias = fmt.Sprintf(" with alias '%s'", name)
		}
		return fmt.Errorf("component '%s'%s is a singleton, but another instance called '%s' is already declared", c.ComponentName(), withAlias, other)
	}

	p.components[c.name] = &c
	p.seenComponents[c.ComponentName()] = c.name
	return nil
}

func (p *Plan) HasComponent(name string) bool {
	return p.components[name] != nil
}

func (p *Plan) Validate() error {
	if err := p.checkDependencies(); err != nil {
		return err
	}

	if err := p.processIntegrations(); err != nil {
		return err
	}

	return nil
}

func (p *Plan) checkDependencies() error {
	for n, c := range p.components {
		for _, dn := range c.dependencies {
			if !p.HasComponent(dn) {
				return fmt.Errorf("missing dependency: component '%s' requires '%s'", n, dn)
			}
		}
	}
	return nil
}

func (p *Plan) processIntegrations() error {
	for _, c := range p.components {
		c.integrations = make(map[string]bool)
		for integ, dd := range c.integrationsDeps {
			active := len(dd) > 0
			for _, dn := range dd {
				active = active && p.HasComponent(dn)
			}
			c.integrations[integ] = active
		}
		if err := c.patchIntegrations(p.log); err != nil {
			return err
		}
	}
	return nil
}
