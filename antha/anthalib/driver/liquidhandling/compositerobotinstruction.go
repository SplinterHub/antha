// /anthalib/driver/liquidhandling/compositerobotinstruction.go: Part of the Antha language
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

//
//
//
//
// contact license@antha-lang.org or write to the Antha team c/o

package liquidhandling

import (
	"errors"
	"fmt"
	"github.com/antha-lang/antha/antha/anthalib/wtype"
	"github.com/antha-lang/antha/antha/anthalib/wunit"
	//	"github.com/antha-lang/antha/antha/anthalib/wutil"
	"math"
)

type TransferParams struct {
	What       string
	PltFrom    string
	PltTo      string
	WellFrom   string
	WellTo     string
	Volume     *wunit.Volume
	FPlateType string
	TPlateType string
	FVolume    *wunit.Volume
	TVolume    *wunit.Volume
	Channel    *wtype.LHChannelParameter
}

func (tp TransferParams) ToString() string {
	return fmt.Sprintf("%s %s %s %s %s %s %s %s %s %s %s", tp.What, tp.PltFrom, tp.PltTo, tp.WellFrom, tp.WellTo, tp.Volume.ToString(), tp.FPlateType, tp.TPlateType, tp.FVolume.ToString(), tp.TVolume.ToString(), tp.Channel)
}

type MultiTransferParams struct {
	What       []string
	PltFrom    []string
	PltTo      []string
	WellFrom   []string
	WellTo     []string
	Volume     []*wunit.Volume
	FPlateType []string
	TPlateType []string
	FVolume    []*wunit.Volume
	TVolume    []*wunit.Volume
}

func NewMultiTransferParams(multi int) MultiTransferParams {
	var v MultiTransferParams
	v.What = make([]string, 0, multi)
	v.PltFrom = make([]string, 0, multi)
	v.PltTo = make([]string, 0, multi)
	v.WellFrom = make([]string, 0, multi)
	v.WellTo = make([]string, 0, multi)
	v.Volume = make([]*wunit.Volume, 0, multi)
	v.FVolume = make([]*wunit.Volume, 0, multi)
	v.TVolume = make([]*wunit.Volume, 0, multi)
	v.FPlateType = make([]string, 0, multi)
	v.TPlateType = make([]string, 0, multi)
	return v
}

func ChooseChannel(vol *wunit.Volume, prms *LHProperties) (*wtype.LHChannelParameter, string) {
	// very quick and dirty, basically just makes sure the minimum volume is
	// below the required volume

	var headchosen *wtype.LHHead = nil

	for _, head := range prms.HeadsLoaded {
		if head.Params.Minvol.LessThan(vol) || head.Params.Minvol.EqualTo(vol) {
			headchosen = head
			break
		}
	}

	if headchosen == nil {
		panic("NO HEAD CHOSEN")
	}

	// check if we need to change adaptor

	if headchosen.Adaptor.Params.Minvol.GreaterThan(vol) {
		panic("ADAPTOR CHANGE NEEDED BUT NOT IMPLEMENTED")
	}

	// now get the requisite tip
	// this is just a big bowl of wrong... </JeffGreene>
	// need to make this more dependent on what's actually there
	tiptype := ""
	// get the closest to the min vol
	d := 99999.0
	for _, tip := range prms.Tips {
		//if tip.MinVol.LessThan(vol) || tip.MinVol.EqualTo(vol) {
		//	tiptype = tip.Type
		//}

		dif := vol.ConvertTo(tip.MinVol.Unit()) - tip.MinVol.RawValue()
		if dif >= 0.0 && dif < d {
			tiptype = tip.Type
			d = dif
		}

	}

	if tiptype == "" {
		panic("NO TIP TYPE FOUND")
	}

	return headchosen.GetParams(), tiptype
}

type TransferInstruction struct {
	Type       int
	What       []string
	PltFrom    []string
	PltTo      []string
	WellFrom   []string
	WellTo     []string
	Volume     []*wunit.Volume
	FPlateType []string
	TPlateType []string
	FVolume    []*wunit.Volume
	TVolume    []*wunit.Volume
}

func (ti *TransferInstruction) ToString() string {
	s := fmt.Sprintf("%s ", Robotinstructionnames[ti.Type])

	for i := 0; i < len(ti.What); i++ {
		s += ti.ParamSet(i).ToString()
		s += "\n"
	}

	return s
}

func (ti *TransferInstruction) ParamSet(n int) TransferParams {
	return TransferParams{ti.What[n], ti.PltFrom[n], ti.PltTo[n], ti.WellFrom[n], ti.WellTo[n], ti.Volume[n], ti.FPlateType[n], ti.TPlateType[n], ti.FVolume[n], ti.TVolume[n], nil}
}

func NewTransferInstruction(what, pltfrom, pltto, wellfrom, wellto, fplatetype, tplatetype []string, volume, fvolume, tvolume []*wunit.Volume) *TransferInstruction {
	var v TransferInstruction
	v.Type = TFR
	v.What = what
	v.PltFrom = pltfrom
	v.PltTo = pltto
	v.WellFrom = wellfrom
	v.WellTo = wellto
	v.Volume = volume
	v.FPlateType = fplatetype
	v.TPlateType = tplatetype
	v.FVolume = fvolume
	v.TVolume = tvolume
	return &v
}
func (ins *TransferInstruction) InstructionType() int {
	return ins.Type
}

func (ins *TransferInstruction) GetParameter(name string) interface{} {
	switch name {
	case "LIQUIDCLASS":
		return ins.What
	case "VOLUME":
		return ins.Volume
	case "VOLUNT":
		return nil
	case "FROMPLATETYPE":
		return ins.FPlateType
	case "WELLFROMVOLUME":
		return ins.FVolume
	case "POSFROM":
		return ins.PltFrom
	case "POSTO":
		return ins.PltTo
	case "WELLFROM":
		return ins.WellFrom
	case "WELLTO":
		return ins.WellTo
	case "WELLTOVOLUME":
		return ins.TVolume
	case "TOPLATETYPE":
		return ins.TPlateType
	case "INSTRUCTIONTYPE":
		return ins.InstructionType()
	}
	return nil
}

func TransferVolumes(Vol, Min, Max wunit.Volume) []wunit.Volume {
	ret := make([]wunit.Volume, 0)

	vol := Vol.ConvertTo(Min.Unit())
	min := Min.RawValue()
	max := Max.RawValue()

	if vol < min {
		panic(errors.New(fmt.Sprintf("Error: %f below min vol %f", vol, min)))
	}

	if vol < max {
		ret = append(ret, Vol)
		return ret
	}

	// vol is > max, need to know by how much
	// if vol/max = n then we do n+1 equal transfers of vol / (n+1)
	// this should never be outside the range

	n, _ := math.Modf(vol / max)

	n += 1

	// should make sure of no rounding errors here... we want to
	// make sure these are within the resolution of the channel

	for i := 0; i < int(n); i++ {
		ret = append(ret, wunit.NewVolume(vol/n, Vol.Unit().PrefixedSymbol()))
	}

	return ret
}

func (vs VolumeSet) MaxMultiTransferVolume() *wunit.Volume {
	// the minimum volume in the set

	ret := vs.Vols[0]

	for _, v := range vs.Vols {
		if v.LessThan(ret) {
			ret = v
		}
	}

	return ret
}

// CHECK PLATES
func (ins *TransferInstruction) GetParallelSetsFor(channel *wtype.LHChannelParameter) [][]int {
	// if the channel is not multi just return nil

	if channel.Multi == 1 {
		return nil
	}

	ret := make([][]int, 0)

	tfrs := make([][][]int, 12)

	// hash out all transfers which are multiable

	for i, _ := range ins.What {
		var tcoord int = -1
		var fcoord int = -1
		wcFrom := wtype.MakeWellCoordsA1(ins.WellFrom[i])
		wcTo := wtype.MakeWellCoordsA1(ins.WellTo[i])

		if channel.Orientation == wtype.LHVChannel {
			// we hash on the X
			tcoord = wcTo.X
			fcoord = wcFrom.X
		} else {
			// horizontal orientation
			// hash on the Y
			tcoord = wcTo.Y
			fcoord = wcFrom.Y
		}

		a := tfrs[fcoord]

		if a == nil {
			a = make([][]int, 12)
		}

		t := a[tcoord]

		if t == nil {
			t = make([]int, 0, 12)
		}

		t = append(t, i)
		a[tcoord] = t
		tfrs[fcoord] = a
	}

	// now have we got any which are multiable?

	for _, a := range tfrs {
		for _, t := range a {
			if len(t) == channel.Multi {
				ret = append(ret, t)
			}
		}
	}

	if len(ret) == 0 {
		return nil
	}

	return ret
}

// helper thing

type VolumeSet struct {
	Vols []*wunit.Volume
}

func NewVolumeSet(n int) VolumeSet {
	var vs VolumeSet
	vs.Vols = make([]*wunit.Volume, n)
	for i := 0; i < n; i++ {
		v := (wunit.NewVolume(0.0, "ul"))
		vs.Vols[i] = &v
	}
	return vs
}

func (vs VolumeSet) Add(v *wunit.Volume) {
	for i := 0; i < len(vs.Vols); i++ {
		vs.Vols[i].Add(v)
	}
}

func (vs VolumeSet) Sub(v *wunit.Volume) []*wunit.Volume {
	ret := make([]*wunit.Volume, len(vs.Vols))
	for i := 0; i < len(vs.Vols); i++ {
		vs.Vols[i].Subtract(v)
		ret[i] = wunit.CopyVolume(v)
	}
	return ret
}

func (vs VolumeSet) SetEqualTo(v *wunit.Volume) {
	for i := 0; i < len(vs.Vols); i++ {
		vs.Vols[i] = wunit.CopyVolume(v)
	}
}

func (vs VolumeSet) GetACopy() []*wunit.Volume {
	r := make([]*wunit.Volume, len(vs.Vols))
	for i := 0; i < len(vs.Vols); i++ {
		r[i] = wunit.CopyVolume(vs.Vols[i])
	}
	return r
}

func (ins *TransferInstruction) Generate(policy *LHPolicyRuleSet, prms *LHProperties) []RobotInstruction {
	pol := policy.GetPolicyFor(ins)

	ret := make([]RobotInstruction, 0)

	// if we can multi we do this first

	if pol["CAN_MULTI"].(bool) {
		// break out the sets of parallel instructions

		// fix this HARD CODE here
		parallelsets := ins.GetParallelSetsFor(prms.HeadsLoaded[0].Params)
		mci := NewMultiChannelBlockInstruction()
		mci.Multi = prms.HeadsLoaded[0].Params.Multi // TODO Remove Hard code here
		mci.Prms = prms.HeadsLoaded[0].Params        // TODO Remove Hard code here

		for _, set := range parallelsets {
			// assemble the info

			vols := NewVolumeSet(len(set))
			fvols := NewVolumeSet(len(set))
			tvols := NewVolumeSet(len(set))
			What := make([]string, len(set))
			PltFrom := make([]string, len(set))
			PltTo := make([]string, len(set))
			WellFrom := make([]string, len(set))
			WellTo := make([]string, len(set))
			FPlateType := make([]string, len(set))
			TPlateType := make([]string, len(set))

			for i, s := range set {
				vols.Vols[i] = wunit.CopyVolume(ins.Volume[s])
				fvols.Vols[i] = wunit.CopyVolume(ins.FVolume[s])
				tvols.Vols[i] = wunit.CopyVolume(ins.TVolume[s])
				What[i] = ins.What[s]
				PltFrom[i] = ins.PltFrom[s]
				PltTo[i] = ins.PltTo[s]
				WellFrom[i] = ins.WellFrom[s]
				WellTo[i] = ins.WellTo[s]
				FPlateType[i] = ins.FPlateType[s]
				TPlateType[i] = ins.TPlateType[s]
			}

			// get the max transfer volume

			maxvol := vols.MaxMultiTransferVolume()

			// now set the vols for the transfer and remove this from the instruction's volume

			for i, _ := range vols.Vols {
				vols.Vols[i] = wunit.CopyVolume(maxvol)
				ins.Volume[set[i]].Subtract(maxvol)

				// set the from and to volumes for the relevant part of the instruction
				// NB -- this is a design issue which should probably be fixed: at the moment
				// if we have two instructions which refer to the same underlying well their
				// volume levels will not be in sync
				// therefore this implementation is not correct as regards changes of underlying
				// state
				//... instead the right thing would be for all of these instructions to reference
				// plate objects instead - this will work OK as long as we have a shared memory
				// system... otherwise we'll need to use channels
				ins.FVolume[set[i]].Subtract(maxvol)
				ins.TVolume[set[i]].Add(maxvol)
			}

			tp := NewMultiTransferParams(mci.Multi)
			tp.What = What
			tp.Volume = vols.Vols
			tp.FVolume = fvols.Vols
			tp.TVolume = tvols.Vols
			tp.PltFrom = PltFrom
			tp.PltTo = PltTo
			tp.WellFrom = WellFrom
			tp.WellTo = WellTo
			tp.FPlateType = FPlateType
			tp.TPlateType = TPlateType

			mci.AddTransferParams(tp)
		}

		if len(parallelsets) > 0 {
			ret = append(ret, mci)
		}
	}

	// mop up all the single instructions which are left
	sci := NewSingleChannelBlockInstruction()
	sci.Prms = prms.HeadsLoaded[0].Params // TODO Fix Hard Code Here

	for i, _ := range ins.What {
		if ins.Volume[i].LessThanFloat(0.001) {
			continue
		}

		var tp TransferParams
		tp.What = ins.What[i]
		tp.PltFrom = ins.PltFrom[i]
		tp.PltTo = ins.PltTo[i]
		tp.WellFrom = ins.WellFrom[i]
		tp.WellTo = ins.WellTo[i]
		tp.Volume = wunit.CopyVolume(ins.Volume[i])
		tp.FVolume = wunit.CopyVolume(ins.FVolume[i])
		tp.TVolume = wunit.CopyVolume(ins.TVolume[i])
		tp.FPlateType = ins.FPlateType[i]
		tp.TPlateType = ins.TPlateType[i]

		sci.AddTransferParams(tp)

		// make sure we keep volumes up to date

		ins.FVolume[i].Subtract(ins.Volume[i])
		ins.TVolume[i].Add(ins.Volume[i])
	}
	ret = append(ret, sci)
	return ret
}

type SingleChannelBlockInstruction struct {
	Type       int
	What       []string
	PltFrom    []string
	PltTo      []string
	WellFrom   []string
	WellTo     []string
	Volume     []*wunit.Volume
	FPlateType []string
	TPlateType []string
	FVolume    []*wunit.Volume
	TVolume    []*wunit.Volume
	Prms       *wtype.LHChannelParameter
}

func NewSingleChannelBlockInstruction() *SingleChannelBlockInstruction {
	var v SingleChannelBlockInstruction
	v.Type = SCB
	v.What = make([]string, 0)
	v.PltFrom = make([]string, 0)
	v.PltTo = make([]string, 0)
	v.WellFrom = make([]string, 0)
	v.WellTo = make([]string, 0)
	v.Volume = make([]*wunit.Volume, 0)
	v.FVolume = make([]*wunit.Volume, 0)
	v.TVolume = make([]*wunit.Volume, 0)
	v.FPlateType = make([]string, 0)
	v.TPlateType = make([]string, 0)
	return &v
}

func (ins *SingleChannelBlockInstruction) AddTransferParams(mct TransferParams) {
	ins.What = append(ins.What, mct.What)
	ins.PltFrom = append(ins.PltFrom, mct.PltFrom)
	ins.PltTo = append(ins.PltTo, mct.PltTo)
	ins.WellFrom = append(ins.WellFrom, mct.WellFrom)
	ins.WellTo = append(ins.WellTo, mct.WellTo)
	ins.Volume = append(ins.Volume, mct.Volume)
	ins.FPlateType = append(ins.FPlateType, mct.FPlateType)
	ins.TPlateType = append(ins.TPlateType, mct.TPlateType)
	ins.FVolume = append(ins.FVolume, mct.FVolume)
	ins.TVolume = append(ins.TVolume, mct.TVolume)
}
func (ins *SingleChannelBlockInstruction) InstructionType() int {
	return ins.Type
}

func (ins *SingleChannelBlockInstruction) GetParameter(name string) interface{} {
	switch name {
	case "LIQUIDCLASS":
		return ins.What
	case "VOLUME":
		return ins.Volume
	case "VOLUNT":
		return nil
	case "FROMPLATETYPE":
		return ins.FPlateType
	case "WELLFROMVOLUME":
		return ins.FVolume
	case "POSFROM":
		return ins.PltFrom
	case "POSTO":
		return ins.PltTo
	case "WELLFROM":
		return ins.WellFrom
	case "PARAMS":
		return ins.Prms
	case "PLATFORM":
		return ins.Prms.Name
	case "WELLTO":
		return ins.WellTo
	case "WELLTOVOLUME":
		return ins.TVolume
	case "TOPLATETYPE":
		return ins.TPlateType
	case "INSTRUCTIONTYPE":
		return ins.InstructionType()
	}
	return nil
}

func (ins *SingleChannelBlockInstruction) Generate(policy *LHPolicyRuleSet, prms *LHProperties) []RobotInstruction {
	pol := policy.GetPolicyFor(ins)
	ret := make([]RobotInstruction, 0)
	// get tips
	channel, tiptype := ChooseChannel(ins.Volume[0], prms)
	ret = append(ret, GetTips(tiptype, prms, channel, 1, false))
	n_tip_uses := 0

	for t := 0; t < len(ins.Volume); t++ {
		newchannel, newtiptype := ChooseChannel(ins.Volume[t], prms)
		tvs := TransferVolumes(*ins.Volume[t], *channel.Minvol, *channel.Maxvol)
		for _, vol := range tvs {
			// change tips if we need to
			if n_tip_uses > pol["TIP_REUSE_LIMIT"].(int) || channel != newchannel || newtiptype != tiptype {
				// maybe wrap this as a ChangeTips function call
				// these need parameters
				ret = append(ret, DropTips(tiptype, prms, channel, 1))
				ret = append(ret, GetTips(newtiptype, prms, newchannel, 1, false))
				tiptype = newtiptype
				channel = newchannel
				n_tip_uses = 0
			}

			stci := NewSingleChannelTransferInstruction()

			stci.What = ins.What[t]
			stci.PltFrom = ins.PltFrom[t]
			stci.PltTo = ins.PltTo[t]
			stci.WellFrom = ins.WellFrom[t]
			stci.WellTo = ins.WellTo[t]
			stci.Volume = &vol
			stci.FPlateType = ins.FPlateType[t]
			stci.TPlateType = ins.TPlateType[t]
			stci.FVolume = wunit.CopyVolume(ins.FVolume[t])
			stci.TVolume = wunit.CopyVolume(ins.TVolume[t])
			stci.Prms = ins.Prms

			ret = append(ret, stci)

			ins.FVolume[t].Subtract(&vol)
			ins.TVolume[t].Add(&vol)
			n_tip_uses += 1
		}

	}
	ret = append(ret, DropTips(tiptype, prms, channel, 1))

	return ret
}

type MultiChannelBlockInstruction struct {
	Type       int
	What       [][]string
	PltFrom    [][]string
	PltTo      [][]string
	WellFrom   [][]string
	WellTo     [][]string
	Volume     [][]*wunit.Volume
	FPlateType [][]string
	TPlateType [][]string
	FVolume    [][]*wunit.Volume
	TVolume    [][]*wunit.Volume
	Multi      int
	Prms       *wtype.LHChannelParameter
}

func NewMultiChannelBlockInstruction() *MultiChannelBlockInstruction {
	var v MultiChannelBlockInstruction
	v.Type = MCB
	v.What = make([][]string, 0)
	v.PltFrom = make([][]string, 0)
	v.PltTo = make([][]string, 0)
	v.WellFrom = make([][]string, 0)
	v.WellTo = make([][]string, 0)
	v.Volume = make([][]*wunit.Volume, 0)
	v.FPlateType = make([][]string, 0)
	v.TPlateType = make([][]string, 0)
	v.FVolume = make([][]*wunit.Volume, 0)
	v.TVolume = make([][]*wunit.Volume, 0)
	return &v
}

func (ins *MultiChannelBlockInstruction) AddTransferParams(mct MultiTransferParams) {
	ins.What = append(ins.What, mct.What)
	ins.PltFrom = append(ins.PltFrom, mct.PltFrom)
	ins.PltTo = append(ins.PltTo, mct.PltTo)
	ins.WellFrom = append(ins.WellFrom, mct.WellFrom)
	ins.WellTo = append(ins.WellTo, mct.WellTo)
	ins.Volume = append(ins.Volume, mct.Volume)
	ins.FPlateType = append(ins.FPlateType, mct.FPlateType)
	ins.TPlateType = append(ins.TPlateType, mct.TPlateType)
	ins.FVolume = append(ins.FVolume, mct.FVolume)
	ins.TVolume = append(ins.TVolume, mct.TVolume)
}

func (ins *MultiChannelBlockInstruction) InstructionType() int {
	return ins.Type
}

func (ins *MultiChannelBlockInstruction) GetParameter(name string) interface{} {
	switch name {
	case "LIQUIDCLASS":
		return ins.What
	case "VOLUME":
		return ins.Volume
	case "VOLUNT":
		return nil
	case "FROMPLATETYPE":
		return ins.FPlateType
	case "WELLFROMVOLUME":
		return ins.FVolume
	case "POSFROM":
		return ins.PltFrom
	case "POSTO":
		return ins.PltTo
	case "WELLFROM":
		return ins.WellFrom
	case "PARAMS":
		return ins.Prms
	case "PLATFORM":
		return ins.Prms.Name
	case "WELLTO":
		return ins.WellTo
	case "WELLTOVOLUME":
		return ins.TVolume
	case "TOPLATETYPE":
		return ins.TPlateType
	case "INSTRUCTIONTYPE":
		return ins.InstructionType()
	}
	return nil
}

func (ins *MultiChannelBlockInstruction) Generate(policy *LHPolicyRuleSet, prms *LHProperties) []RobotInstruction {
	pol := policy.GetPolicyFor(ins)
	ret := make([]RobotInstruction, 0)
	// get some tips
	channel, tiptype := ChooseChannel(ins.Volume[0][0], prms)
	ret = append(ret, GetTips(tiptype, prms, channel, ins.Multi, false))
	n_tip_uses := 0

	for t := 0; t < len(ins.Volume); t++ {
		tvols := NewVolumeSet(ins.Prms.Multi)
		vols := NewVolumeSet(ins.Prms.Multi)
		fvols := NewVolumeSet(ins.Prms.Multi)
		for i, _ := range ins.Volume[t] {
			fvols.Vols[i] = wunit.CopyVolume(ins.FVolume[t][i])
			tvols.Vols[i] = wunit.CopyVolume(ins.TVolume[t][i])
		}

		// choose tips
		newchannel, newtiptype := ChooseChannel(ins.Volume[0][0], prms)

		// load tips

		// split the transfer up
		// NB we assume all volumes are equal here;
		tvs := TransferVolumes(*ins.Volume[t][0], *newchannel.Minvol, *newchannel.Maxvol)

		for _, vol := range tvs {
			// enforce tip usage policy

			if n_tip_uses > pol["TIP_REUSE_LIMIT"].(int) || newchannel != channel || newtiptype != tiptype {
				// these need parameters
				ret = append(ret, DropTips(tiptype, prms, channel, ins.Multi))
				ret = append(ret, GetTips(newtiptype, prms, newchannel, ins.Multi, false))
				n_tip_uses = 0
			}

			mci := NewMultiChannelTransferInstruction()
			vols.SetEqualTo(&vol)
			mci.Volume = vols.GetACopy()
			mci.FVolume = fvols.GetACopy()
			mci.TVolume = tvols.GetACopy()
			mci.PltFrom = ins.PltFrom[t]
			mci.PltTo = ins.PltTo[t]
			mci.WellFrom = ins.WellFrom[t]
			mci.WellTo = ins.WellTo[t]
			mci.FPlateType = ins.FPlateType[t]
			mci.TPlateType = ins.TPlateType[t]
			mci.Prms = ins.Prms

			ret = append(ret, mci)

			tiptype = newtiptype
			channel = newchannel
			fvols.Sub(&vol)
			tvols.Add(&vol)
		}
	}

	// remove tips
	ret = append(ret, DropTips(tiptype, prms, channel, ins.Multi))

	return ret
}

type SingleChannelTransferInstruction struct {
	Type       int
	What       string
	PltFrom    string
	PltTo      string
	WellFrom   string
	WellTo     string
	Volume     *wunit.Volume
	FPlateType string
	TPlateType string
	FVolume    *wunit.Volume
	TVolume    *wunit.Volume
	Prms       *wtype.LHChannelParameter
}

func (scti *SingleChannelTransferInstruction) Params() TransferParams {
	var tp TransferParams
	tp.What = scti.What
	tp.PltFrom = scti.PltFrom
	tp.PltTo = scti.PltTo
	tp.WellTo = scti.WellTo
	tp.WellFrom = scti.WellFrom
	tp.Volume = wunit.CopyVolume(scti.Volume)
	tp.FPlateType = scti.FPlateType
	tp.TPlateType = scti.TPlateType
	tp.FVolume = wunit.CopyVolume(scti.FVolume)
	tp.TVolume = wunit.CopyVolume(scti.TVolume)
	tp.Channel = scti.Prms
	return tp
}

func NewSingleChannelTransferInstruction() *SingleChannelTransferInstruction {
	var v SingleChannelTransferInstruction
	v.Type = SCT
	return &v
}
func (ins *SingleChannelTransferInstruction) InstructionType() int {
	return ins.Type
}

func (ins *SingleChannelTransferInstruction) GetParameter(name string) interface{} {
	switch name {
	case "LIQUIDCLASS":
		return ins.What
	case "VOLUME":
		return ins.Volume
	case "VOLUNT":
		return nil
	case "FROMPLATETYPE":
		return ins.FPlateType
	case "WELLFROMVOLUME":
		return ins.FVolume
	case "POSFROM":
		return ins.PltFrom
	case "POSTO":
		return ins.PltTo
	case "WELLFROM":
		return ins.WellFrom
	case "PARAMS":
		return ins.Prms
	case "PLATFORM":
		return ins.Prms.Name
	case "WELLTO":
		return ins.WellTo
	case "WELLTOVOLUME":
		return ins.TVolume
	case "TOPLATETYPE":
		return ins.TPlateType
	case "INSTRUCTIONTYPE":
		return ins.InstructionType()
	}
	return nil
}

func (ins *SingleChannelTransferInstruction) Generate(policy *LHPolicyRuleSet, prms *LHProperties) []RobotInstruction {
	ret := make([]RobotInstruction, 0)
	// make the instructions

	suckinstruction := NewSuckInstruction()
	suckinstruction.AddTransferParams(ins.Params())
	suckinstruction.Multi = 1
	suckinstruction.Prms = ins.Prms

	blowinstruction := NewBlowInstruction()
	blowinstruction.AddTransferParams(ins.Params())
	blowinstruction.Multi = 1
	blowinstruction.Prms = ins.Prms

	ret = append(ret, suckinstruction)
	ret = append(ret, blowinstruction)

	// need to append to reset command

	resetinstruction := NewResetInstruction()
	resetinstruction.AddTransferParams(ins.Params())
	resetinstruction.Prms = ins.Prms
	ret = append(ret, resetinstruction)

	return ret
}

type MultiChannelTransferInstruction struct {
	Type       int
	What       []string
	PltFrom    []string
	PltTo      []string
	WellFrom   []string
	WellTo     []string
	Volume     []*wunit.Volume
	FPlateType []string
	TPlateType []string
	FVolume    []*wunit.Volume
	TVolume    []*wunit.Volume
	Multi      int
	Prms       *wtype.LHChannelParameter
}

func (scti *MultiChannelTransferInstruction) Params(k int) TransferParams {
	var tp TransferParams
	tp.What = scti.What[k]
	tp.PltFrom = scti.PltFrom[k]
	tp.PltTo = scti.PltTo[k]
	tp.WellFrom = scti.WellFrom[k]
	tp.WellTo = scti.WellTo[k]
	tp.Volume = wunit.CopyVolume(scti.Volume[k])
	tp.FPlateType = scti.FPlateType[k]
	tp.TPlateType = scti.TPlateType[k]
	tp.FVolume = wunit.CopyVolume(scti.FVolume[k])
	tp.TVolume = wunit.CopyVolume(scti.TVolume[k])
	tp.Channel = scti.Prms
	return tp
}
func NewMultiChannelTransferInstruction() *MultiChannelTransferInstruction {
	var v MultiChannelTransferInstruction
	v.Type = MCT
	v.What = make([]string, 0)
	v.PltFrom = make([]string, 0)
	v.PltTo = make([]string, 0)
	v.WellFrom = make([]string, 0)
	v.WellTo = make([]string, 0)
	v.Volume = make([]*wunit.Volume, 0)
	v.FVolume = make([]*wunit.Volume, 0)
	v.TVolume = make([]*wunit.Volume, 0)
	v.FPlateType = make([]string, 0)
	v.TPlateType = make([]string, 0)
	return &v
}
func (ins *MultiChannelTransferInstruction) InstructionType() int {
	return ins.Type
}

func (ins *MultiChannelTransferInstruction) GetParameter(name string) interface{} {
	switch name {
	case "LIQUIDCLASS":
		return ins.What
	case "VOLUME":
		return ins.Volume
	case "VOLUNT":
		return nil
	case "FROMPLATETYPE":
		return ins.FPlateType
	case "WELLFROMVOLUME":
		return ins.FVolume
	case "POSFROM":
		return ins.PltFrom
	case "POSTO":
		return ins.PltTo
	case "WELLFROM":
		return ins.WellFrom
	case "PARAMS":
		return ins.Prms
	case "PLATFORM":
		return ins.Prms.Name
	case "WELLTO":
		return ins.WellTo
	case "WELLTOVOLUME":
		return ins.TVolume
	case "TOPLATETYPE":
		return ins.TPlateType
	case "INSTRUCTIONTYPE":
		return ins.InstructionType()
	}
	return nil
}

func (ins *MultiChannelTransferInstruction) Generate(policy *LHPolicyRuleSet, prms *LHProperties) []RobotInstruction {
	ret := make([]RobotInstruction, 0)

	// make the instructions

	suckinstruction := NewSuckInstruction()
	blowinstruction := NewBlowInstruction()
	suckinstruction.Multi = ins.Multi
	blowinstruction.Multi = ins.Multi
	suckinstruction.Prms = ins.Prms
	blowinstruction.Prms = ins.Prms
	resetinstruction := NewResetInstruction()

	for i := 0; i < len(ins.Volume); i++ {
		suckinstruction.AddTransferParams(ins.Params(i))
		blowinstruction.AddTransferParams(ins.Params(i))
		resetinstruction.AddTransferParams(ins.Params(i))
	}

	ret = append(ret, suckinstruction)
	ret = append(ret, blowinstruction)

	ret = append(ret, resetinstruction)

	return ret
}

type StateChangeInstruction struct {
	Type     int
	OldState *wtype.LHChannelParameter
	NewState *wtype.LHChannelParameter
}

func NewStateChangeInstruction(oldstate, newstate *wtype.LHChannelParameter) *StateChangeInstruction {
	var v StateChangeInstruction
	v.Type = CCC
	v.OldState = oldstate
	v.NewState = newstate
	return &v
}
func (ins *StateChangeInstruction) InstructionType() int {
	return ins.Type
}

func (ins *StateChangeInstruction) GetParameter(name string) interface{} {
	switch name {
	case "OLDSTATE":
		return ins.OldState
	case "NEWSTATE":
		return ins.NewState
	case "INSTRUCTIONTYPE":
		return ins.InstructionType()
	}
	return nil
}

func (ins *StateChangeInstruction) Generate(policy *LHPolicyRuleSet, prms *LHProperties) []RobotInstruction {
	return nil
}

type ChangeAdaptorInstruction struct {
	Type           int
	Head           int
	DropPosition   string
	GetPosition    string
	OldAdaptorType string
	NewAdaptorType string
}

func NewChangeAdaptorInstruction(head int, droppos, getpos, oldad, newad string) *ChangeAdaptorInstruction {
	var v ChangeAdaptorInstruction
	v.Type = CHA
	v.Head = head
	v.DropPosition = droppos
	v.GetPosition = getpos
	v.OldAdaptorType = oldad
	v.NewAdaptorType = newad
	return &v
}
func (ins *ChangeAdaptorInstruction) InstructionType() int {
	return ins.Type
}

func (ins *ChangeAdaptorInstruction) GetParameter(name string) interface{} {
	switch name {
	case "HEAD":
		return ins.Head
	case "POSFROM":
		return ins.DropPosition
	case "POSTO":
		return ins.GetPosition
	case "OLDADAPTOR":
		return ins.OldAdaptorType
	case "NEWADAPTOR":
		return ins.NewAdaptorType
	case "INSTRUCTIONTYPE":
		return ins.InstructionType()
	}
	return nil
}

func (ins *ChangeAdaptorInstruction) Generate(policy *LHPolicyRuleSet, prms *LHProperties) []RobotInstruction {
	ret := make([]RobotInstruction, 4)
	/*
		ret[0]=NewMoveInstruction(ins.DropPosition,...)
		ret[1]=NewUnloadAdaptorInstruction(ins.DropPosition,...)
		ret[2]=NewMoveInstruction(ins.GetPosition, ...)
		ret[3]=NewLoadAdaptorInstruction(ins.GetPosition,...)
	*/

	return ret
}

type LoadTipsMoveInstruction struct {
	Type       int
	Head       int
	Well       []string
	FPosition  []string
	FPlateType []string
	Multi      int
}

func NewLoadTipsMoveInstruction() *LoadTipsMoveInstruction {
	var v LoadTipsMoveInstruction
	v.Type = LDT
	v.Well = make([]string, 0)
	v.FPosition = make([]string, 0)
	v.FPlateType = make([]string, 0)
	return &v
}
func (ins *LoadTipsMoveInstruction) InstructionType() int {
	return ins.Type
}

func (ins *LoadTipsMoveInstruction) GetParameter(name string) interface{} {
	switch name {
	case "HEAD":
		return ins.Head
	case "FROMPLATETYPE":
		return ins.FPlateType
	case "POSFROM":
		return ins.FPosition
	case "WELLFROM":
		return ins.Well
	case "Multi":
		return ins.Multi
	case "INSTRUCTIONTYPE":
		return ins.InstructionType()
	}
	return nil
}

func (ins *LoadTipsMoveInstruction) Generate(policy *LHPolicyRuleSet, prms *LHProperties) []RobotInstruction {
	ret := make([]RobotInstruction, 2)

	// move

	mov := NewMoveInstruction()
	mov.Head = ins.Head
	mov.Pos = ins.FPosition
	mov.Well = ins.Well
	mov.Plt = ins.FPlateType
	ret[0] = mov

	// load tips

	lod := NewLoadTipsInstruction()
	lod.Head = ins.Head
	lod.TipType = ins.FPlateType
	lod.HolderType = ins.FPlateType
	lod.Multi = ins.Multi
	lod.Pos = ins.FPosition
	lod.HolderType = ins.FPlateType
	lod.Well = ins.Well
	ret[1] = lod

	return ret
}

type UnloadTipsMoveInstruction struct {
	Type       int
	Head       int
	PltTo      []string
	WellTo     []string
	TPlateType []string
	Multi      int
}

func NewUnloadTipsMoveInstruction() *UnloadTipsMoveInstruction {
	var v UnloadTipsMoveInstruction
	v.Type = UDT
	v.PltTo = make([]string, 0)
	v.WellTo = make([]string, 0)
	v.TPlateType = make([]string, 0)
	return &v
}
func (ins *UnloadTipsMoveInstruction) InstructionType() int {
	return ins.Type
}

func (ins *UnloadTipsMoveInstruction) GetParameter(name string) interface{} {
	switch name {
	case "HEAD":
		return ins.Head
	case "TOPLATETYPE":
		return ins.TPlateType
	case "POSTO":
		return ins.PltTo
	case "WELLTO":
		return ins.WellTo
	case "INSTRUCTIONTYPE":
		return ins.InstructionType()
	case "MULTI":
		return ins.Multi
	}
	return nil
}

func (ins *UnloadTipsMoveInstruction) Generate(policy *LHPolicyRuleSet, prms *LHProperties) []RobotInstruction {
	ret := make([]RobotInstruction, 2)

	// move

	mov := NewMoveInstruction()
	mov.Head = ins.Head
	mov.Pos = ins.PltTo
	mov.Well = ins.WellTo
	mov.Plt = ins.TPlateType
	ret[0] = mov

	// unload tips

	uld := NewUnloadTipsInstruction()
	uld.Head = ins.Head
	uld.TipType = ins.TPlateType
	uld.HolderType = ins.TPlateType
	uld.Multi = ins.Multi
	uld.Pos = ins.PltTo
	uld.HolderType = ins.TPlateType
	uld.Well = ins.WellTo
	ret[1] = uld

	return ret
}

type AspirateInstruction struct {
	Type       int
	Head       int
	Volume     []*wunit.Volume
	Overstroke bool
	Multi      int
	Plt        []string
	What       []string
	LLF        []bool
}

func NewAspirateInstruction() *AspirateInstruction {
	var v AspirateInstruction
	v.Type = ASP
	v.Volume = make([]*wunit.Volume, 0)
	v.Plt = make([]string, 0)
	v.What = make([]string, 0)
	v.LLF = make([]bool, 0)
	return &v
}
func (ins *AspirateInstruction) InstructionType() int {
	return ins.Type
}

func (ins *AspirateInstruction) GetParameter(name string) interface{} {
	switch name {
	case "VOLUME":
		return ins.Volume
	case "HEAD":
		return ins.Head
	case "MULTI":
		return ins.Multi
	case "OVERSTROKE":
		return ins.Overstroke
	case "INSTRUCTIONTYPE":
		return ins.InstructionType()
	case "WHAT":
		return ins.What
	case "PLATE":
		return ins.Plt
	case "LLF":
		return ins.LLF
	}
	return nil
}

func (ins *AspirateInstruction) Generate(policy *LHPolicyRuleSet, prms *LHProperties) []RobotInstruction {
	return nil
}

func (ins *AspirateInstruction) OutputTo(driver LiquidhandlingDriver) {
	volumes := make([]float64, len(ins.Volume))
	for i, vol := range ins.Volume {
		volumes[i] = vol.ConvertTo(wunit.ParsePrefixedUnit("ul"))
	}
	os := []bool{ins.Overstroke}
	driver.Aspirate(volumes, os, ins.Head, ins.Multi, ins.Plt, ins.What, ins.LLF)
}

type DispenseInstruction struct {
	Type   int
	Head   int
	Volume []*wunit.Volume
	Multi  int
	Plt    []string
	What   []string
	LLF    []bool
}

func NewDispenseInstruction() *DispenseInstruction {
	var v DispenseInstruction
	v.Type = DSP
	v.Volume = make([]*wunit.Volume, 0)
	v.Plt = make([]string, 0)
	v.What = make([]string, 0)
	v.LLF = make([]bool, 0)
	return &v
}
func (ins *DispenseInstruction) InstructionType() int {
	return ins.Type
}

func (ins *DispenseInstruction) GetParameter(name string) interface{} {
	switch name {
	case "VOLUME":
		return ins.Volume
	case "HEAD":
		return ins.Head
	case "MULTI":
		return ins.Multi
	case "INSTRUCTIONTYPE":
		return ins.InstructionType()
	case "WHAT":
		return ins.What
	case "LLF":
		return ins.LLF
	case "PLT":
		return ins.Plt
	}
	return nil
}

func (ins *DispenseInstruction) Generate(policy *LHPolicyRuleSet, prms *LHProperties) []RobotInstruction {
	return nil
}

func (ins *DispenseInstruction) OutputTo(driver LiquidhandlingDriver) {
	volumes := make([]float64, len(ins.Volume))
	for i, vol := range ins.Volume {
		volumes[i] = vol.ConvertTo(wunit.ParsePrefixedUnit("ul"))
	}

	os := []bool{false}
	driver.Dispense(volumes, os, ins.Head, ins.Multi, ins.Plt, ins.What, ins.LLF)
}

type BlowoutInstruction struct {
	Type   int
	Head   int
	Volume []*wunit.Volume
	Multi  int
	Plt    []string
	What   []string
	LLF    []bool
}

func NewBlowoutInstruction() *BlowoutInstruction {
	var v BlowoutInstruction
	v.Type = BLO
	v.Volume = make([]*wunit.Volume, 0)
	return &v
}
func (ins *BlowoutInstruction) InstructionType() int {
	return ins.Type
}

func (ins *BlowoutInstruction) GetParameter(name string) interface{} {
	switch name {
	case "VOLUME":
		return ins.Volume
	case "HEAD":
		return ins.Head
	case "MULTI":
		return ins.Multi
	case "INSTRUCTIONTYPE":
		return ins.InstructionType()
	case "WHAT":
		return ins.What
	case "LLF":
		return ins.LLF
	case "PLT":
		return ins.Plt
	}
	return nil
}

func (ins *BlowoutInstruction) Generate(policy *LHPolicyRuleSet, prms *LHProperties) []RobotInstruction {
	return nil
}

func (ins *BlowoutInstruction) OutputTo(driver LiquidhandlingDriver) {
	volumes := make([]float64, len(ins.Volume))
	for i, vol := range ins.Volume {
		volumes[i] = vol.ConvertTo(wunit.ParsePrefixedUnit("ul"))
	}
	bo := make([]bool, ins.Multi)
	for i := 0; i < ins.Multi; i++ {
		bo[i] = true
	}
	driver.Dispense(volumes, bo, ins.Head, ins.Multi, ins.Plt, ins.What, ins.LLF)
}

type PTZInstruction struct {
	Type    int
	Head    int
	Channel int
}

func NewPTZInstruction() *PTZInstruction {
	var v PTZInstruction
	v.Type = PTZ
	return &v
}
func (ins *PTZInstruction) InstructionType() int {
	return ins.Type
}

func (ins *PTZInstruction) GetParameter(name string) interface{} {
	switch name {
	case "HEAD":
		return ins.Head
	case "CHANNEL":
		return ins.Channel
	case "INSTRUCTIONTYPE":
		return ins.InstructionType()
	}
	return nil
}

func (ins *PTZInstruction) Generate(policy *LHPolicyRuleSet, prms *LHProperties) []RobotInstruction {
	return nil
}

func (ins *PTZInstruction) OutputTo(driver LiquidhandlingDriver) {
	driver.ResetPistons(ins.Head, ins.Channel)
}

type MoveInstruction struct {
	Type      int
	Head      int
	Pos       []string
	Plt       []string
	Well      []string
	WVolume   []*wunit.Volume
	Reference []int
	OffsetX   []float64
	OffsetY   []float64
	OffsetZ   []float64
}

func NewMoveInstruction() *MoveInstruction {
	var v MoveInstruction
	v.Type = MOV
	v.Plt = make([]string, 0)
	v.Pos = make([]string, 0)
	v.Well = make([]string, 0)
	v.WVolume = make([]*wunit.Volume, 0)
	v.Reference = make([]int, 0)
	v.OffsetX = make([]float64, 0)
	v.OffsetY = make([]float64, 0)
	v.OffsetZ = make([]float64, 0)
	return &v
}
func (ins *MoveInstruction) InstructionType() int {
	return ins.Type
}

func (ins *MoveInstruction) GetParameter(name string) interface{} {
	switch name {
	case "TOWELLVOLUME":
		return ins.WVolume
	case "HEAD":
		return ins.Head
	case "TOPLATETYPE":
		return ins.Plt
	case "POSTO":
		return ins.Pos
	case "WELLTO":
		return ins.Well
	case "REFERENCE":
		return ins.Reference
	case "OFFSETX":
		return ins.OffsetX
	case "OFFSETY":
		return ins.OffsetY
	case "OFFSETZ":
		return ins.OffsetZ
	case "INSTRUCTIONTYPE":
		return ins.InstructionType()
	}
	return nil
}

func (ins *MoveInstruction) Generate(policy *LHPolicyRuleSet, prms *LHProperties) []RobotInstruction {
	return nil
}

func (ins *MoveInstruction) OutputTo(driver LiquidhandlingDriver) {
	driver.Move(ins.Pos, ins.Well, ins.Reference, ins.OffsetX, ins.OffsetY, ins.OffsetZ, ins.Plt, ins.Head)
}

type MoveRawInstruction struct {
	Type       int
	Head       int
	What       []string
	PltFrom    []string
	PltTo      []string
	WellFrom   []string
	WellTo     []string
	Volume     []*wunit.Volume
	FPlateType []string
	TPlateType []string
	FVolume    []*wunit.Volume
	TVolume    []*wunit.Volume
	Prms       *wtype.LHChannelParameter
}

func NewMoveRawInstruction() *MoveRawInstruction {
	var v MoveRawInstruction
	v.Type = MRW
	v.What = make([]string, 0)
	v.PltFrom = make([]string, 0)
	v.PltTo = make([]string, 0)
	v.WellFrom = make([]string, 0)
	v.WellTo = make([]string, 0)
	v.FPlateType = make([]string, 0)
	v.TPlateType = make([]string, 0)
	v.Volume = make([]*wunit.Volume, 0)
	v.FVolume = make([]*wunit.Volume, 0)
	v.TVolume = make([]*wunit.Volume, 0)
	return &v
}
func (ins *MoveRawInstruction) InstructionType() int {
	return ins.Type
}

func (ins *MoveRawInstruction) GetParameter(name string) interface{} {
	switch name {
	case "HEAD":
		return ins.Head
	case "LIQUIDCLASS":
		return ins.What
	case "VOLUME":
		return ins.Volume
	case "FROMPLATETYPE":
		return ins.FPlateType
	case "TOPLATETYPE":
		return ins.TPlateType
	case "WELLFROMVOLUME":
		return ins.FVolume
	case "WELLTOVOLUME":
		return ins.TVolume
	case "POSFROM":
		return ins.PltFrom
	case "POSTO":
		return ins.PltTo
	case "WELLFROM":
		return ins.WellFrom
	case "PARAMS":
		return ins.Prms
	case "PLATFORM":
		return ins.Prms.Name
	case "INSTRUCTIONTYPE":
		return ins.InstructionType()
	}
	return nil
}

func (ins *MoveRawInstruction) Generate(policy *LHPolicyRuleSet, prms *LHProperties) []RobotInstruction {
	return nil
}

func (ins *MoveRawInstruction) OutputTo(driver LiquidhandlingDriver) {
	panic("Not yet implemented")
}

type LoadTipsInstruction struct {
	Type       int
	Head       int
	Pos        []string
	Well       []string
	Channels   []int
	TipType    []string
	HolderType []string
	Multi      int
}

func NewLoadTipsInstruction() *LoadTipsInstruction {
	var v LoadTipsInstruction
	v.Type = LOD
	v.Channels = make([]int, 0)
	v.TipType = make([]string, 0)
	v.HolderType = make([]string, 0)
	v.Pos = make([]string, 0)
	v.Well = make([]string, 0)
	return &v
}
func (ins *LoadTipsInstruction) InstructionType() int {
	return ins.Type
}

func (ins *LoadTipsInstruction) GetParameter(name string) interface{} {
	switch name {
	case "HEAD":
		return ins.Head
	case "CHANNEL":
		return ins.Channels
	case "TIPTYPE":
		return ins.TipType
	case "FROMPLATETYPE":
		return ins.HolderType
	case "INSTRUCTIONTYPE":
		return ins.InstructionType()
	case "MULTI":
		return ins.Multi
	case "WELL":
		return ins.Well
	case "PLATE":
		return ins.HolderType
	case "POS":
		return ins.Pos
	}
	return nil
}

func (ins *LoadTipsInstruction) Generate(policy *LHPolicyRuleSet, prms *LHProperties) []RobotInstruction {
	return nil
}

func (ins *LoadTipsInstruction) OutputTo(driver LiquidhandlingDriver) {
	driver.LoadTips(ins.Channels, ins.Head, len(ins.TipType), ins.HolderType, ins.Pos, ins.Well)
}

type UnloadTipsInstruction struct {
	Type       int
	Head       int
	Channels   []int
	TipType    []string
	HolderType []string
	Multi      int
	Pos        []string
	Well       []string
}

func NewUnloadTipsInstruction() *UnloadTipsInstruction {
	var v UnloadTipsInstruction
	v.Type = ULD
	v.TipType = make([]string, 0)
	v.HolderType = make([]string, 0)
	v.Channels = make([]int, 0)
	v.Pos = make([]string, 0)
	v.Well = make([]string, 0)
	return &v
}
func (ins *UnloadTipsInstruction) InstructionType() int {
	return ins.Type
}

func (ins *UnloadTipsInstruction) GetParameter(name string) interface{} {
	switch name {
	case "HEAD":
		return ins.Head
	case "CHANNEL":
		return ins.Channels
	case "TIPTYPE":
		return ins.TipType
	case "TOPLATETYPE":
		return ins.HolderType
	case "INSTRUCTIONTYPE":
		return ins.InstructionType()
	case "MULTI":
		return ins.Multi
	case "WELL":
		return ins.Well
	case "POS":
		return ins.Pos
	}
	return nil
}

func (ins *UnloadTipsInstruction) Generate(policy *LHPolicyRuleSet, prms *LHProperties) []RobotInstruction {
	return nil
}

func (ins *UnloadTipsInstruction) OutputTo(driver LiquidhandlingDriver) {
	driver.UnloadTips(ins.Channels, ins.Head, len(ins.TipType), ins.HolderType, ins.Pos, ins.Well)
}

type SuckInstruction struct {
	Type       int
	Head       int
	What       []string
	PltFrom    []string
	WellFrom   []string
	Volume     []*wunit.Volume
	FPlateType []string
	FVolume    []*wunit.Volume
	Prms       *wtype.LHChannelParameter
	Multi      int
	Overstroke bool
}

func NewSuckInstruction() *SuckInstruction {
	var v SuckInstruction
	v.Type = SUK
	v.What = make([]string, 0)
	v.PltFrom = make([]string, 0)
	v.WellFrom = make([]string, 0)
	v.Volume = make([]*wunit.Volume, 0)
	v.FPlateType = make([]string, 0)
	v.FVolume = make([]*wunit.Volume, 0)
	return &v
}
func (ins *SuckInstruction) InstructionType() int {
	return ins.Type
}

func (ins *SuckInstruction) AddTransferParams(tp TransferParams) {
	ins.What = append(ins.What, tp.What)
	ins.PltFrom = append(ins.PltFrom, tp.PltFrom)
	ins.WellFrom = append(ins.WellFrom, tp.WellFrom)
	ins.Volume = append(ins.Volume, tp.Volume)
	ins.FPlateType = append(ins.FPlateType, tp.FPlateType)
	ins.FVolume = append(ins.FVolume, tp.FVolume)
}

func (ins *SuckInstruction) GetParameter(name string) interface{} {
	switch name {
	case "HEAD":
		return ins.Head
	case "LIQUIDCLASS":
		return ins.What
	case "VOLUME":
		return ins.Volume
	case "FROMPLATETYPE":
		return ins.FPlateType
	case "WELLFROMVOLUME":
		return ins.FVolume
	case "POSFROM":
		return ins.PltFrom
	case "WELLFROM":
		return ins.WellFrom
	case "PARAMS":
		return ins.Prms
	case "MULTI":
		return ins.Multi
	case "OVERSTROKE":
		return ins.Overstroke
	case "PLATFORM":
		return ins.Prms.Name
	case "INSTRUCTIONTYPE":
		return ins.InstructionType()
	}
	return nil
}

func (ins *SuckInstruction) Generate(policy *LHPolicyRuleSet, prms *LHProperties) []RobotInstruction {
	ret := make([]RobotInstruction, 0, 1)

	// this is where the policies come into effect

	pol := policy.GetPolicyFor(ins)

	// so a simple list of questions

	// first we generate the move

	// do we need to enter slowly?

	entryspeed, gentlynow := pol["ASPENTRYSPEED"]

	if gentlynow {
		// go to the well top
		mov := NewMoveInstruction()

		mov.Head = ins.Head
		mov.Pos = ins.PltFrom
		mov.Plt = ins.FPlateType
		mov.Well = ins.WellFrom
		mov.WVolume = ins.FVolume
		for i := 0; i < ins.Multi; i++ {
			mov.Reference = append(mov.Reference, 1)
			mov.OffsetZ = append(mov.OffsetZ, 5.0)
		}
		ret = append(ret, mov)

		// set the speed
		spd := NewSetDriveSpeedInstruction()
		spd.Drive = "Z"
		spd.Speed = entryspeed.(float64)
		ret = append(ret, spd)

		// now move into the liquid
		mov = NewMoveInstruction()
		mov.Head = ins.Head
		mov.Pos = ins.PltFrom
		mov.Plt = ins.FPlateType
		mov.Well = ins.WellFrom
		mov.WVolume = ins.FVolume
		for i := 0; i < ins.Multi; i++ {
			mov.Reference = append(mov.Reference, 0)
			mov.OffsetZ = append(mov.OffsetZ, pol["ASPZOFFSET"].(float64))
		}

		ret = append(ret, mov)
		// reset the drive speed
		spd = NewSetDriveSpeedInstruction()
		spd.Drive = "Z"
		spd.Speed = pol["DEFAULTZSPEED"].(float64)
		ret = append(ret, spd)
	} else {
		mov := NewMoveInstruction()
		mov.Head = ins.Head

		mov.Pos = ins.PltFrom
		mov.Plt = ins.FPlateType
		mov.Well = ins.WellFrom
		mov.WVolume = ins.FVolume
		for i := 0; i < ins.Multi; i++ {
			mov.Reference = append(mov.Reference, 0)
			mov.OffsetZ = append(mov.OffsetZ, pol["ASPZOFFSET"].(float64))
		}
		ret = append(ret, mov)
	}

	// do we pre-mix?

	cycles, premix := pol["PRE_MIX"]

	if premix {
		// add the premix step
		mix := NewMoveMixInstruction()
		mix.Plt = ins.PltFrom
		mix.PlateType = ins.FPlateType
		mix.Well = ins.WellFrom
		mix.FVolume = ins.FVolume
		mix.Multi = ins.Multi

		// this is not safe
		mixvol, ok := pol["PRE_MIX_VOL"]
		mix.Volume = ins.Volume

		if ok {
			v := make([]*wunit.Volume, ins.Multi)
			for i := 0; i < ins.Multi; i++ {
				vl := wunit.NewVolume(mixvol.(float64), "ul")
				v[i] = &vl
			}
			mix.Volume = v
		}

		c := make([]int, ins.Multi)

		for i := 0; i < ins.Multi; i++ {
			c[i] = cycles.(int)
		}

		mix.Cycles = c
		ret = append(ret, mix)
	}

	// Set the pipette speed if needed

	pspeed, setpspeed := pol["ASPSPEED"]

	if setpspeed {
		sps := NewSetPipetteSpeedInstruction()
		sps.Head = ins.Head
		sps.Channel = -1 // all channels
		sps.Speed = pspeed.(float64)
		ret = append(ret, sps)
	}

	// now we aspirate

	aspins := NewAspirateInstruction()
	aspins.Head = ins.Head
	aspins.Volume = ins.Volume
	aspins.Multi = ins.Multi
	aspins.Overstroke = ins.Overstroke
	aspins.What = ins.What
	aspins.Plt = ins.FPlateType
	ret = append(ret, aspins)

	// do we reset the pipette speed?

	if setpspeed {
		sps := NewSetPipetteSpeedInstruction()
		sps.Head = ins.Head
		sps.Channel = -1 // all channels
		sps.Speed = pol["DEFAULTPIPETTESPEED"].(float64)
		ret = append(ret, sps)
	}

	// do we wait

	wait_time, wait := pol["ASP_WAIT"]

	if wait {
		waitins := NewWaitInstruction()
		waitins.Time = wait_time.(float64)
		ret = append(ret, waitins)
	}

	return ret
}

type BlowInstruction struct {
	Type       int
	Head       int
	What       []string
	PltTo      []string
	WellTo     []string
	Volume     []*wunit.Volume
	TPlateType []string
	TVolume    []*wunit.Volume
	Prms       *wtype.LHChannelParameter
	Multi      int
}

func NewBlowInstruction() *BlowInstruction {
	var v BlowInstruction
	v.Type = BLW
	v.What = make([]string, 0)
	v.PltTo = make([]string, 0)
	v.WellTo = make([]string, 0)
	v.Volume = make([]*wunit.Volume, 0)
	v.TPlateType = make([]string, 0)
	v.TVolume = make([]*wunit.Volume, 0)
	return &v
}
func (ins *BlowInstruction) InstructionType() int {
	return ins.Type
}

func (ins *BlowInstruction) GetParameter(name string) interface{} {
	switch name {
	case "HEAD":
		return ins.Head
	case "LIQUIDCLASS":
		return ins.What
	case "VOLUME":
		return ins.Volume
	case "TOPLATETYPE":
		return ins.TPlateType
	case "WELLTOVOLUME":
		return ins.TVolume
	case "POSTO":
		return ins.PltTo
	case "WELLTO":
		return ins.WellTo
	case "PARAMS":
		return ins.Prms
	case "PLATFORM":
		return ins.Prms.Name
	case "MULTI":
		return ins.Multi
	case "INSTRUCTIONTYPE":
		return ins.InstructionType()
	}
	return nil
}
func (ins *BlowInstruction) AddTransferParams(tp TransferParams) {
	ins.What = append(ins.What, tp.What)
	ins.PltTo = append(ins.PltTo, tp.PltTo)
	ins.WellTo = append(ins.WellTo, tp.WellTo)
	ins.Volume = append(ins.Volume, tp.Volume)
	ins.TPlateType = append(ins.TPlateType, tp.TPlateType)
	ins.TVolume = append(ins.TVolume, tp.TVolume)
}

func (ins *BlowInstruction) Generate(policy *LHPolicyRuleSet, prms *LHProperties) []RobotInstruction {
	ret := make([]RobotInstruction, 0)
	// apply policies here

	pol := policy.GetPolicyFor(ins)

	// first, are we breaking up the move?

	entryspeed, gentlydoesit := pol["DSPENTRYSPEED"]

	if gentlydoesit {
		// go to the well top
		mov := NewMoveInstruction()

		mov.Head = ins.Head
		mov.Pos = ins.PltTo
		mov.Plt = ins.TPlateType
		mov.Well = ins.WellTo
		mov.WVolume = ins.TVolume
		for i := 0; i < ins.Multi; i++ {
			mov.Reference = append(mov.Reference, 1)
			mov.OffsetZ = append(mov.OffsetZ, 5.0)
		}
		ret = append(ret, mov)

		// set the speed
		spd := NewSetDriveSpeedInstruction()
		spd.Drive = "Z"
		spd.Speed = entryspeed.(float64)
		ret = append(ret, spd)

		mov = NewMoveInstruction()
		mov.Head = ins.Head
		mov.Pos = ins.PltTo
		mov.Plt = ins.TPlateType
		mov.Well = ins.WellTo
		mov.WVolume = ins.TVolume
		for i := 0; i < ins.Multi; i++ {
			mov.Reference = append(mov.Reference, pol["DSPREFERENCE"].(int))
			mov.OffsetZ = append(mov.OffsetZ, pol["DSPZOFFSET"].(float64))
		}
		ret = append(ret, mov)
		// reset the drive speed
		spd = NewSetDriveSpeedInstruction()
		spd.Drive = "Z"
		spd.Speed = pol["DEFAULTZSPEED"].(float64)
		ret = append(ret, spd)

	} else {
		mov := NewMoveInstruction()
		mov.Head = ins.Head
		mov.Pos = ins.PltTo
		mov.Plt = ins.TPlateType
		mov.Well = ins.WellTo
		mov.WVolume = ins.TVolume
		for i := 0; i < ins.Multi; i++ {
			mov.Reference = append(mov.Reference, pol["DSPREFERENCE"].(int))
			mov.OffsetZ = append(mov.OffsetZ, pol["DSPZOFFSET"].(float64))
		}

		ret = append(ret, mov)
	}

	// next, are we setting the pipette speed

	pspeed, setpspeed := pol["DSPSPEED"]

	if setpspeed {
		sps := NewSetPipetteSpeedInstruction()
		sps.Head = ins.Head
		sps.Channel = -1 // all channels
		sps.Speed = pspeed.(float64)
		ret = append(ret, sps)
	}

	// now we dispense

	dspins := NewDispenseInstruction()
	dspins.Head = ins.Head
	dspins.Volume = ins.Volume
	dspins.Multi = ins.Multi
	dspins.Plt = ins.TPlateType
	dspins.What = ins.What
	ret = append(ret, dspins)

	// do we reset the pipette speed?

	if setpspeed {
		sps := NewSetPipetteSpeedInstruction()
		sps.Head = ins.Head
		sps.Channel = -1 // all channels
		sps.Speed = pol["DEFAULTPIPETTESPEED"].(float64)
		ret = append(ret, sps)
	}

	// do we wait?

	wait_time, wait := pol["DSP_WAIT"]

	if wait {
		waitins := NewWaitInstruction()
		waitins.Time = wait_time.(float64)
		ret = append(ret, waitins)
	}

	// do we mix?
	cycles, postmix := pol["POST_MIX"]

	if postmix {
		// add the postmix step
		mix := NewMoveMixInstruction()
		mix.Plt = ins.PltTo
		mix.PlateType = ins.TPlateType
		mix.Well = ins.WellTo
		mix.FVolume = ins.TVolume
		mix.Multi = ins.Multi

		// this is not safe, need to verify volume is OK
		mixvol, ok := pol["POST_MIX_VOL"]
		mix.Volume = ins.Volume

		if ok {
			v := make([]*wunit.Volume, ins.Multi)
			for i := 0; i < ins.Multi; i++ {
				vl := wunit.NewVolume(mixvol.(float64), "ul")
				v[i] = &vl
			}
			mix.Volume = v
		}

		c := make([]int, ins.Multi)

		for i := 0; i < ins.Multi; i++ {
			c[i] = cycles.(int)
		}

		mix.Cycles = c
		ret = append(ret, mix)
	}

	// do we need to touch off?

	touch_off := pol["TOUCHOFF"].(bool)

	if touch_off {
		touch_offset := pol["TOUCHOFFSET"].(float64)
		mov := NewMoveInstruction()
		mov.Head = ins.Head
		mov.Pos = ins.PltTo
		mov.Plt = ins.TPlateType
		mov.Well = ins.WellTo
		mov.WVolume = ins.TVolume

		ref := make([]int, ins.Multi)
		off := make([]float64, ins.Multi)
		for i := 0; i < ins.Multi; i++ {
			ref[i] = 0
			off[i] = touch_offset
		}

		mov.Reference = ref
		mov.OffsetZ = off

		ret = append(ret, mov)
	}

	return ret
}

type SetPipetteSpeedInstruction struct {
	Type    int
	Head    int
	Channel int
	Speed   float64
}

func NewSetPipetteSpeedInstruction() *SetPipetteSpeedInstruction {
	var v SetPipetteSpeedInstruction
	v.Type = SPS
	return &v
}
func (ins *SetPipetteSpeedInstruction) InstructionType() int {
	return ins.Type
}

func (ins *SetPipetteSpeedInstruction) GetParameter(name string) interface{} {
	switch name {
	case "HEAD":
		return ins.Head
	case "CHANNEL":
		return ins.Channel
	case "SPEED":
		return ins.Speed
	case "INSTRUCTIONTYPE":
		return ins.InstructionType()
	}
	return nil
}

func (ins *SetPipetteSpeedInstruction) Generate(policy *LHPolicyRuleSet, prms *LHProperties) []RobotInstruction {
	return nil
}

func (ins *SetPipetteSpeedInstruction) OutputTo(driver LiquidhandlingDriver) {
	driver.SetPipetteSpeed(ins.Head, ins.Channel, ins.Speed)
}

type SetDriveSpeedInstruction struct {
	Type  int
	Drive string
	Speed float64
}

func NewSetDriveSpeedInstruction() *SetDriveSpeedInstruction {
	var v SetDriveSpeedInstruction
	v.Type = SDS
	return &v
}
func (ins *SetDriveSpeedInstruction) InstructionType() int {
	return ins.Type
}

func (ins *SetDriveSpeedInstruction) GetParameter(name string) interface{} {
	switch name {
	case "DRIVE":
		return ins.Drive
	case "SPEED":
		return ins.Speed
	case "INSTRUCTIONTYPE":
		return ins.InstructionType()
	}
	return nil
}

func (ins *SetDriveSpeedInstruction) Generate(policy *LHPolicyRuleSet, prms *LHProperties) []RobotInstruction {
	return nil
}

func (ins *SetDriveSpeedInstruction) OutputTo(driver LiquidhandlingDriver) {
	driver.SetDriveSpeed(ins.Drive, ins.Speed)
}

type InitializeInstruction struct {
	Type int
}

func NewInitializeInstruction() *InitializeInstruction {
	var v InitializeInstruction
	v.Type = INI
	return &v
}
func (ins *InitializeInstruction) InstructionType() int {
	return ins.Type
}

func (ins *InitializeInstruction) GetParameter(name string) interface{} {
	return nil
}

func (ins *InitializeInstruction) Generate(policy *LHPolicyRuleSet, prms *LHProperties) []RobotInstruction {
	return nil
}

func (ins *InitializeInstruction) OutputTo(driver LiquidhandlingDriver) {
	driver.Initialize()
}

type FinalizeInstruction struct {
	Type int
}

func NewFinalizeInstruction() *FinalizeInstruction {
	var v FinalizeInstruction
	v.Type = FIN
	return &v
}
func (ins *FinalizeInstruction) InstructionType() int {
	return ins.Type
}

func (ins *FinalizeInstruction) GetParameter(name string) interface{} {
	return nil
}

func (ins *FinalizeInstruction) Generate(policy *LHPolicyRuleSet, prms *LHProperties) []RobotInstruction {
	return nil
}

func (ins *FinalizeInstruction) OutputTo(driver LiquidhandlingDriver) {
	driver.Finalize()
}

type WaitInstruction struct {
	Type int
	Time float64
}

func NewWaitInstruction() *WaitInstruction {
	var v WaitInstruction
	v.Type = WAI
	return &v
}
func (ins *WaitInstruction) InstructionType() int {
	return ins.Type
}

func (ins *WaitInstruction) GetParameter(name string) interface{} {
	switch name {
	case "TIME":
		return ins.Time
	case "INSTRUCTIONTYPE":
		return ins.InstructionType()
	}
	return nil
}

func (ins *WaitInstruction) Generate(policy *LHPolicyRuleSet, prms *LHProperties) []RobotInstruction {
	return nil
}

func (ins *WaitInstruction) OutputTo(driver ExtendedLiquidhandlingDriver) {
	driver.Wait(ins.Time)
}

type LightsOnInstruction struct {
	Type       int
	What       []string
	PltFrom    []string
	PltTo      []string
	WellFrom   []string
	WellTo     []string
	Volume     []*wunit.Volume
	FPlateType []string
	TPlateType []string
	FVolume    *wunit.Volume
	TVolume    *wunit.Volume
	Prms       *wtype.LHChannelParameter
}

func NewLightsOnInstruction() *LightsOnInstruction {
	var v LightsOnInstruction
	v.Type = LON
	v.What = make([]string, 0)
	v.PltFrom = make([]string, 0)
	v.PltTo = make([]string, 0)
	v.WellFrom = make([]string, 0)
	v.WellTo = make([]string, 0)
	v.Volume = make([]*wunit.Volume, 0)
	return &v
}
func (ins *LightsOnInstruction) InstructionType() int {
	return ins.Type
}

func (ins *LightsOnInstruction) GetParameter(name string) interface{} {
	return nil
}

func (ins *LightsOnInstruction) Generate(policy *LHPolicyRuleSet, prms *LHProperties) []RobotInstruction {
	return nil
}

func (ins *LightsOnInstruction) OutputTo(driver LiquidhandlingDriver) {
	panic("Not yet implemented")
}

type LightsOffInstruction struct {
	Type       int
	What       []string
	PltFrom    []string
	PltTo      []string
	WellFrom   []string
	WellTo     []string
	Volume     []*wunit.Volume
	FPlateType []string
	TPlateType []string
	FVolume    *wunit.Volume
	TVolume    *wunit.Volume
	Prms       *wtype.LHChannelParameter
}

func NewLightsOffInstruction() *LightsOffInstruction {
	var v LightsOffInstruction
	v.Type = LOF
	v.What = make([]string, 0)
	v.PltFrom = make([]string, 0)
	v.PltTo = make([]string, 0)
	v.WellFrom = make([]string, 0)
	v.WellTo = make([]string, 0)
	v.Volume = make([]*wunit.Volume, 0)
	return &v
}
func (ins *LightsOffInstruction) InstructionType() int {
	return ins.Type
}

func (ins *LightsOffInstruction) GetParameter(name string) interface{} {
	return nil
}

func (ins *LightsOffInstruction) Generate(policy *LHPolicyRuleSet, prms *LHProperties) []RobotInstruction {
	return nil
}

func (ins *LightsOffInstruction) OutputTo(driver LiquidhandlingDriver) {
	panic("Not yet implemented")
}

type OpenInstruction struct {
	Type       int
	What       []string
	PltFrom    []string
	PltTo      []string
	WellFrom   []string
	WellTo     []string
	Volume     []*wunit.Volume
	FPlateType []string
	TPlateType []string
	FVolume    *wunit.Volume
	TVolume    *wunit.Volume
	Prms       *wtype.LHChannelParameter
}

func NewOpenInstruction() *OpenInstruction {
	var v OpenInstruction
	v.Type = OPN
	v.What = make([]string, 0)
	v.PltFrom = make([]string, 0)
	v.PltTo = make([]string, 0)
	v.WellFrom = make([]string, 0)
	v.WellTo = make([]string, 0)
	v.Volume = make([]*wunit.Volume, 0)
	return &v
}
func (ins *OpenInstruction) InstructionType() int {
	return ins.Type
}

func (ins *OpenInstruction) GetParameter(name string) interface{} {
	return nil
}

func (ins *OpenInstruction) Generate(policy *LHPolicyRuleSet, prms *LHProperties) []RobotInstruction {
	return nil
}

func (ins *OpenInstruction) OutputTo(driver LiquidhandlingDriver) {
	panic("Not yet implemented")
}

type CloseInstruction struct {
	Type       int
	What       []string
	PltFrom    []string
	PltTo      []string
	WellFrom   []string
	WellTo     []string
	Volume     []*wunit.Volume
	FPlateType []string
	TPlateType []string
	FVolume    *wunit.Volume
	TVolume    *wunit.Volume
	Prms       *wtype.LHChannelParameter
}

func NewCloseInstruction() *CloseInstruction {
	var v CloseInstruction
	v.Type = CLS
	v.What = make([]string, 0)
	v.PltFrom = make([]string, 0)
	v.PltTo = make([]string, 0)
	v.WellFrom = make([]string, 0)
	v.WellTo = make([]string, 0)
	v.Volume = make([]*wunit.Volume, 0)
	return &v
}
func (ins *CloseInstruction) InstructionType() int {
	return ins.Type
}

func (ins *CloseInstruction) GetParameter(name string) interface{} {
	return nil
}

func (ins *CloseInstruction) Generate(policy *LHPolicyRuleSet, prms *LHProperties) []RobotInstruction {
	return nil
}

func (ins *CloseInstruction) OutputTo(driver LiquidhandlingDriver) {
	panic("Not yet implemented")
}

type LoadAdaptorInstruction struct {
	Type       int
	What       []string
	PltFrom    []string
	PltTo      []string
	WellFrom   []string
	WellTo     []string
	Volume     []*wunit.Volume
	FPlateType []string
	TPlateType []string
	FVolume    *wunit.Volume
	TVolume    *wunit.Volume
	Prms       *wtype.LHChannelParameter
}

func NewLoadAdaptorInstruction() *LoadAdaptorInstruction {
	var v LoadAdaptorInstruction
	v.Type = LAD
	v.What = make([]string, 0)
	v.PltFrom = make([]string, 0)
	v.PltTo = make([]string, 0)
	v.WellFrom = make([]string, 0)
	v.WellTo = make([]string, 0)
	v.Volume = make([]*wunit.Volume, 0)
	return &v
}
func (ins *LoadAdaptorInstruction) InstructionType() int {
	return ins.Type
}

func (ins *LoadAdaptorInstruction) GetParameter(name string) interface{} {
	return nil
}

func (ins *LoadAdaptorInstruction) Generate(policy *LHPolicyRuleSet, prms *LHProperties) []RobotInstruction {
	return nil
}

func (ins *LoadAdaptorInstruction) OutputTo(driver LiquidhandlingDriver) {
	panic("Not yet implemented")
}

type UnloadAdaptorInstruction struct {
	Type       int
	What       []string
	PltFrom    []string
	PltTo      []string
	WellFrom   []string
	WellTo     []string
	Volume     []*wunit.Volume
	FPlateType []string
	TPlateType []string
	FVolume    *wunit.Volume
	TVolume    *wunit.Volume
	Prms       *wtype.LHChannelParameter
}

func NewUnloadAdaptorInstruction() *UnloadAdaptorInstruction {
	var v UnloadAdaptorInstruction
	v.Type = UAD
	v.What = make([]string, 0)
	v.PltFrom = make([]string, 0)
	v.PltTo = make([]string, 0)
	v.WellFrom = make([]string, 0)
	v.WellTo = make([]string, 0)
	v.Volume = make([]*wunit.Volume, 0)
	return &v
}
func (ins *UnloadAdaptorInstruction) InstructionType() int {
	return ins.Type
}

func (ins *UnloadAdaptorInstruction) GetParameter(name string) interface{} {
	return nil
}

func (ins *UnloadAdaptorInstruction) Generate(policy *LHPolicyRuleSet, prms *LHProperties) []RobotInstruction {
	return nil
}

func (ins *UnloadAdaptorInstruction) OutputTo(driver LiquidhandlingDriver) {
	panic("Not yet implemented")
}

type ResetInstruction struct {
	Type       int
	What       []string
	PltFrom    []string
	PltTo      []string
	WellFrom   []string
	WellTo     []string
	Volume     []*wunit.Volume
	FPlateType []string
	TPlateType []string
	FVolume    []*wunit.Volume
	TVolume    []*wunit.Volume
	Prms       *wtype.LHChannelParameter
}

func NewResetInstruction() *ResetInstruction {
	var ri ResetInstruction
	ri.Type = RST
	ri.What = make([]string, 0)
	ri.PltFrom = make([]string, 0)
	ri.WellFrom = make([]string, 0)
	ri.WellTo = make([]string, 0)
	ri.Volume = make([]*wunit.Volume, 0)
	ri.FPlateType = make([]string, 0)
	ri.TPlateType = make([]string, 0)
	ri.FVolume = make([]*wunit.Volume, 0)
	ri.TVolume = make([]*wunit.Volume, 0)
	return &ri
}

func (ins *ResetInstruction) InstructionType() int {
	return ins.Type
}

func (ins *ResetInstruction) GetParameter(name string) interface{} {
	switch name {
	case "LIQUIDCLASS":
		return ins.What
	case "VOLUME":
		return ins.Volume
	case "VOLUNT":
		return nil
	case "FROMPLATETYPE":
		return ins.FPlateType
	case "WELLFROMVOLUME":
		return ins.FVolume
	case "POSFROM":
		return ins.PltFrom
	case "POSTO":
		return ins.PltTo
	case "WELLFROM":
		return ins.WellFrom
	case "WELLTO":
		return ins.WellTo
	case "PARAMS":
		return ins.Prms
	case "PLATFORM":
		return ins.Prms.Name
	case "INSTRUCTIONTYPE":
		return ins.InstructionType()
	}
	return nil
}

func (ins *ResetInstruction) AddTransferParams(tp TransferParams) {
	ins.What = append(ins.What, tp.What)
	ins.PltTo = append(ins.PltTo, tp.PltTo)
	ins.WellTo = append(ins.WellTo, tp.WellTo)
	ins.PltFrom = append(ins.PltFrom, tp.PltFrom)
	ins.WellFrom = append(ins.WellFrom, tp.WellFrom)
	ins.Volume = append(ins.Volume, tp.Volume)
	ins.FPlateType = append(ins.FPlateType, tp.FPlateType)
	ins.TPlateType = append(ins.TPlateType, tp.TPlateType)
	ins.FVolume = append(ins.FVolume, tp.FVolume)
	ins.TVolume = append(ins.TVolume, tp.TVolume)
}

func (ins *ResetInstruction) Generate(policy *LHPolicyRuleSet, prms *LHProperties) []RobotInstruction {
	pol := policy.GetPolicyFor(ins)
	ret := make([]RobotInstruction, 0)

	mov := NewMoveInstruction()
	mov.Well = ins.WellTo
	mov.Pos = ins.PltTo
	mov.Plt = ins.TPlateType
	mov.WVolume = ins.TVolume
	mov.Head = ins.Prms.Head
	mov.Reference = append(mov.Reference, pol["BLOWOUTREFERENCE"].(int))
	mov.OffsetZ = append(mov.OffsetZ, pol["BLOWOUTOFFSET"].(float64))

	blow := NewBlowoutInstruction()

	blow.Head = ins.Prms.Head
	bov := wunit.NewVolume(pol["BLOWOUTVOLUME"].(float64), pol["BLOWOUTVOLUMEUNIT"].(string))
	blow.Volume = append(blow.Volume, &bov)
	blow.Multi = len(ins.What)
	blow.Plt = ins.TPlateType
	blow.What = ins.What

	mov2 := NewMoveInstruction()
	mov2.Well = ins.WellTo
	mov2.Pos = ins.PltTo
	mov2.Plt = ins.TPlateType
	mov2.WVolume = ins.TVolume
	mov2.Head = ins.Prms.Head
	mov2.Reference = append(mov2.Reference, pol["PTZREFERENCE"].(int))
	mov2.OffsetZ = append(mov2.OffsetZ, pol["PTZOFFSET"].(float64))

	ptz := NewPTZInstruction()

	ptz.Head = ins.Prms.Head
	ptz.Channel = -1 // all channels

	ret = append(ret, mov)
	ret = append(ret, blow)
	ret = append(ret, mov2)
	ret = append(ret, ptz)
	return ret
}

type MoveMixInstruction struct {
	Type      int
	Head      int
	Plt       []string
	Well      []string
	Volume    []*wunit.Volume
	PlateType []string
	FVolume   []*wunit.Volume
	Cycles    []int
	Multi     int
	Prms      map[string]interface{}
}

func NewMoveMixInstruction() *MoveMixInstruction {
	var mi MoveMixInstruction

	mi.Type = MMX
	mi.Plt = make([]string, 0)
	mi.Well = make([]string, 0)
	mi.Volume = make([]*wunit.Volume, 0)
	mi.FVolume = make([]*wunit.Volume, 0)
	mi.PlateType = make([]string, 0)
	mi.Cycles = make([]int, 0)
	mi.Prms = make(map[string]interface{})
	return &mi
}

func (ins *MoveMixInstruction) GetParameter(name string) interface{} {
	switch name {
	case "VOLUME":
		return ins.Volume
	case "VOLUNT":
		return nil
	case "PLATETYPE":
		return ins.PlateType
	case "WELLVOLUME":
		return ins.FVolume
	case "POS":
		return ins.Plt
	case "WELL":
		return ins.Well
	case "PARAMS":
		return ins.Prms
	case "CYCLES":
		return ins.Cycles
	case "INSTRUCTIONTYPE":
		return ins.InstructionType()
	}
	return nil

}

func (ins *MoveMixInstruction) InstructionType() int {
	return MMX
}

func (ins *MoveMixInstruction) Generate(policy *LHPolicyRuleSet, prms *LHProperties) []RobotInstruction {
	ret := make([]RobotInstruction, 2)

	// move

	mov := NewMoveInstruction()
	mov.Well = ins.Well
	mov.Pos = ins.Plt
	mov.Plt = ins.PlateType
	mov.WVolume = ins.FVolume
	mov.Head = ins.Head
	zoff := make([]float64, ins.Multi)
	zoff[0] = 0.5
	mov.OffsetZ = zoff
	ret[0] = mov

	// mix

	mix := NewMixInstruction()
	mix.Head = ins.Head
	mix.PlateType = ins.PlateType
	mix.Cycles = ins.Cycles
	mix.Volume = ins.Volume
	mix.PlateType = ins.PlateType
	mix.FVolume = ins.FVolume
	mix.Multi = ins.Multi
	ret[1] = mix

	return ret
}

type MixInstruction struct {
	Type      int
	Head      int
	Volume    []*wunit.Volume
	FVolume   []*wunit.Volume
	PlateType []string
	Multi     int
	Cycles    []int
	Prms      map[string]interface{}
}

func NewMixInstruction() *MixInstruction {
	var mi MixInstruction

	mi.Type = MMX
	mi.Volume = make([]*wunit.Volume, 0)
	mi.FVolume = make([]*wunit.Volume, 0)
	mi.PlateType = make([]string, 0)
	mi.Cycles = make([]int, 0)
	mi.Prms = make(map[string]interface{})
	return &mi
}

func (mi *MixInstruction) InstructionType() int {
	return mi.Type
}

func (ins *MixInstruction) Generate(policy *LHPolicyRuleSet, prms *LHProperties) []RobotInstruction {
	return nil
}

func (ins *MixInstruction) GetParameter(name string) interface{} {
	switch name {
	case "VOLUME":
		return ins.Volume
	case "VOLUNT":
		return nil
	case "PLATETYPE":
		return ins.PlateType
	case "WELLVOLUME":
		return ins.FVolume
	case "PARAMS":
		return ins.Prms
	case "CYCLES":
		return ins.Cycles
	case "INSTRUCTIONTYPE":
		return ins.InstructionType()
	}
	return nil

}

func (mi *MixInstruction) OutputTo(driver LiquidhandlingDriver) {
	vols := make([]float64, len(mi.Volume))
	fvols := make([]float64, len(mi.Volume))

	for i := 0; i < len(mi.Volume); i++ {
		vols[i] = mi.Volume[i].ConvertTo(wunit.ParsePrefixedUnit("ul"))
		fvols[i] = mi.Volume[i].ConvertTo(wunit.ParsePrefixedUnit("ul"))
	}

	driver.Mix(mi.Head, vols, fvols, mi.PlateType, mi.Cycles, mi.Multi, mi.Prms)
}

// TODO -- implement MESSAGE

func GetTips(tiptype string, params *LHProperties, channel *wtype.LHChannelParameter, multi int, mirror bool) RobotInstruction {
	tipwells, tipboxpositions, tipboxtypes := params.GetCleanTips(tiptype, channel, mirror, multi)

	if tipwells == nil {
		panic("NO TIPS LEFT BOYO")
	}

	ins := NewLoadTipsMoveInstruction()
	ins.Head = channel.Head
	ins.Well = tipwells
	ins.FPosition = tipboxpositions
	ins.FPlateType = tipboxtypes
	ins.Multi = multi
	return ins
}

func DropTips(tiptype string, params *LHProperties, channel *wtype.LHChannelParameter, multi int) RobotInstruction {
	tipwells, tipwastepositions, tipwastetypes := params.DropDirtyTips(channel, multi)

	if tipwells == nil {
		panic("NO ROOM AT THE INN FOR THESE LITTLE TIPS")
	}

	ins := NewUnloadTipsMoveInstruction()
	ins.Head = channel.Head
	ins.WellTo = tipwells
	ins.PltTo = tipwastepositions
	ins.TPlateType = tipwastetypes
	ins.Multi = multi
	return ins
}
