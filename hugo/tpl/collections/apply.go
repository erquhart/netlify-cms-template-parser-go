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

package collections

import (
  "reflect"
)

// indirect is taken from 'text/template/exec.go'
func indirect(v reflect.Value) (rv reflect.Value, isNil bool) {
  for ; v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface; v = v.Elem() {
    if v.IsNil() {
      return v, true
    }
    if v.Kind() == reflect.Interface && v.NumMethod() > 0 {
      break
    }
  }
  return v, false
}
