/*
 * Copyright 2018 NEC Corporation
 *
 *   Licensed under the Apache License, Version 2.0 (the "License");
 *   you may not use this file except in compliance with the License.
 *   You may obtain a copy of the License at
 *
 *       http://www.apache.org/licenses/LICENSE-2.0
 *
 *   Unless required by applicable law or agreed to in writing, software
 *   distributed under the License is distributed on an "AS IS" BASIS,
 *   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *   See the License for the specific language governing permissions and
 *   limitations under the License.
 */

package threshold

import (
	"github.com/distributed-monitoring/policy-engine-sandbox/policyexpr"
	"strconv"
)

func compareFalse(_ []float64, _ float64) bool {
	return false
}

func compareEq(list []float64, val float64) bool {
	return false
}

func compareNe(list []float64, val float64) bool {
	return false
}

func compareLe(list []float64, val float64) bool {
	for _, el := range list {
		if el <= val {
			return true
		}
	}
	return false
}

func compareGe(list []float64, val float64) bool {
	for _, el := range list {
		if el >= val {
			return true
		}
	}
	return false
}

func compareLt(list []float64, val float64) bool {
	for _, el := range list {
		if el < val {
			return true
		}
	}
	return false
}

func compareGt(list []float64, val float64) bool {
	for _, el := range list {
		if el > val {
			return true
		}
	}
	return false
}

func Evaluate(p *policyexpr.Parser, rdlist []rawData) []ResourceLabel {
	rllist := []ResourceLabel{}

	value, _ := strconv.ParseFloat(p.Right.ExprNum, 64)

	for _, rd := range rdlist {
		compareFunc := compareFalse
		switch p.Ops {
		case policyexpr.ExprEq:
			compareFunc = compareEq
		case policyexpr.ExprNe:
			compareFunc = compareNe
		case policyexpr.ExprLe:
			compareFunc = compareLe
		case policyexpr.ExprGe:
			compareFunc = compareGe
		case policyexpr.ExprLt:
			compareFunc = compareLt
		case policyexpr.ExprGt:
			compareFunc = compareGt
		}
		if compareFunc(rd.datalist, value) {
			rllist = append(rllist, rd.key)
		}
	}
	return rllist
}
