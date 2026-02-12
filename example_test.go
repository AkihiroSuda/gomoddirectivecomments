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
	"fmt"
	"log"

	"golang.org/x/mod/modfile"

	"github.com/AkihiroSuda/gomoddirectivecomments"
)

func Example() {
	const (
		goMod = `module example.com/main

go 1.23

require example.com/dependency v1.2.3 // gomodjail:confined
`
		namespace = "gomodjail"
		nilPolicy = "unconfined"
	)
	mod, err := modfile.Parse("go.mod", []byte(goMod), nil)
	if err != nil {
		log.Fatal(err)
	}
	policies, err := gomoddirectivecomments.Parse(mod, namespace, nilPolicy)
	if err != nil {
		log.Fatal(err)
	}
	for modPath, policy := range policies {
		fmt.Printf("module %q has policy %q\n", modPath, policy)
	}
	// Output:
	// module "example.com/dependency" has policy "confined"
}
