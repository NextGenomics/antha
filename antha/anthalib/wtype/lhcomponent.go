// liquidhandling/lhtypes.Go: Part of the Antha language
// Copyright (C) 2014 the Antha authors. All rights reserved.
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
// contact license@antha-lang.Org or write to the Antha team c/o
// Synthace Ltd. The London Bioscience Innovation Centre
// 2 Royal College St, London NW1 0NH UK

// defines types for dealing with liquid handling requests
package wtype

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/antha-lang/antha/antha/anthalib/wunit"
	"github.com/antha-lang/antha/microArch/logger"
)

// structure describing a liquid component and its desired properties
type LHComponent struct {
	ID                 string
	BlockID            BlockID
	DaughterID         string
	ParentID           string
	Inst               string
	Order              int
	CName              string
	Type               LiquidType
	Vol                float64
	Conc               float64
	Vunit              string
	Cunit              string
	Tvol               float64
	Smax               float64
	Visc               float64
	StockConcentration float64
	Extra              map[string]interface{}
	Loc                string
	Destination        string
}

func (lhc *LHComponent) Generation() int {
	g, ok := lhc.Extra["Generation"]

	if ok {
		return g.(int)
	}

	return 0
}

func (lhc *LHComponent) SetGeneration(i int) {
	lhc.Extra["Generation"] = i
}

func (lhc *LHComponent) IsZero() bool {
	if lhc == nil || lhc.Type == LTNIL || lhc.CName == "" || lhc.Vol < 0.0000001 {
		return true
	}
	return false
}

func (lhc *LHComponent) SetVolume(v wunit.Volume) {
	lhc.Vol = v.RawValue()
	lhc.Vunit = v.Unit().PrefixedSymbol()
}

func (lhc *LHComponent) HasParent(s string) bool {
	return strings.Contains(lhc.ParentID, s)
}

func (lhc *LHComponent) HasDaughter(s string) bool {
	return strings.Contains(lhc.DaughterID, s)
}

func (lhc *LHComponent) Name() string {
	return lhc.CName
}

func (lhc *LHComponent) TypeName() string {
	return LiquidTypeName(lhc.Type)
}

func (lhc *LHComponent) Volume() wunit.Volume {
	return wunit.NewVolume(lhc.Vol, lhc.Vunit)
}

func (lhc *LHComponent) Remove(v wunit.Volume) {
	///TODO -- catch errors
	lhc.Vol -= v.ConvertToString(lhc.Vunit)

	if lhc.Vol < 0.0 {
		lhc.Vol = 0.0
	}
}

func (lhc *LHComponent) Dup() *LHComponent {
	c := NewLHComponent()
	c.ID = lhc.ID
	c.Order = lhc.Order
	c.CName = lhc.CName
	c.Type = lhc.Type
	c.Vol = lhc.Vol
	c.Conc = lhc.Conc
	c.Vunit = lhc.Vunit
	c.Tvol = lhc.Vol
	c.Smax = lhc.Smax
	c.Visc = lhc.Visc
	c.StockConcentration = lhc.StockConcentration
	c.Extra = make(map[string]interface{}, len(lhc.Extra))
	for k, v := range lhc.Extra {
		c.Extra[k] = v
	}
	c.Loc = lhc.Loc
	c.Destination = lhc.Destination
	return c
}

func (cmp *LHComponent) SetSample(flag bool) bool {
	if cmp == nil {
		return false
	}

	if cmp.Extra == nil {
		cmp.Extra = make(map[string]interface{})
	}

	cmp.Extra["IsSample"] = flag

	return true
}

func (cmp *LHComponent) IsSample() bool {
	if cmp == nil {
		return false
	}

	f, ok := cmp.Extra["IsSample"]

	if !ok || !f.(bool) {
		return false
	}

	return true
}

func (cmp *LHComponent) HasAnyParent() bool {
	if cmp.ParentID != "" {
		return true
	}

	return false
}

func (cmp *LHComponent) AddParent(parentID string) {
	cmp.ParentID += parentID + "_"
}

func (cmp *LHComponent) AddDaughter(daughterID string) {
	cmp.DaughterID += daughterID + "_"
}

func (cmp *LHComponent) Mix(cmp2 *LHComponent) {
	// if this component is zero we inherit the id of the other one
	// unless the other component is a sample: in this case it retains
	// the parent ID so this would not be safe
	// do I want to do this at all?
	/*
		if cmp.IsZero() && !cmp2.IsSample() {
			fmt.Println("CMP IS ZERO: REDEFINING ID")
			cmp.ID = cmp2.ID
		}
	*/
	cmp.Smax = mergeSolubilities(cmp, cmp2)
	// determine type of final
	cmp.Type = mergeTypes(cmp, cmp2)
	// add cmp2 to cmp
	vcmp := wunit.NewVolume(cmp.Vol, cmp.Vunit)
	vcmp2 := wunit.NewVolume(cmp2.Vol, cmp2.Vunit)
	vcmp.Add(vcmp2)
	cmp.Vol = vcmp.RawValue() // same units
	cmp.CName = mergeNames(cmp.CName, cmp2.CName)
	// allow trace back
	logger.Track(fmt.Sprintf("MIX %s %s %s", cmp.ID, cmp2.ID, vcmp.ToString()))
}

// @implement Liquid
// @deprecate Liquid

func (lhc *LHComponent) GetSmax() float64 {
	return lhc.Smax
}

func (lhc *LHComponent) GetVisc() float64 {
	return lhc.Visc
}

func (lhc *LHComponent) GetExtra() map[string]interface{} {
	return lhc.Extra
}

func (lhc *LHComponent) GetConc() float64 {
	return lhc.Conc
}

func (lhc *LHComponent) GetCunit() string {
	return lhc.Cunit
}

// new
func (lhc *LHComponent) Concentration() (conc wunit.Concentration) {
	conc = wunit.NewConcentration(lhc.Conc, lhc.Cunit)
	return conc
}

func (lhc *LHComponent) GetVunit() string {
	return lhc.Vunit
}

func (lhc *LHComponent) GetType() string {
	return LiquidTypeName(lhc.Type)
}

func NewLHComponent() *LHComponent {
	var lhc LHComponent
	lhc.ID = GetUUID()
	lhc.Vunit = "ul"
	lhc.Extra = make(map[string]interface{})
	return &lhc
}

// XXX -- why is this different from Dup?
func CopyLHComponent(lhc *LHComponent) *LHComponent {
	tmp, _ := json.Marshal(lhc)
	var lhc2 LHComponent
	json.Unmarshal(tmp, &lhc2)
	lhc2.ID = GetUUID()
	if lhc2.Inst != "" {
		lhc2.Inst = GetUUID()
		// this needs some thought
	}
	return &lhc2
}
