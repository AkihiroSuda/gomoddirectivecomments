// Copyright The gomoddirectivecomments Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package gomoddirectivecomments_test

import (
	"testing"

	"golang.org/x/mod/modfile"
	"gotest.tools/v3/assert"

	"github.com/AkihiroSuda/gomoddirectivecomments"
)

func TestParse(t *testing.T) {
	type testCase struct {
		name     string
		goMod    string
		expected map[string]string
	}
	testCases := []testCase{
		{
			name: "basic",
			goMod: `
module example.com/foo

go 1.23

require (
	example.com/mod100 v1.2.3 // gomodjail:confined
	example.com/mod101 v1.2.3
	// gomodjail:confined
	example.com/mod102 v1.2.3
	example.com/mod103 v1.2.3
	example.com/mod104 v1.2.3 //gomodjail:confined
)

require (
	example.com/mod200 v1.2.3 // indirect
	example.com/mod201 v1.2.3 // indirect; gomodjail:confined
	example.com/mod202 v1.2.3 // indirect // gomodjail:confined
	// gomodjail:confined
	example.com/mod203 v1.2.4 // indirect
	example.com/mod204 v1.2.3 // indirect //gomodjail:confined
)
`,
			expected: map[string]string{
				"example.com/mod100": "confined",
				"example.com/mod102": "confined",
				"example.com/mod104": "confined",
				"example.com/mod201": "confined",
				"example.com/mod202": "confined",
				"example.com/mod203": "confined",
				"example.com/mod204": "confined",
			},
		},

		{
			name: "global",
			goMod: `
// gomodjail:confined
module example.com/foo

go 1.23

require (
	example.com/mod100 v1.2.3
	example.com/mod101 v1.2.3 // gomodjail:unconfined
	example.com/mod102 v1.2.3
	// gomodjail:unconfined
	example.com/mod103 v1.2.3
)

require (
	// gomodjail:unconfined
	example.com/mod200 v1.2.3 // indirect
	example.com/mod201 v1.2.3 // indirect
	example.com/mod202 v1.2.3 // indirect
)

//gomodjail:unconfined
require (
	example.com/mod300 v1.2.3
	example.com/mod301 v1.2.3 // gomodjail:confined
	example.com/mod302 v1.2.3
)

// gomodjail:confined
require (
	example.com/mod400 v1.2.3 // indirect
	example.com/mod401 v1.2.3 // indirect // gomodjail:unconfined
	example.com/mod402 v1.2.3 // indirect
)
`,
			expected: map[string]string{
				"example.com/mod100": "confined",
				"example.com/mod102": "confined",
				"example.com/mod201": "confined",
				"example.com/mod202": "confined",
				"example.com/mod301": "confined",
				"example.com/mod400": "confined",
				"example.com/mod402": "confined",
			},
		},

		{
			name: "blockless",
			goMod: `
module example.com/foo

go 1.23

require example.com/mod v1.2.3 // gomodjail:confined
`,
			expected: map[string]string{
				"example.com/mod": "confined",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mod, err := modfile.Parse(tc.name, []byte(tc.goMod), nil)
			assert.NilError(t, err)
			policies, err := gomoddirectivecomments.Parse(mod, "gomodjail", "unconfined")
			assert.NilError(t, err)
			assert.DeepEqual(t, tc.expected, policies)
		})
	}
}
