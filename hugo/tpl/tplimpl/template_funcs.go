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

package tplimpl

import (
  "html/template"

  "github.com/erquhart/netlify-cms-template-parser-go/hugo/tpl/internal"

  // Init the namespaces
  _ "github.com/erquhart/netlify-cms-template-parser-go/hugo/tpl/collections"
  _ "github.com/erquhart/netlify-cms-template-parser-go/hugo/tpl/urls"
)

func (t *templateFuncster) initFuncMap() {
  funcMap := template.FuncMap{}

  // Merge the namespace funcs
  for _, nsf := range internal.TemplateFuncsNamespaceRegistry {
    ns := nsf(t.Deps)
    if _, exists := funcMap[ns.Name]; exists {
      panic(ns.Name + " is a duplicate template func")
    }
    funcMap[ns.Name] = ns.Context
    for _, mm := range ns.MethodMappings {
      for _, alias := range mm.Aliases {
        if _, exists := funcMap[alias]; exists {
          panic(alias + " is a duplicate template func")
        }
        funcMap[alias] = mm.Method
      }

    }
  }

  t.funcMap = funcMap
  t.Tmpl.(*templateHandler).setFuncs(funcMap)
}
