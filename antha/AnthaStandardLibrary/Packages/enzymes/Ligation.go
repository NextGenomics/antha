// antha/AnthaStandardLibrary/Packages/enzymes/Ligation.go: Part of the Antha language
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
// 2 Royal College St, London NW1 0NH UK

package enzymes

import (
	"fmt"
	"github.com/antha-lang/antha/antha/anthalib/wtype"
	"strings"
)

func Jointwoparts(upstreampart []Digestedfragment, downstreampart []Digestedfragment) (assembledfragments []Digestedfragment, plasmidproducts []wtype.DNASequence) {

	sequencestojoin := make([]string, 0)

	assembledfragments = make([]Digestedfragment, 0)
	plasmidproducts = make([]wtype.DNASequence, 0)

	for _, upfragment := range upstreampart {
		for _, downfragment := range downstreampart {
			if RevComp(upfragment.BottomStickyend_5prime) == downfragment.TopStickyend_5prime && RevComp(downfragment.BottomStickyend_5prime) == upfragment.TopStickyend_5prime {
				sequencestojoin = append(sequencestojoin, upfragment.Topstrand, downfragment.Topstrand)
				dnastring := strings.Join(sequencestojoin, "")
				fullyassembledfragment := wtype.DNASequence{Nm: "simulatedassemblysequence", Seq: dnastring, Plasmid: true}
				plasmidproducts = append(plasmidproducts, fullyassembledfragment)
				sequencestojoin = make([]string, 0)
			}
			if upfragment.BottomStickyend_5prime == RevComp(downfragment.BottomStickyend_5prime) && downfragment.TopStickyend_5prime == RevComp(upfragment.TopStickyend_5prime) {
				sequencestojoin = append(sequencestojoin, upfragment.Topstrand, downfragment.Bottomstrand)
				dnastring := strings.Join(sequencestojoin, "")
				fullyassembledfragment := wtype.DNASequence{Nm: "simulatedassemblysequence", Seq: dnastring, Plasmid: true}
				plasmidproducts = append(plasmidproducts, fullyassembledfragment)
				sequencestojoin = make([]string, 0)
			}
			if /*upfragment.BottomStickyend_5prime == RevComp(downfragment.TopStickyend_5prime) ||*/ RevComp(upfragment.BottomStickyend_5prime) == downfragment.TopStickyend_5prime {
				sequencestojoin = append(sequencestojoin, upfragment.Topstrand, downfragment.Topstrand)
				dnastring := strings.Join(sequencestojoin, "")
				assembledfragment := Digestedfragment{dnastring, "", upfragment.TopStickyend_5prime, downfragment.TopStickyend_3prime, downfragment.BottomStickyend_5prime, upfragment.BottomStickyend_3prime}
				assembledfragments = append(assembledfragments, assembledfragment)
				sequencestojoin = make([]string, 0)
			}
			if upfragment.BottomStickyend_5prime == RevComp(downfragment.BottomStickyend_5prime) {
				sequencestojoin = append(sequencestojoin, upfragment.Topstrand, downfragment.Bottomstrand)
				dnastring := strings.Join(sequencestojoin, "")
				assembledfragment := Digestedfragment{dnastring, "", upfragment.TopStickyend_5prime, downfragment.BottomStickyend_3prime, downfragment.TopStickyend_5prime, upfragment.BottomStickyend_3prime}
				assembledfragments = append(assembledfragments, assembledfragment)
				sequencestojoin = make([]string, 0)
			}
		}
	}
	return assembledfragments, plasmidproducts
}

func Jointwopartsfromsequence(vector wtype.DNASequence, part1 wtype.DNASequence, enzyme TypeIIs) (assembledfragments []Digestedfragment, plasmidproducts []wtype.DNASequence) {
	doublestrandedpart1 := MakedoublestrandedDNA(part1)
	digestedpart1 := DigestionPairs(doublestrandedpart1, enzyme)

	doublestrandedvector := MakedoublestrandedDNA(vector)
	digestedvector := DigestionPairs(doublestrandedvector, enzyme)

	assembledfragments, plasmidproducts = Jointwoparts(digestedvector, digestedpart1)

	return assembledfragments, plasmidproducts
}

func JoinXnumberofparts(vector wtype.DNASequence, partsinorder []wtype.DNASequence, enzyme TypeIIs) (assembledfragments []Digestedfragment, plasmidproducts []wtype.DNASequence, err error) {

	if vector.Seq == "" {
		err = fmt.Errorf("No Vector sequence found")
		return
	}
	doublestrandedvector := MakedoublestrandedDNA(vector)
	digestedvector := DigestionPairs(doublestrandedvector, enzyme)

	if len(partsinorder) == 0 {
		err = fmt.Errorf("No parts found")
		return
	}

	doublestrandedpart := MakedoublestrandedDNA(partsinorder[0])
	digestedpart := DigestionPairs(doublestrandedpart, enzyme)
	assembledfragments, plasmidproducts = Jointwoparts(digestedvector, digestedpart)
	for i := 1; i < len(partsinorder); i++ {
		if partsinorder[i].Seq == "" {
			name := partsinorder[i].Nm
			errorstring := name + "has no sequence found"
			err = fmt.Errorf(errorstring)
			return
		}
		doublestrandedpart = MakedoublestrandedDNA(partsinorder[i])
		digestedpart := DigestionPairs(doublestrandedpart, enzyme)
		//for _, newfragments := range assembledfragments {
		assembledfragments, plasmidproducts = Jointwoparts(assembledfragments, digestedpart)
		//}
	}
	return assembledfragments, plasmidproducts, err
}

type Assemblyparameters struct {
	Constructname string
	Enzymename    string
	Vector        wtype.DNASequence
	Partsinorder  []wtype.DNASequence
}

func Assemblysimulator(assemblyparameters Assemblyparameters) (s string, successfulassemblies int, sites []Restrictionsites, newDNASequence wtype.DNASequence, err error) {

	// fetch enzyme properties from map (this is basically a look up table for those who don't know)
	successfulassemblies = 0
	enzymename := strings.ToUpper(assemblyparameters.Enzymename)

	// should change this to rebase lookup; what happens if this fails?
	enzyme := TypeIIsEnzymeproperties[enzymename]
	if enzyme.Name == "" {
		s = "Incorrect Enzyme or no enzyme specified"
		err = fmt.Errorf(s)
		return s, successfulassemblies, sites, newDNASequence, err
	}

	//assemble (note that sapIenz is found in package enzymes)
	failedassemblies, plasmidproductsfromXprimaryseq, err := JoinXnumberofparts(assemblyparameters.Vector, assemblyparameters.Partsinorder, enzyme)

	if err != nil {
		s = err.Error()
		return
	}

	if len(plasmidproductsfromXprimaryseq) == 1 {
		sites = Restrictionsitefinder(plasmidproductsfromXprimaryseq[0], []wtype.LogicalRestrictionEnzyme{BsaI, SapI})
		newDNASequence = plasmidproductsfromXprimaryseq[0]
	}
	if len(plasmidproductsfromXprimaryseq) > 1 {
		sites = make([]Restrictionsites, 0)
		for i := 0; i < len(plasmidproductsfromXprimaryseq); i++ {
			sitesperplasmid := Restrictionsitefinder(plasmidproductsfromXprimaryseq[i], []wtype.LogicalRestrictionEnzyme{BsaI, SapI})
			for _, site := range sitesperplasmid {
				sites = append(sites, site)
			}
		}
		//return first for now
		newDNASequence = plasmidproductsfromXprimaryseq[0]
	}

	s = "hmmm I'm confused, this doesn't seem to make any sense"

	if len(plasmidproductsfromXprimaryseq) == 0 && len(failedassemblies) == 0 {
		s = "Nope! this won't work"
	}
	if len(plasmidproductsfromXprimaryseq) == 1 {
		s = "Yay! this should work"
		successfulassemblies = successfulassemblies + 1
	}

	if len(plasmidproductsfromXprimaryseq) > 1 {
		s = "Yay! this should work but there seems to be more than one possible plasmid which could form"
	}

	if len(plasmidproductsfromXprimaryseq) == 0 && len(failedassemblies) != 0 {
		s = "Ooh, only partial assembly expected"
	}

	for _, assemblyproduct := range plasmidproductsfromXprimaryseq {

		fileprefix := "./"
		tojoin := make([]string, 0)
		tojoin = append(tojoin, fileprefix, assemblyparameters.Constructname)
		filename := strings.Join(tojoin, "")
		Exporttofile(filename, &assemblyproduct)
		ExportFasta(filename, &assemblyproduct)
	}

	if s != "Yay! this should work" {
		err = fmt.Errorf(s)
	}

	return s, successfulassemblies, sites, newDNASequence, err
}

//MultipleAssemblies will perform simulated assemblies on multiple constructs and return a description of whether each was successful and how many are expected to work
func MultipleAssemblies(parameters []Assemblyparameters) (s string, successfulassemblies int, errorDescription map[string]string, err error) {
	allOK := true
	errorDescription = make(map[string]string)

	successfulassemblies = 0
	// for each construct
	for _, construct := range parameters {

		output, _, _, _, err := Assemblysimulator(construct)
		if err != nil {
			allOK = false
			errorDescription[construct.Constructname] = output
		}
		if output == "Yay! this should work" {
			successfulassemblies = successfulassemblies + 1
		}

		s = "not all assemblies seem to work out"
		if successfulassemblies == len(parameters) {
			s = "success, all assemblies seem to work"
		}
	}
	if !allOK {
		err = fmt.Errorf("Not all constructs worked")
	}
	return s, successfulassemblies, errorDescription, err
}
