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

// Package gomoddirectivecomments provides a parser for Go module directive comments
// that specify "policies" for module dependencies.
package gomoddirectivecomments

import (
	"fmt"
	"log/slog"
	"strings"

	"golang.org/x/mod/modfile"
)

// Parse parses Go module directive comments in the given modfile.File.
//
// The comments are expected to be in the form of
//
//	// <namespace>:<policy>
//
// For example, with namespace "gomodjail", a comment like
//
//	// gomodjail:confined
//
// specifies the "confined" policy.
//
// The namespace argument specifies the namespace to look for in comments.
// The nilPolicy argument specifies the default policy to apply when no
// per-module policy is specified.
// In the case of gomodjail, the nilPolicy is typically "unconfined".
//
// The returned map maps module paths to their policies.
// Modules with no specified policy and the nilPolicy are omitted from the map.
// An error is returned if any comment cannot be parsed.
func Parse(mod *modfile.File, namespace, nilPolicy string) (map[string]string, error) {
	res := make(map[string]string)
	currentDefaultPolicy := nilPolicy

	for _, c := range append(mod.Module.Syntax.Before, mod.Module.Syntax.Suffix...) {
		if tok := c.Token; tok != "" {
			pol, err := policyFromComment(tok, namespace)
			if err != nil {
				err = fmt.Errorf("failed to parse comment %+v: %w", c, err)
				return nil, err
			}
			currentDefaultPolicy = pol
		}
	}

	for _, c := range append(mod.Go.Syntax.Before, mod.Go.Syntax.Suffix...) {
		if tok := c.Token; tok != "" {
			pol, err := policyFromComment(tok, namespace)
			if err != nil {
				err = fmt.Errorf("failed to parse comment %+v: %w", c, err)
				return nil, err
			}
			return nil, fmt.Errorf("policy %q is specified in an invalid position", pol)
		}
	}

	for _, f := range mod.Require {
		if syn := f.Syntax; syn != nil {
			pol := currentDefaultPolicy
			if syn.InBlock {
				// TODO: cache line blocks
				if lineBlock := findLineBlock(mod.Syntax.Stmt, syn); lineBlock != nil {
					lineBlockPol, err := policyFromLineBlock(lineBlock, namespace)
					if err != nil {
						err = fmt.Errorf("failed to parse line block %+v: %w", lineBlock, err)
						return nil, err
					}
					if lineBlockPol != "" {
						pol = lineBlockPol
					}
				}
			}
			for _, c := range append(syn.Before, syn.Suffix...) {
				if tok := c.Token; tok != "" {
					polFromComment, err := policyFromComment(tok, namespace)
					if err != nil {
						err = fmt.Errorf("failed to parse comment %+v: %w", c, err)
						return nil, err
					}
					if polFromComment != "" {
						pol = polFromComment
					}
				}
			}
			if pol == "" {
				pol = currentDefaultPolicy
			}
			if pol == nilPolicy {
				pol = "" // reduce map size
			}
			if existPol, ok := res[f.Mod.Path]; ok && existPol != pol {
				slog.Warn("Overwriting an existing policy", "module", f.Mod.Path, "old", existPol, "new", pol)
			}
			if pol == "" {
				delete(res, f.Mod.Path)
			} else {
				res[f.Mod.Path] = pol
			}
		}
	}
	return res, nil
}

func policyFromComment(token, namespace string) (string, error) {
	token = strings.TrimPrefix(token, "//")
	// TODO: support /* ... */
	for _, f := range strings.Fields(token) {
		f = strings.TrimPrefix(f, "//")
		if strings.HasPrefix(f, namespace+":") {
			pol := strings.TrimPrefix(f, namespace+":")
			return pol, nil
		}
	}
	return "", nil
}

func findLineBlock(exprs []modfile.Expr, line modfile.Expr) *modfile.LineBlock {
	start, end := line.Span()
	for _, expr := range exprs {
		lb, ok := expr.(*modfile.LineBlock)
		if !ok {
			continue
		}
		lbStart, lbEnd := lb.Span()
		if start.Line >= lbStart.Line && end.Line <= lbEnd.Line {
			return lb
		}
	}
	return nil
}

func policyFromLineBlock(lb *modfile.LineBlock, namespace string) (string, error) {
	for _, c := range append(lb.Before, lb.Suffix...) {
		if tok := c.Token; tok != "" {
			pol, err := policyFromComment(tok, namespace)
			if err != nil {
				return "", err
			}
			if pol != "" {
				return pol, nil
			}
		}
	}
	return "", nil
}
