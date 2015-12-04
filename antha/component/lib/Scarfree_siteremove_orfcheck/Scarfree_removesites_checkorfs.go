// This protocol is intended to design assembly parts using a specified enzyme.
// overhangs are added to complement the adjacent parts and leave no scar.
// parts can be entered as genbank (.gb) files, sequences or biobrick IDs
// If assembly simulation fails after overhangs are added. In order to help the user
// diagnose the reason, a report of the part overhangs
// is returned to the user along with a list of cut sites in each part.

package Scarfree_siteremove_orfcheck

import (
	"encoding/json"
	"fmt"
	"github.com/antha-lang/antha/antha/AnthaStandardLibrary/Packages/AnthaPath"
	"github.com/antha-lang/antha/antha/AnthaStandardLibrary/Packages/Parser"
	"github.com/antha-lang/antha/antha/AnthaStandardLibrary/Packages/enzymes"
	"github.com/antha-lang/antha/antha/AnthaStandardLibrary/Packages/enzymes/lookup"
	"github.com/antha-lang/antha/antha/AnthaStandardLibrary/Packages/igem"
	"github.com/antha-lang/antha/antha/AnthaStandardLibrary/Packages/sequences"
	"github.com/antha-lang/antha/antha/AnthaStandardLibrary/Packages/text"
	"github.com/antha-lang/antha/antha/anthalib/wtype"
	"github.com/antha-lang/antha/antha/anthalib/wunit"
	"github.com/antha-lang/antha/antha/execute"
	"github.com/antha-lang/antha/flow"
	"github.com/antha-lang/antha/microArch/execution"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
)

// Input parameters for this protocol (data)

// enter each as amino acid sequence

// Physical Inputs to this protocol with types

// Physical outputs from this protocol with types

// Data which is returned from this protocol, and data types

// parts to order
// desired sequence to end up with after assembly

// Input Requirement specification
func (e *Scarfree_siteremove_orfcheck) requirements() {
	_ = wunit.Make_units

	// e.g. are MoClo types valid?
}

// Conditions to run on startup
func (e *Scarfree_siteremove_orfcheck) setup(p Scarfree_siteremove_orfcheckParamBlock) {
	_wrapper := execution.NewWrapper(p.ID,
		p.BlockID, p)
	_ = _wrapper
	_ = _wrapper.WaitToEnd()

}

// The core process for this protocol, with the steps to be performed
// for every input
func (e *Scarfree_siteremove_orfcheck) steps(p Scarfree_siteremove_orfcheckParamBlock, r *Scarfree_siteremove_orfcheckResultBlock) {
	_wrapper := execution.NewWrapper(p.ID,
		p.BlockID, p)
	_ = _wrapper

	//var msg string
	// set warnings reported back to user to none initially
	warnings := make([]string, 0)

	var warning string
	var err error
	// make an empty array of DNA Sequences ready to fill
	partsinorder := make([]wtype.DNASequence, 0)

	var partDNA wtype.DNASequence

	r.Status = "all parts available"
	for i, part := range p.Seqsinorder {
		if strings.Contains(part, ".gb") && strings.Contains(part, "Feature:") {

			split := strings.SplitAfter(part, ".gb")
			file := split[0]

			split2 := strings.Split(split[1], ":")
			feature := split2[1]

			partDNA, _ = parser.GenbankFeaturetoDNASequence(file, feature)
		} else if strings.Contains(part, ".gb") {

			/*annotated,_ := parser.GenbanktoAnnotatedSeq(part)
			partDNA = annotated.DNASequence */

			partDNA, _ = parser.GenbanktoDNASequence(part)
		} else {

			if strings.Contains(part, "BBa_") {
				part = igem.GetSequence(part)
			}
			partDNA = wtype.MakeLinearDNASequence("Part "+strconv.Itoa(i), part)
		}
		partsinorder = append(partsinorder, partDNA)
	}
	// check parts for restriction sites first and remove if the user has chosen to
	enz := lookup.EnzymeLookup(p.Enzymename)

	warning = text.Print("RemoveproblemRestrictionSites =", p.RemoveproblemRestrictionSites)
	warnings = append(warnings, warning)
	if p.RemoveproblemRestrictionSites {
		newparts := make([]wtype.DNASequence, 0)
		warning = "Starting process or removing restrictionsite"
		warnings = append(warnings, warning)

		for _, part := range partsinorder {
			fmt.Println("PARRRTS:", part)
			info := enzymes.Restrictionsitefinder(part, []wtype.LogicalRestrictionEnzyme{enz})

			for _, anysites := range info {
				if anysites.Sitefound {
					warning = "problem site found in " + part.Nm
					warnings = append(warnings, warning)
					orf, orftrue := sequences.FindBiggestORF(part.Seq)
					warning = fmt.Sprintln("site found in orf ", part.Nm, " ", orftrue, " site positions ", anysites.Positions("ALL"), "orf between", orf.StartPosition, " and ", orf.EndPosition /*orf.DNASeq[orf.StartPosition:orf.EndPosition]*/)
					warnings = append(warnings, warning)
					if orftrue /* && len(orf.ProtSeq) > 20 */ {
						allsitestoavoid := make([]string, 0)
						allsitestoavoid = append(allsitestoavoid, anysites.Recognitionsequence, sequences.RevComp(anysites.Recognitionsequence))
						orfcoordinates := sequences.MakeStartendPair(orf.StartPosition, orf.EndPosition)
						for _, position := range anysites.Positions("ALL") {
							if orf.StartPosition < position && position < orf.EndPosition {
								originalcodon := ""
								codonoption := ""
								part, originalcodon, codonoption, err = sequences.ReplaceCodoninORF(part, orfcoordinates, position, allsitestoavoid)
								warning = fmt.Sprintln("sites to avoid: ", allsitestoavoid[0], allsitestoavoid[1])
								warnings = append(warnings, warning)
								warnings = append(warnings, "Paaaaerrttseq: "+part.Seq+"position: "+strconv.Itoa(position)+" original: "+originalcodon+" replacementcodon: "+codonoption)
								if err != nil {
									warning := text.Print("removal of site from orf "+orf.DNASeq, " failed! improve your algorithm! "+err.Error())
									warnings = append(warnings, warning)
								}
							} else {
								allsitestoavoid := make([]string, 0)
								part, err = sequences.RemoveSite(part, anysites.Enzyme, allsitestoavoid)
								if err != nil {

									warning = text.Print("position found to be outside of orf: "+orf.DNASeq, " failed! improve your algorithm! "+err.Error())
									warnings = append(warnings, warning)
								}
							}
						}
					} else {
						allsitestoavoid := make([]string, 0)
						temppart, err := sequences.RemoveSite(part, anysites.Enzyme, allsitestoavoid)
						fmt.Println("part= ", part)
						fmt.Println("temppart= ", temppart)
						if err != nil {
							warning := text.Print("removal of site failed! improve your algorithm!", err.Error())
							warnings = append(warnings, warning)

						}
						warning = fmt.Sprintln("modified "+temppart.Nm+"new seq: ", temppart.Seq)
						warnings = append(warnings, warning)
						part = temppart

						//	}
					}
				}
				//part.Seq = ""
				newparts = append(newparts, part)

				partsinorder = newparts
			}
		}
	}
	// make vector into an antha type DNASequence
	vectordata := wtype.MakePlasmidDNASequence("Vector", p.Vector)

	//lookup restriction enzyme
	restrictionenzyme, err := lookup.TypeIIsLookup(p.Enzymename)
	if err != nil {
		warnings = append(warnings, text.Print("Error", err.Error()))
	}

	//  Add overhangs for scarfree assembly based on part seqeunces only, i.e. no Assembly standard
	fmt.Println("warnings:", warnings)
	r.PartswithOverhangs = enzymes.MakeScarfreeCustomTypeIIsassemblyParts(partsinorder, vectordata, restrictionenzyme)

	// Check that assembly is feasible with designed parts by simulating assembly of the sequences with the chosen enzyme
	assembly := enzymes.Assemblyparameters{p.Constructname, restrictionenzyme.Name, vectordata, r.PartswithOverhangs}
	status, numberofassemblies, _, newDNASequence, err := enzymes.Assemblysimulator(assembly)

	endreport := "Endreport only run in the event of assembly simulation failure"
	//sites := "Restriction mapper only run in the event of assembly simulation failure"
	r.NewDNASequence = newDNASequence
	if err == nil && numberofassemblies == 1 {

		r.Simulationpass = true
	} else {

		warnings = append(warnings, status)
		// perform mock digest to test fragement overhangs (fragments are hidden by using _, )
		_, stickyends5, stickyends3 := enzymes.TypeIIsdigest(vectordata, restrictionenzyme)

		allends := make([]string, 0)
		ends := ""

		ends = text.Print(vectordata.Nm+" 5 Prime end: ", stickyends5)
		allends = append(allends, ends)
		ends = text.Print(vectordata.Nm+" 3 Prime end: ", stickyends3)
		allends = append(allends, ends)

		for _, part := range r.PartswithOverhangs {
			_, stickyends5, stickyends3 := enzymes.TypeIIsdigest(part, restrictionenzyme)
			ends = text.Print(part.Nm+" 5 Prime end: ", stickyends5)
			allends = append(allends, ends)
			ends = text.Print(part.Nm+" 3 Prime end: ", stickyends3)
			allends = append(allends, ends)
		}
		endreport = strings.Join(allends, " ")
		warnings = append(warnings, endreport)
	}

	// check number of sites per part !

	sites := make([]int, 0)
	multiple := make([]string, 0)
	for _, part := range r.PartswithOverhangs {

		info := enzymes.Restrictionsitefinder(part, []wtype.LogicalRestrictionEnzyme{enz})

		sitepositions := enzymes.SitepositionString(info[0])

		sites = append(sites, info[0].Numberofsites)
		sitepositions = text.Print(part.Nm+" "+p.Enzymename+" positions:", sitepositions)
		multiple = append(multiple, sitepositions)
	}

	for _, orf := range p.ORFstoConfirm {
		if sequences.LookforSpecificORF(r.NewDNASequence.Seq, orf) == false {
			warning = text.Print("orf not present: ", orf)
			warnings = append(warnings, warning)
			r.ORFmissing = true
		}
	}

	if len(warnings) == 0 {
		warnings = append(warnings, "none")
	}
	r.Warnings = fmt.Errorf(strings.Join(warnings, ";"))

	partsummary := make([]string, 0)
	for _, part := range r.PartswithOverhangs {
		partsummary = append(partsummary, text.Print(part.Nm, part.Seq))
	}

	partstoorder := text.Print("PartswithOverhangs: ", partsummary)

	// Print status
	if r.Status != "all parts available" {
		r.Status = fmt.Sprintln(r.Status)
	} else {
		r.Status = fmt.Sprintln(
			text.Print("simulator status: ", status),
			text.Print("Endreport after digestion: ", endreport),
			text.Print("Sites per part for "+p.Enzymename, sites),
			text.Print("Positions: ", multiple),
			text.Print("Warnings:", r.Warnings.Error()),
			text.Print("Simulationpass=", r.Simulationpass),
			text.Print("NewDNASequence: ", r.NewDNASequence),
			text.Print("Any Orfs to confirm missing from new DNA sequence:", r.ORFmissing),
			partstoorder,
		)
		// export data to file
		anthapath.ExporttoFile("Report"+"_"+p.Constructname+".txt", []byte(r.Status))
		//anthapath.ExportTextFile("Report"+"_"+Constructname+".txt",Status)

	}
	_ = _wrapper.WaitToEnd()

}

// Run after controls and a steps block are completed to
// post process any data and provide downstream results
func (e *Scarfree_siteremove_orfcheck) analysis(p Scarfree_siteremove_orfcheckParamBlock, r *Scarfree_siteremove_orfcheckResultBlock) {
	_wrapper := execution.NewWrapper(p.ID,
		p.BlockID, p)
	_ = _wrapper
	_ = _wrapper.WaitToEnd()

}

// A block of tests to perform to validate that the sample was processed correctly
// Optionally, destructive tests can be performed to validate results on a
// dipstick basis
func (e *Scarfree_siteremove_orfcheck) validation(p Scarfree_siteremove_orfcheckParamBlock, r *Scarfree_siteremove_orfcheckResultBlock) {
	_wrapper := execution.NewWrapper(p.ID,
		p.BlockID, p)
	_ = _wrapper
	_ = _wrapper.WaitToEnd()

}

// AsyncBag functions
func (e *Scarfree_siteremove_orfcheck) Complete(params interface{}) {
	p := params.(Scarfree_siteremove_orfcheckParamBlock)
	if p.Error {
		e.NewDNASequence <- execute.ThreadParam{Value: nil, ID: p.ID, Error: true}
		e.ORFmissing <- execute.ThreadParam{Value: nil, ID: p.ID, Error: true}
		e.PartswithOverhangs <- execute.ThreadParam{Value: nil, ID: p.ID, Error: true}
		e.Simulationpass <- execute.ThreadParam{Value: nil, ID: p.ID, Error: true}
		e.Status <- execute.ThreadParam{Value: nil, ID: p.ID, Error: true}
		e.Warnings <- execute.ThreadParam{Value: nil, ID: p.ID, Error: true}
		return
	}
	r := new(Scarfree_siteremove_orfcheckResultBlock)
	defer func() {
		if res := recover(); res != nil {
			e.NewDNASequence <- execute.ThreadParam{Value: res, ID: p.ID, Error: true}
			e.ORFmissing <- execute.ThreadParam{Value: res, ID: p.ID, Error: true}
			e.PartswithOverhangs <- execute.ThreadParam{Value: res, ID: p.ID, Error: true}
			e.Simulationpass <- execute.ThreadParam{Value: res, ID: p.ID, Error: true}
			e.Status <- execute.ThreadParam{Value: res, ID: p.ID, Error: true}
			e.Warnings <- execute.ThreadParam{Value: res, ID: p.ID, Error: true}
			execute.AddError(&execute.RuntimeError{BaseError: res, Stack: debug.Stack()})
			return
		}
	}()
	e.startup.Do(func() { e.setup(p) })
	e.steps(p, r)

	e.NewDNASequence <- execute.ThreadParam{Value: r.NewDNASequence, ID: p.ID, Error: false}

	e.ORFmissing <- execute.ThreadParam{Value: r.ORFmissing, ID: p.ID, Error: false}

	e.PartswithOverhangs <- execute.ThreadParam{Value: r.PartswithOverhangs, ID: p.ID, Error: false}

	e.Simulationpass <- execute.ThreadParam{Value: r.Simulationpass, ID: p.ID, Error: false}

	e.Status <- execute.ThreadParam{Value: r.Status, ID: p.ID, Error: false}

	e.Warnings <- execute.ThreadParam{Value: r.Warnings, ID: p.ID, Error: false}

	e.analysis(p, r)

	e.validation(p, r)

}

// init function, read characterization info from seperate file to validate ranges?
func (e *Scarfree_siteremove_orfcheck) init() {
	e.params = make(map[execute.ThreadID]*execute.AsyncBag)
}

func (e *Scarfree_siteremove_orfcheck) NewConfig() interface{} {
	return &Scarfree_siteremove_orfcheckConfig{}
}

func (e *Scarfree_siteremove_orfcheck) NewParamBlock() interface{} {
	return &Scarfree_siteremove_orfcheckParamBlock{}
}

func NewScarfree_siteremove_orfcheck() interface{} { //*Scarfree_siteremove_orfcheck {
	e := new(Scarfree_siteremove_orfcheck)
	e.init()
	return e
}

// Mapper function
func (e *Scarfree_siteremove_orfcheck) Map(m map[string]interface{}) interface{} {
	var res Scarfree_siteremove_orfcheckParamBlock
	res.Error = false || m["Constructname"].(execute.ThreadParam).Error || m["Enzymename"].(execute.ThreadParam).Error || m["ORFstoConfirm"].(execute.ThreadParam).Error || m["RemoveproblemRestrictionSites"].(execute.ThreadParam).Error || m["Seqsinorder"].(execute.ThreadParam).Error || m["Vector"].(execute.ThreadParam).Error

	vConstructname, is := m["Constructname"].(execute.ThreadParam).Value.(execute.JSONValue)
	if is {
		var temp Scarfree_siteremove_orfcheckJSONBlock
		json.Unmarshal([]byte(vConstructname.JSONString), &temp)
		res.Constructname = *temp.Constructname
	} else {
		res.Constructname = m["Constructname"].(execute.ThreadParam).Value.(string)
	}

	vEnzymename, is := m["Enzymename"].(execute.ThreadParam).Value.(execute.JSONValue)
	if is {
		var temp Scarfree_siteremove_orfcheckJSONBlock
		json.Unmarshal([]byte(vEnzymename.JSONString), &temp)
		res.Enzymename = *temp.Enzymename
	} else {
		res.Enzymename = m["Enzymename"].(execute.ThreadParam).Value.(string)
	}

	vORFstoConfirm, is := m["ORFstoConfirm"].(execute.ThreadParam).Value.(execute.JSONValue)
	if is {
		var temp Scarfree_siteremove_orfcheckJSONBlock
		json.Unmarshal([]byte(vORFstoConfirm.JSONString), &temp)
		res.ORFstoConfirm = *temp.ORFstoConfirm
	} else {
		res.ORFstoConfirm = m["ORFstoConfirm"].(execute.ThreadParam).Value.([]string)
	}

	vRemoveproblemRestrictionSites, is := m["RemoveproblemRestrictionSites"].(execute.ThreadParam).Value.(execute.JSONValue)
	if is {
		var temp Scarfree_siteremove_orfcheckJSONBlock
		json.Unmarshal([]byte(vRemoveproblemRestrictionSites.JSONString), &temp)
		res.RemoveproblemRestrictionSites = *temp.RemoveproblemRestrictionSites
	} else {
		res.RemoveproblemRestrictionSites = m["RemoveproblemRestrictionSites"].(execute.ThreadParam).Value.(bool)
	}

	vSeqsinorder, is := m["Seqsinorder"].(execute.ThreadParam).Value.(execute.JSONValue)
	if is {
		var temp Scarfree_siteremove_orfcheckJSONBlock
		json.Unmarshal([]byte(vSeqsinorder.JSONString), &temp)
		res.Seqsinorder = *temp.Seqsinorder
	} else {
		res.Seqsinorder = m["Seqsinorder"].(execute.ThreadParam).Value.([]string)
	}

	vVector, is := m["Vector"].(execute.ThreadParam).Value.(execute.JSONValue)
	if is {
		var temp Scarfree_siteremove_orfcheckJSONBlock
		json.Unmarshal([]byte(vVector.JSONString), &temp)
		res.Vector = *temp.Vector
	} else {
		res.Vector = m["Vector"].(execute.ThreadParam).Value.(string)
	}

	res.ID = m["Constructname"].(execute.ThreadParam).ID
	res.BlockID = m["Constructname"].(execute.ThreadParam).BlockID

	return res
}

func (e *Scarfree_siteremove_orfcheck) OnConstructname(param execute.ThreadParam) {
	e.lock.Lock()
	var bag *execute.AsyncBag = e.params[param.ID]
	if bag == nil {
		bag = new(execute.AsyncBag)
		bag.Init(6, e, e)
		e.params[param.ID] = bag
	}
	e.lock.Unlock()

	fired := bag.AddValue("Constructname", param)
	if fired {
		e.lock.Lock()
		delete(e.params, param.ID)
		e.lock.Unlock()
	}
}
func (e *Scarfree_siteremove_orfcheck) OnEnzymename(param execute.ThreadParam) {
	e.lock.Lock()
	var bag *execute.AsyncBag = e.params[param.ID]
	if bag == nil {
		bag = new(execute.AsyncBag)
		bag.Init(6, e, e)
		e.params[param.ID] = bag
	}
	e.lock.Unlock()

	fired := bag.AddValue("Enzymename", param)
	if fired {
		e.lock.Lock()
		delete(e.params, param.ID)
		e.lock.Unlock()
	}
}
func (e *Scarfree_siteremove_orfcheck) OnORFstoConfirm(param execute.ThreadParam) {
	e.lock.Lock()
	var bag *execute.AsyncBag = e.params[param.ID]
	if bag == nil {
		bag = new(execute.AsyncBag)
		bag.Init(6, e, e)
		e.params[param.ID] = bag
	}
	e.lock.Unlock()

	fired := bag.AddValue("ORFstoConfirm", param)
	if fired {
		e.lock.Lock()
		delete(e.params, param.ID)
		e.lock.Unlock()
	}
}
func (e *Scarfree_siteremove_orfcheck) OnRemoveproblemRestrictionSites(param execute.ThreadParam) {
	e.lock.Lock()
	var bag *execute.AsyncBag = e.params[param.ID]
	if bag == nil {
		bag = new(execute.AsyncBag)
		bag.Init(6, e, e)
		e.params[param.ID] = bag
	}
	e.lock.Unlock()

	fired := bag.AddValue("RemoveproblemRestrictionSites", param)
	if fired {
		e.lock.Lock()
		delete(e.params, param.ID)
		e.lock.Unlock()
	}
}
func (e *Scarfree_siteremove_orfcheck) OnSeqsinorder(param execute.ThreadParam) {
	e.lock.Lock()
	var bag *execute.AsyncBag = e.params[param.ID]
	if bag == nil {
		bag = new(execute.AsyncBag)
		bag.Init(6, e, e)
		e.params[param.ID] = bag
	}
	e.lock.Unlock()

	fired := bag.AddValue("Seqsinorder", param)
	if fired {
		e.lock.Lock()
		delete(e.params, param.ID)
		e.lock.Unlock()
	}
}
func (e *Scarfree_siteremove_orfcheck) OnVector(param execute.ThreadParam) {
	e.lock.Lock()
	var bag *execute.AsyncBag = e.params[param.ID]
	if bag == nil {
		bag = new(execute.AsyncBag)
		bag.Init(6, e, e)
		e.params[param.ID] = bag
	}
	e.lock.Unlock()

	fired := bag.AddValue("Vector", param)
	if fired {
		e.lock.Lock()
		delete(e.params, param.ID)
		e.lock.Unlock()
	}
}

type Scarfree_siteremove_orfcheck struct {
	flow.Component                // component "superclass" embedded
	lock                          sync.Mutex
	startup                       sync.Once
	params                        map[execute.ThreadID]*execute.AsyncBag
	Constructname                 <-chan execute.ThreadParam
	Enzymename                    <-chan execute.ThreadParam
	ORFstoConfirm                 <-chan execute.ThreadParam
	RemoveproblemRestrictionSites <-chan execute.ThreadParam
	Seqsinorder                   <-chan execute.ThreadParam
	Vector                        <-chan execute.ThreadParam
	NewDNASequence                chan<- execute.ThreadParam
	ORFmissing                    chan<- execute.ThreadParam
	PartswithOverhangs            chan<- execute.ThreadParam
	Simulationpass                chan<- execute.ThreadParam
	Status                        chan<- execute.ThreadParam
	Warnings                      chan<- execute.ThreadParam
}

type Scarfree_siteremove_orfcheckParamBlock struct {
	ID                            execute.ThreadID
	BlockID                       execute.BlockID
	Error                         bool
	Constructname                 string
	Enzymename                    string
	ORFstoConfirm                 []string
	RemoveproblemRestrictionSites bool
	Seqsinorder                   []string
	Vector                        string
}

type Scarfree_siteremove_orfcheckConfig struct {
	ID                            execute.ThreadID
	BlockID                       execute.BlockID
	Error                         bool
	Constructname                 string
	Enzymename                    string
	ORFstoConfirm                 []string
	RemoveproblemRestrictionSites bool
	Seqsinorder                   []string
	Vector                        string
}

type Scarfree_siteremove_orfcheckResultBlock struct {
	ID                 execute.ThreadID
	BlockID            execute.BlockID
	Error              bool
	NewDNASequence     wtype.DNASequence
	ORFmissing         bool
	PartswithOverhangs []wtype.DNASequence
	Simulationpass     bool
	Status             string
	Warnings           error
}

type Scarfree_siteremove_orfcheckJSONBlock struct {
	ID                            *execute.ThreadID
	BlockID                       *execute.BlockID
	Error                         *bool
	Constructname                 *string
	Enzymename                    *string
	ORFstoConfirm                 *[]string
	RemoveproblemRestrictionSites *bool
	Seqsinorder                   *[]string
	Vector                        *string
	NewDNASequence                *wtype.DNASequence
	ORFmissing                    *bool
	PartswithOverhangs            *[]wtype.DNASequence
	Simulationpass                *bool
	Status                        *string
	Warnings                      *error
}

func (c *Scarfree_siteremove_orfcheck) ComponentInfo() *execute.ComponentInfo {
	inp := make([]execute.PortInfo, 0)
	outp := make([]execute.PortInfo, 0)
	inp = append(inp, *execute.NewPortInfo("Constructname", "string", "Constructname", true, true, nil, nil))
	inp = append(inp, *execute.NewPortInfo("Enzymename", "string", "Enzymename", true, true, nil, nil))
	inp = append(inp, *execute.NewPortInfo("ORFstoConfirm", "[]string", "ORFstoConfirm", true, true, nil, nil))
	inp = append(inp, *execute.NewPortInfo("RemoveproblemRestrictionSites", "bool", "RemoveproblemRestrictionSites", true, true, nil, nil))
	inp = append(inp, *execute.NewPortInfo("Seqsinorder", "[]string", "Seqsinorder", true, true, nil, nil))
	inp = append(inp, *execute.NewPortInfo("Vector", "string", "Vector", true, true, nil, nil))
	outp = append(outp, *execute.NewPortInfo("NewDNASequence", "wtype.DNASequence", "NewDNASequence", true, true, nil, nil))
	outp = append(outp, *execute.NewPortInfo("ORFmissing", "bool", "ORFmissing", true, true, nil, nil))
	outp = append(outp, *execute.NewPortInfo("PartswithOverhangs", "[]wtype.DNASequence", "PartswithOverhangs", true, true, nil, nil))
	outp = append(outp, *execute.NewPortInfo("Simulationpass", "bool", "Simulationpass", true, true, nil, nil))
	outp = append(outp, *execute.NewPortInfo("Status", "string", "Status", true, true, nil, nil))
	outp = append(outp, *execute.NewPortInfo("Warnings", "error", "Warnings", true, true, nil, nil))

	ci := execute.NewComponentInfo("Scarfree_siteremove_orfcheck", "Scarfree_siteremove_orfcheck", "", false, inp, outp)

	return ci
}