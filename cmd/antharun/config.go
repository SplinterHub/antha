// /antharun/config.go: Part of the Antha language
// Copyright (C) 2015 The Antha authors. All rights reserved.
//
// This program is free software; you can redistribute it and/or
// modify it under the terms of the GNU General Public License
// as published by the Free Software Foundation; either version 2
// of the License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program; if not, write to the Free Software
// Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA  02110-1301, USA.
//
// For more information relating to the software or licensing issues please
// contact license@antha-lang.org or write to the Antha team c/o
// Synthace Ltd. The London Bioscience Innovation Centre
// 1 Royal College St, London NW1 0NH UK

package main

import (
	"encoding/json"
	"fmt"
	"github.com/SplinterHub/antha/antha/anthalib/factory"
	"github.com/SplinterHub/antha/antha/anthalib/wtype"
	"github.com/SplinterHub/antha/antha/component"
	"github.com/SplinterHub/antha/internal/github.com/ghodss/yaml"
	"reflect"
	"strings"
)

type Config struct {
	// Parameters to components; values are ComponentParamBlocks
	Parameters map[string]interface{}
	// Additional untyped configuration values
	Config map[string]interface{}
}

// Transfer fields from conf to params instantiating certain fields from a
// factory. conf is like the type of params but with wtype.FromFactory entries
// for certain fields/elements. Assign values of conf to params. Use
// factory.GetComponentByType to instantiate FromFactory instances and assign
// them to the appropriate entry in params.
//
// Assumes that params is a nil instance and conf is like type of params but
// with some values as FromFactory rather than the appropriate interface type
func (p *Config) transfer(cv reflect.Value, pv reflect.Value) error {
	factoryType := reflect.TypeOf(wtype.FromFactory{})
	ct := cv.Type()
	pt := pv.Type()
	switch {
	case ct.AssignableTo(pt):
		pv.Set(cv)
	case ct == factoryType:
		v := cv.Interface().(wtype.FromFactory)
		terms := strings.Split(v.String, ":")
		if len(terms) != 2 {
			return fmt.Errorf("cannot parse factory string: %s", v.String)
		}
		var newV interface{}
		switch terms[0] {
		case "component":
			newV = factory.GetComponentByType(terms[1])
		case "tipbox":
			newV = factory.GetTipboxByType(terms[1])
		case "plate":
			newV = factory.GetPlateByType(terms[1])
		default:
			return fmt.Errorf("cannot parse factory string: %s", v.String)
		}
		nv := reflect.ValueOf(newV)
		nt := nv.Type()
		if !nt.AssignableTo(pt) {
			return fmt.Errorf("cannot convert %v to %v", nt, pt)
		}
		pv.Set(nv)
	default:
		pk := pt.Kind()
		ck := ct.Kind()

		if pk != ck {
			return fmt.Errorf("cannot convert %v to %v", ct, pt)
		}
		switch ck {
		case reflect.Slice:
			cvl := cv.Len()
			newSlice := reflect.MakeSlice(pt, cvl, cvl)
			pv.Set(newSlice)
			for i := 0; i < cvl; i += 1 {
				if err := p.transfer(cv.Index(i), pv.Index(i)); err != nil {
					return err
				}
			}
		case reflect.Map:
			newMap := reflect.MakeMap(pt)
			pv.Set(newMap)
			for _, key := range cv.MapKeys() {
				if err := p.transfer(pv.MapIndex(key), pv.MapIndex(key)); err != nil {
					return err
				}
			}
		case reflect.Struct:
			ctn := ct.NumField()
			for i := 0; i < ctn; i += 1 {
				if err := p.transfer(cv.Field(i), pv.Field(i)); err != nil {
					return err
				}
			}
		case reflect.Ptr:
			if err := p.transfer(cv.Elem(), pv.Elem()); err != nil {
				return err
			}
		default:
			return fmt.Errorf("cannot convert %v to %v", ct, pt)
		}
	}

	return nil
}

type ConfigFile struct {
	// Parameters to components; values are ComponentConfigs
	Parameters map[string]interface{}
	// Additional untyped configuration values
	Config map[string]interface{}
}

func NewConfig(yml []byte, wf *Workflow) (*Config, error) {
	conf := &Config{}
	rawParameters := make(map[string]interface{})
	conf.Parameters = make(map[string]interface{})
	for k, v := range wf.Processes() {
		c, ok := v.(component.Component)
		if !ok {
			return nil, fmt.Errorf("component %v is not an antha component", k)
		}
		rawParameters[k] = c.NewConfig()
		conf.Parameters[k] = c.NewParamBlock()
	}

	cf := &ConfigFile{}
	if err := yaml.Unmarshal(yml, cf); err != nil {
		return nil, err
	}
	conf.Config = cf.Config

	// Convert raw map[string]interface{} to ComponentConfig
	for k, v := range cf.Parameters {
		if _, ok := rawParameters[k]; !ok {
			return nil, fmt.Errorf("parameter for unknown component %v", k)
		}
		bytes, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(bytes, rawParameters[k]); err != nil {
			return nil, err
		}
	}

	// Convert ComponentConfig to ComponentParamBlock
	for k, v := range rawParameters {
		vv := reflect.ValueOf(v)
		pv := reflect.ValueOf(conf.Parameters[k])
		if err := conf.transfer(vv, pv); err != nil {
			return nil, err
		}
	}

	return conf, nil
}
