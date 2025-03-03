// anthalib//liquidhandling/serialize.go: Part of the Antha language
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

package liquidhandling

import (
	"encoding/json"
	"fmt"
	"github.com/antha-lang/antha/antha/anthalib/driver/liquidhandling"
	"github.com/antha-lang/antha/antha/anthalib/wtype"
	"github.com/antha-lang/antha/antha/anthalib/wutil"
	"strconv"
)

// functions to deal with how to serialize / deserialize the relevant objects.
// Despite the availability of good JSON serialization in Go it is necessary
// to include this to allow object structure to be sensibly defined for
// runtime purposes without making the network traffic too heavy

// marshal / unmarshal methods for the top-level lhrequest class

type SLHRequest struct {
	ID                         string
	Output_solutions           map[string]*wtype.LHSolution
	Input_solutions            map[string][]*wtype.LHComponent
	Plates                     map[string]*wtype.LHPlate
	Tips                       []*wtype.LHTipbox
	Locats                     []string
	Setup                      wtype.LHSetup
	InstructionSet             *liquidhandling.RobotInstructionSet
	Instructions               []liquidhandling.TerminalRobotInstruction
	Robotfn                    string
	Input_assignments          map[string][]string
	Output_plates              map[string]*wtype.LHPlate
	Input_platetypes           []*wtype.LHPlate
	Input_major_group_layouts  map[string][]string
	Input_minor_group_layouts  [][]string
	Input_plate_layout         map[string]string
	Output_platetype           *wtype.LHPlate
	Output_major_group_layouts map[string][]string
	Output_minor_group_layouts [][]string
	Output_plate_layout        map[string]string
	Plate_lookup               map[string]string
	Stockconcs                 map[string]float64
	Policies                   *liquidhandling.LHPolicyRuleSet
}

func (req *LHRequest) MarshalJSON() ([]byte, error) {
	new_input_major_layouts := make(map[string][]string, len(req.Input_major_group_layouts))

	for k, v := range req.Input_major_group_layouts {
		new_input_major_layouts[strconv.Itoa(k)] = v
	}

	new_input_plate_layout := make(map[string]string, len(req.Input_plate_layout))

	for k, v := range req.Input_plate_layout {
		new_input_plate_layout[strconv.Itoa(k)] = v
	}
	new_output_major_layouts := make(map[string][]string, len(req.Output_major_group_layouts))

	for k, v := range req.Output_major_group_layouts {
		new_output_major_layouts[strconv.Itoa(k)] = v
	}
	new_output_plate_layout := make(map[string]string, len(req.Output_plate_layout))

	for k, v := range req.Output_plate_layout {
		new_output_plate_layout[strconv.Itoa(k)] = v
	}

	slhr := SLHRequest{req.ID, req.Output_solutions, req.Input_solutions, req.Plates, req.Tips, req.Locats, req.Setup, req.InstructionSet, req.Instructions, req.Robotfn, req.Input_assignments, req.Output_plates, req.Input_platetypes, new_input_major_layouts, req.Input_minor_group_layouts, new_input_plate_layout, req.Output_platetype, new_output_major_layouts, req.Output_minor_group_layouts, new_output_plate_layout, req.Plate_lookup, req.Stockconcs, req.Policies}

	return json.Marshal(slhr)
}

func (req *LHRequest) UnmarshalJSON(ar []byte) error {
	var slhr SLHRequest
	e := json.Unmarshal(ar, req)

	fmt.Println("ERR: ", e)

	e = json.Unmarshal(ar, slhr)

	req.Input_major_group_layouts = make(map[int][]string, len(slhr.Input_major_group_layouts))

	for k, v := range slhr.Input_major_group_layouts {
		req.Input_major_group_layouts[wutil.ParseInt(k)] = v
	}

	req.Input_plate_layout = make(map[int]string, len(slhr.Input_plate_layout))

	for k, v := range slhr.Input_plate_layout {
		req.Input_plate_layout[wutil.ParseInt(k)] = v
	}

	req.Output_major_group_layouts = make(map[int][]string, len(slhr.Output_major_group_layouts))

	for k, v := range slhr.Output_major_group_layouts {
		req.Output_major_group_layouts[wutil.ParseInt(k)] = v
	}

	req.Output_plate_layout = make(map[int]string, len(slhr.Output_plate_layout))

	for k, v := range slhr.Output_plate_layout {
		req.Output_plate_layout[wutil.ParseInt(k)] = v
	}

	return e
}
