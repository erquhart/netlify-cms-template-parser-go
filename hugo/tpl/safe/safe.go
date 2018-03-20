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

package safe

import (
  "html/template"

  "github.com/spf13/cast"
)

// JS returns the given string as a html/template JS content.
func JS(a interface{}) (template.JS, error) {
  s, err := cast.ToStringE(a)
  return template.JS(s), err
}
