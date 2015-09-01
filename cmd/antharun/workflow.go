// /antharun/workflow.go: Part of the Antha language
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
	"bytes"
	"fmt"
	"github.com/SplinterHub/antha/antha/anthalib/execution"
	"github.com/SplinterHub/antha/antha/component/lib"
	"github.com/SplinterHub/antha/antha/execute"
	"github.com/SplinterHub/antha/flow"
	"github.com/SplinterHub/antha/internal/github.com/nu7hatch/gouuid"
	"reflect"
)

type Workflow struct {
	flow.Graph
	InPorts  []flow.FullPortName
	OutPorts []flow.FullPortName
}

type Errors []error

func (e Errors) Error() string {
	var buf bytes.Buffer
	for _, v := range e {
		buf.WriteString(v.Error())
	}
	return buf.String()
}

type WorkflowRun struct {
	ID       uuid.UUID
	Outs     map[flow.FullPortName]chan execute.ThreadParam
	Ins      map[flow.FullPortName]chan execute.ThreadParam
	Errors   chan error
	Done     chan bool
	Messages chan execute.ThreadParam
}

func NewWorkflowRun(id uuid.UUID, wf *Workflow, cf *Config) (*WorkflowRun, error) {
	params := make(map[flow.FullPortName]interface{})
	for _, port := range wf.InPorts {
		param, ok := cf.Parameters[port.Proc]
		if !ok {
			return nil, fmt.Errorf("required parameter not found %v", port)
		}

		// Need to unpack fields to map entries; use reflection to avoid
		// copying pointer based structures
		pv := reflect.ValueOf(param)
		fv := pv.Elem().FieldByName(port.Port)
		if !fv.IsValid() {
			return nil, fmt.Errorf("required parameter not found %v", port)
		}
		reflect.ValueOf(params).SetMapIndex(reflect.ValueOf(port), fv)
	}

	ins := make(map[flow.FullPortName]chan execute.ThreadParam)
	outs := make(map[flow.FullPortName]chan execute.ThreadParam)
	for _, port := range wf.InPorts {
		ins[port] = make(chan execute.ThreadParam)
		wf.Graph.SetInPort(fmt.Sprintf("%s.%s", port.Proc, port.Port), ins[port])
	}
	for _, port := range wf.OutPorts {
		outs[port] = make(chan execute.ThreadParam)
		wf.Graph.SetOutPort(fmt.Sprintf("%s.%s", port.Proc, port.Port), outs[port])
	}

	messages := make(chan execute.ThreadParam)
	errors := make(chan error)
	done := make(chan bool)

	// Point of no return... start running workflow
	flow.RunNet(&wf.Graph)

	for _, ch := range outs {
		go func(ch chan execute.ThreadParam) {
			for v := range ch {
				messages <- v
			}
		}(ch)
	}

	tid := execute.ThreadID(id.String())
	go func() {
		defer func() {
			for _, ch := range ins {
				close(ch)
			}
		}()

		for port, v := range params {
			param := execute.ThreadParam{
				Value: v,
				ID:    tid,
			}
			ins[port] <- param
		}
	}()

	go func() {
		<-wf.Graph.Wait()
		close(errors)
		close(messages)
		done <- true
	}()

	return &WorkflowRun{
		ID:       id,
		Outs:     outs,
		Ins:      ins,
		Messages: messages,
		Errors:   errors,
		Done:     done,
	}, nil
}

// Runs a workflow for one sample
func (w *Workflow) Run(cf *Config) ([]interface{}, error) {
	id, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}

	tid := execute.ThreadID(id.String())
	ctx := execution.GetContext()
	ctx.ConfigService.SetConfig(tid, cf.Config)

	wr, err := NewWorkflowRun(*id, w, cf)
	if err != nil {
		return nil, err
	}

	var errors []error
	go func() {
		for v := range wr.Errors {
			if v != nil {
				errors = append(errors, v)
			}
		}
	}()

	var messages []interface{}
	go func() {
		for v := range wr.Messages {
//			messages = append(messages, fmt.Sprintf("%v", v.Value))
			messages = append(messages, v.Value)
		}
	}()

	<-wr.Done

	if len(errors) > 0 {
		return messages, Errors(errors)
	}
	return messages, nil
}

func NewWorkflow(js []byte) (*Workflow, error) {
	//g := flow.ParseJSON(js)
	g, err := flow.ParseJSON(js)
	if err != nil {
		return nil, err
	}
	wf := &Workflow{
		Graph: *g,
	}
	for _, port := range wf.GetUnboundInPorts() {
		wf.InPorts = append(wf.InPorts, port)
		wf.Graph.MapInPort(fmt.Sprintf("%s.%s", port.Proc, port.Port), port.Proc, port.Port)
	}
	for _, port := range wf.GetUnboundOutPorts() {
		wf.OutPorts = append(wf.OutPorts, port)
		wf.Graph.MapOutPort(fmt.Sprintf("%s.%s", port.Proc, port.Port), port.Proc, port.Port)
	}
	return wf, nil
}

func init() {
	cs := lib.GetComponents()
	for _, c := range cs {
		flow.Register(c.Name, c.Constructor)
	}
}
