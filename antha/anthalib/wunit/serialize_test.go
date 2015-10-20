// antha/anthalib/wunit/serialize_test.go: Part of the Antha language
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

package wunit

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestVolumeMarshal(t *testing.T) {
	var v Volume
	var v2 Volume
	var err error
	var enc []byte

	v = NewVolume(5, "l")
	if enc, err = json.Marshal(v); err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(enc, &v2); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(v, v2) {
		t.Fatalf("volumes not equal, expecting %v, got %v", v, v2)
	}

	var v3 Volume

	v = Volume{ConcreteMeasurement{0, nil}}
	if enc, err = json.Marshal(v); err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(enc, &v3); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(v, v3) {
		t.Fatalf("volumes not equal, expecting %v, got %v", v, v3)
	}
}

func TestDeserializeConcreteMeasurement(t *testing.T) {
	str := `{"Mvalue":0,"Munit":null}`
	var res ConcreteMeasurement
	err := json.Unmarshal([]byte(str), &res)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDeserializeTemperature(t *testing.T) {
	str := `"0.000 "`
	var temp Temperature
	err := json.Unmarshal([]byte(str), &temp)
	if err != nil {
		t.Fatal(err)
	}
}