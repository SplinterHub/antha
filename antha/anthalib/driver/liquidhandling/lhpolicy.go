// /anthalib/driver/liquidhandling/lhpolicy.go: Part of the Antha language
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
	"io/ioutil"
)

const (
	LHP_AND int = iota
	LHP_OR
)

func LoadLHPoliciesFrom(filename string) *LHPolicyRuleSet {
	dat, _ := ioutil.ReadFile(filename)

	var lhprs LHPolicyRuleSet

	json.Unmarshal(dat, &lhprs)

	return &lhprs
}

// this structure defines parameters
type LHPolicy map[string]interface{}

// clobber everything in here with the other policy
// then return the merged copy
func (lhp LHPolicy) MergeWith(other LHPolicy) LHPolicy {
	for k, v := range other {
		lhp[k] = v
	}
	return lhp
}

// conditions are ANDed together
// there is no chaining
type LHPolicyRule struct {
	Name       string
	Conditions []LHVariableCondition
	Priority   int // possibly deprecated
	Type       int // AND =0 OR = 1
}

func NewLHPolicyRule(name string) LHPolicyRule {
	var lhpr LHPolicyRule
	lhpr.Name = name
	lhpr.Conditions = make([]LHVariableCondition, 0, 5)
	return lhpr
}

func (lhpr *LHPolicyRule) AddNumericConditionOn(variable string, up, low float64) {
	lhvc := NewLHVariableCondition(variable)
	lhvc.SetNumeric(up, low)
	lhpr.Conditions = append(lhpr.Conditions, lhvc)
}

func (lhpr *LHPolicyRule) AddCategoryConditionOn(variable, category string) {
	lhvc := NewLHVariableCondition(variable)
	lhvc.SetCategoric(category)
	lhpr.Conditions = append(lhpr.Conditions, lhvc)
}

func (lhpr LHPolicyRule) Check(ins RobotInstruction) bool {
	for _, condition := range lhpr.Conditions {
		if !condition.Check(ins) {
			return false
		}
	}
	return true
}

// this just looks for the same conditions, doesn't matter if
// the rules lead to different outcomes...
// not sure if this quite gives us the right behaviour but let's
// plough on for now
func (lhpr LHPolicyRule) IsEqualTo(other LHPolicyRule) bool {
	// cannot be equal if the number of conditions is not equal
	// well we *could have this situation
	//	A: [a,b] B: [c,d] C: [a,d]
	// where rule 1 has both A and B and rule 2 only C but all
	// three have the same consequences but we'll just have to
	// try and enforce some consistency rules to prevent that situation

	if len(lhpr.Conditions) != len(other.Conditions) {
		return false
	}

	// now we have to go through - these are not ordered so there's
	// no general way to find out if the two sets are identical

	for _, c := range lhpr.Conditions {
		if !other.HasCondition(c) {
			return false
		}
	}
	return true
}

func (lhpr LHPolicyRule) HasCondition(cond LHVariableCondition) bool {
	for _, c := range lhpr.Conditions {
		if c.IsEqualTo(cond) {
			return true
		}
	}
	return false
}

type LHVariableCondition struct {
	TestVariable string
	Condition    LHCondition
}

func NewLHVariableCondition(testvariable string) LHVariableCondition {
	var lhvc LHVariableCondition
	lhvc.TestVariable = testvariable
	return lhvc
}

func (lhvc *LHVariableCondition) SetNumeric(up, low float64) {
	if up > low {
		panic("Nonsensical numeric condition requested")
	}
	lhvc.Condition = LHNumericCondition{up, low}
}

func (lhvc *LHVariableCondition) SetCategoric(category string) {
	if category == "" {
		panic("No empty categoric conditions can be made")
	}
	lhvc.Condition = LHCategoryCondition{category}
}

func (lhvc LHVariableCondition) IsEqualTo(other LHVariableCondition) bool {
	if lhvc.TestVariable != other.TestVariable {
		return false
	}
	return lhvc.Condition.IsEqualTo(other.Condition)
}

func (lhvc LHVariableCondition) Check(ins RobotInstruction) bool {
	v := ins.GetParameter(lhvc.TestVariable)
	return lhvc.Condition.Match(v)
}

type LHPolicyRuleSet struct {
	Policies map[string]LHPolicy
	Rules    map[string]LHPolicyRule
}

func NewLHPolicyRuleSet() *LHPolicyRuleSet {
	var lhpr LHPolicyRuleSet
	lhpr.Policies = make(map[string]LHPolicy)
	lhpr.Rules = make(map[string]LHPolicyRule)
	return &lhpr
}

func (lhpr *LHPolicyRuleSet) AddRule(rule LHPolicyRule, consequent LHPolicy) {
	lhpr.Policies[rule.Name] = consequent
	lhpr.Rules[rule.Name] = rule
}

func CloneLHPolicyRuleSet(parent *LHPolicyRuleSet) *LHPolicyRuleSet {
	child := NewLHPolicyRuleSet()
	for k, _ := range parent.Rules {
		child.Policies[k] = parent.Policies[k]
		child.Rules[k] = parent.Rules[k]
	}
	return child
}

func (lhpr LHPolicyRuleSet) GetEquivalentRuleTo(rule LHPolicyRule) string {
	for k, c := range lhpr.Rules {
		if c.IsEqualTo(rule) {
			return k
		}
	}

	return ""
}

func (lhpr *LHPolicyRuleSet) MergeWith(other *LHPolicyRuleSet) {
	for k, rule := range other.Rules {
		name := lhpr.GetEquivalentRuleTo(rule)

		if name != "" {
			// merge the two policies
			pol := other.Policies[k]
			p2 := lhpr.Policies[k]
			p2.MergeWith(pol)
			lhpr.Policies[k] = p2
		}
	}
}

func (lhpr LHPolicyRuleSet) GetPolicyFor(ins RobotInstruction) LHPolicy {
	var match LHPolicyRule
	matchpriority := -1

	for _, rule := range lhpr.Rules {
		if rule.Priority >= matchpriority {
			// TODO:
			// parameters of instructions are typically vectors
			// these have to be dealt with properly
			if rule.Check(ins) {
				if rule.Priority == matchpriority {
					// keep the most specific rule (most conditions)
					if len(rule.Conditions) < len(match.Conditions) {
						match = rule
					} else if len(rule.Conditions) == len(match.Conditions) {
						// what now?
						// we need to forbid this situation: two rules at the same
						// level with the same priority can't match the same thing
						// so we do nothing
					}
				} else {
					match = rule
				}
			}
		}
	}

	// we might prefer to just merge this in

	if match.Name == "" {
		return lhpr.Policies["default"]
	}
	return lhpr.Policies["default"].MergeWith(lhpr.Policies[match.Name])
}

type LHCondition interface {
	Match(interface{}) bool
	Type() string
	IsEqualTo(LHCondition) bool
}

type LHCategoryCondition struct {
	Category string
}

func (lhcc LHCategoryCondition) Match(v interface{}) bool {

	switch v.(type) {
	case string:
		s := v.(string)
		if s == lhcc.Category {
			return true
		}
	case []string:
		// true iff all members of the array are the same category
		for _, s := range v.([]string) {
			if !lhcc.Match(s) {
				return false
			}
		}
		return true
	}
	return false
}

func (lhcc LHCategoryCondition) Type() string {
	return "category"
}

func (lhcc LHCategoryCondition) IsEqualTo(other LHCondition) bool {
	if other.Type() != lhcc.Type() {
		return false
	}
	return other.Match(lhcc.Category)
}

type LHNumericCondition struct {
	Upper float64
	Lower float64
}

func (lhnc LHNumericCondition) Type() string {
	return "Numeric"
}

func (lhnc LHNumericCondition) IsEqualTo(other LHCondition) bool {
	if other.Type() != lhnc.Type() {
		return false
	}
	if other.(LHNumericCondition).Upper == lhnc.Upper && other.(LHNumericCondition).Lower == lhnc.Lower {
		return true
	}
	return false
}

func (lhnc LHNumericCondition) Match(v interface{}) bool {
	switch v.(type) {
	case float64:
		f := v.(float64)

		if f <= lhnc.Upper && f >= lhnc.Lower {
			return true
		}
	case []float64:
		//true iff all values are within range
		// these are simple rules but could need refinement
		for _, f := range v.([]float64) {
			if !lhnc.Match(f) {
				return false
			}
		}
		return true
	} // switch
	return false
}
