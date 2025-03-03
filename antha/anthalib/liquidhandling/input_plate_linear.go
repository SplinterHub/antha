// /anthalib/liquidhandling/input_plate_linear.go: Part of the Antha language
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
	"github.com/antha-lang/antha/antha/anthalib/wtype"
	"github.com/antha-lang/antha/antha/anthalib/wunit"
	"github.com/antha-lang/antha/antha/anthalib/wutil"
	"github.com/antha-lang/antha/internal/github.com/lukpank/go-glpk/glpk"
)

func choose_plate_assignments(component_volumes map[string]wunit.Volume, plate_types []*wtype.LHPlate, weight_constraint map[string]float64) map[string]map[*wtype.LHPlate]int {

	//
	//	optimization is set up as follows:
	//
	//		let:
	//			Xk 	= 	Number of wells of type Y containing component Z (k = 1...YZ)
	//			Vy	= 	Working volume of well Y
	//			RVy	= 	Residual volume of well Y
	//			TVz	= 	Total volume of component Z required
	//			WRy	=	Rate of wells of type y in their plate
	//			PMax	=	Maximum number of plates
	//			WMax	= 	Maximum number of wells
	//
	//	Minimise:
	//			sum of Xk Wrv RVy
	//
	//	Subject to:
	//			sum of Xk Vy 	>= TVz	for each component Z
	//			sum of WRy Xk 	<= PMax
	//			sum of Xk	<= WMax
	//

	// setup

	lp := glpk.New()
	defer lp.Delete()

	lp.SetProbName("Assignments")
	lp.SetObjName("Z")

	// CHECK THIS

	lp.SetObjDir(glpk.MAX)

	// constraints:
	// 		total component volume
	//		number of plates
	//		number of wells
	n_rows := len(component_volumes) + 2

	lp.AddRows(n_rows)

	cur := 1

	component_order := make([]string, len(component_volumes))

	// volume constraints
	for cmp, vol := range component_volumes {
		component_order[cur-1] = cmp
		lp.SetRowBnds(cur, glpk.LO, vol.ConvertTo(wunit.ParsePrefixedUnit("ul")), 99999.0)
		cur += 1
	}

	// from now on we always have to use component_order

	// plate number constraints

	max_n_plates := weight_constraint["MAX_N_PLATES"] - 1.0
	lp.SetRowBnds(cur, glpk.UP, -99999.0, max_n_plates)
	cur += 1

	// well number constraints
	max_n_wells := weight_constraint["MAX_N_WELLS"]
	lp.SetRowBnds(cur, glpk.UP, -99999.0, max_n_wells)
	cur += 1

	// set up the matrix columns

	num_cols := len(component_order) * len(plate_types)
	lp.AddCols(num_cols)
	cur = 1

	for _, component := range component_order {
		for _, plate := range plate_types {
			// set up objective coefficient, column name and lower bound
			rv := plate.Welltype.ResidualVolume()
			coef := rv.ConvertTo(wunit.ParsePrefixedUnit("ul")) * weight_constraint["RESIDUAL_VOLUME_WEIGHT"]
			lp.SetObjCoef(cur, coef)
			lp.SetColName(cur, component+"_"+plate.PlateName)
			lp.SetColBnds(cur, glpk.LO, 0.0, 0.0)
			lp.SetColKind(cur, glpk.IV)
			cur += 1
		}
	}

	// now set up the constraint coefficients
	cur = 1

	ind := wutil.Series(0, num_cols)

	for c, _ := range component_order {
		row := make([]float64, num_cols+1)
		col := 0
		for i := 0; i < len(component_order); i++ {
			for j := 0; j < len(plate_types); j++ {
				vc := 0.0
				// pick out a set of columns according to which row we're on
				// volume constraints are the working volumes of the wells
				if c == i {
					vol := wunit.NewVolume(plate_types[j].Welltype.Vol, plate_types[j].Welltype.Vunit)
					rvol := wunit.NewVolume(plate_types[j].Welltype.Rvol, plate_types[j].Welltype.Vunit)
					vol.Subtract(&rvol)
					vc = vol.ConvertTo(wunit.ParsePrefixedUnit("ul"))
				}
				row[col+1] = vc
				col += 1
			}
		}
		lp.SetMatRow(cur, ind, row)
		cur += 1
	}

	// now the plate constraint

	row := make([]float64, num_cols+1)
	col := 1
	for i := 0; i < len(component_order); i++ {
		for j := 0; j < len(plate_types); j++ {
			// the coefficient here is 1/the number of this well type per plate
			r := 1.0 / float64(plate_types[j].Nwells)
			row[col] = r
			col += 1
		}
	}

	lp.SetMatRow(cur, ind, row)
	cur += 1

	// finally the well constraint

	row = make([]float64, num_cols+1)
	col = 1
	for i := 0; i < len(component_order); i++ {
		for j := 0; j < len(plate_types); j++ {
			// the number of wells is constrained so we just count
			row[col] = 1.0
			col += 1
		}
	}

	lp.SetMatRow(cur, ind, row)

	iocp := glpk.NewIocp()
	iocp.SetPresolve(true)
	iocp.SetMsgLev(0)
	lp.Intopt(iocp)

	// check constraints

	/*
		for i := 1; i <= n_rows; i++ {
			fmt.Println("ROW : ", i, " VAL : ", lp.MipRowVal(i))
		}
	*/
	// fill assignments - this is the number of wells in the plate of each type needed

	assignments := make(map[string]map[*wtype.LHPlate]int, len(component_volumes))

	cur = 1

	for i := 0; i < len(component_order); i++ {
		cmap := make(map[*wtype.LHPlate]int)
		for j := 0; j < len(plate_types); j++ {
			nwells := lp.MipColVal(cur)

			if nwells > 0 {
				//fmt.Println(component_order[i], " : ", plate_types[j].Type, " N WELLS: ", nwells)
				cmap[plate_types[j]] = int(nwells)
			}
			cur += 1
		}
		assignments[component_order[i]] = cmap
	}

	return assignments
}
