// wunit/siprefix.go: Part of the Antha language
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
// contact license@antha-lang.org or write to the Antha team c/o 
// Synthace Ltd. The London Bioscience Innovation Centre
// 1 Royal College St, London NW1 0NH UK

package wunit

import(
	"math"
)


func RoundInt(v float64)int{
	f:=1.0
	if (v<0){
		f=-1.0
	}

	return int(f*(0.5+(f*v)))
}

// prefix library
var prefices map [string]SIPrefix
// maps log(prefix value) back to a symbol e.g. 2: c
var seciferp map [int] string

// structure defining an SI prefix
type SIPrefix struct{
	// prefix name
	Name string
	// meaning in base 10
	Value float64
}

// helper function to allow lookup of prefix
func SIPrefixBySymbol(symbol string) SIPrefix{
	if prefices==nil{
		prefices=MakePrefices()
	}
	// sugar to allow using empty prefix
	if symbol==""{
		symbol=" "
	}

	return prefices[symbol]
}

// helper function for reverse lookup of prefix
func ReverseLookupPrefix(i int) string{
	if (seciferp == nil){
		seciferp=make(map[int]string, 26)
		for k,v := range prefices{
			lg:=RoundInt(math.Log10(v.Value))
			seciferp[lg]=k
		}
	}
	return seciferp[i]
}


// multiply two prefix values
// take care: there are no checks for going out of bounds
// e.g. Z*Z will generate an error! 
func PrefixMul(x string, y string) string{
	//multiply x by y, what do you get?

	l1:=RoundInt(math.Log10(prefices[x].Value))
	l2:=RoundInt(math.Log10(prefices[y].Value))

	return ReverseLookupPrefix(l1+l2)
}
// divide one prefix by another
// take care: there are no checks for going out of bounds
// e.g. Z/z will give an error!
func PrefixDiv(x string, y string) string{
	// divide x by y, what do you get?

	l1:=RoundInt(math.Log10(prefices[x].Value))
	l2:=RoundInt(math.Log10(prefices[y].Value))
	return ReverseLookupPrefix(l1-l2)
}

// make the prefix structure
func MakePrefices() map[string]SIPrefix{
	pref_map := make(map [string]SIPrefix, 20)
	exponent:=-24
	pfcs:="yzafpnum"

	for _,rune := range(pfcs){
		prefix:=SIPrefix{string(rune), math.Pow10(exponent)}
	//	fmt.Println(prefix)
		pref_map[string(rune)]=prefix
		exponent+=3
	}

	pfcs="cd h"

	exponent=-2

	for _,rune := range(pfcs){
		prefix:=SIPrefix{string(rune), math.Pow10(exponent)}
		pref_map[string(rune)]=prefix
		exponent+=1
	}

	exponent=3

	pfcs="kMGTPEZY"

	for _,rune := range(pfcs){
		prefix:=SIPrefix{string(rune), math.Pow10(exponent)}
	//	fmt.Println(prefix)
		pref_map[string(rune)]=prefix
		exponent+=3
	}

	return pref_map
}

