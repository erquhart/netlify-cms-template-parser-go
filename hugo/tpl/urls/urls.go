// Copyright 2017 The Hugo Authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// With modifications by the Netlify CMS Authors.

package urls

import (
  "github.com/erquhart/netlify-cms-template-parser-go/hugo/deps"
  "github.com/spf13/cast"
)

// New returns a new instance of the urls-namespaced template functions.
func New(deps *deps.Deps) *Namespace {
  return &Namespace{
    deps:      deps,
  }
}

// Namespace provides template functions for the "urls" namespace.
type Namespace struct {
  deps      *deps.Deps
}

func (ns *Namespace) URLize(a interface{}) (string, error) {
  s, err := cast.ToStringE(a)
  if err != nil {
    return "", nil
  }
  return ns.deps.PathSpec.URLize(s), nil
}
