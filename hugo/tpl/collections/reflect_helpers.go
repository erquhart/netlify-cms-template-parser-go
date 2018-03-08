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
  "errors"
  "reflect"
  "time"
)

var (
  zero reflect.Value
  errorType = reflect.TypeOf((*error)(nil)).Elem()
  timeType  = reflect.TypeOf((*time.Time)(nil)).Elem()
)

func numberToFloat(v reflect.Value) (float64, error) {
  switch kind := v.Kind(); {
  case isFloat(kind):
    return v.Float(), nil
  case isInt(kind):
    return float64(v.Int()), nil
  case isUint(kind):
    return float64(v.Uint()), nil
  case kind == reflect.Interface:
    return numberToFloat(v.Elem())
  default:
    return 0, errors.New("Invalid kind in numberToFloat")
  }
}

func isNumber(kind reflect.Kind) bool {
  return isInt(kind) || isUint(kind) || isFloat(kind)
}

func isInt(kind reflect.Kind) bool {
  switch kind {
  case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
    return true
  default:
    return false
  }
}

func isUint(kind reflect.Kind) bool {
  switch kind {
  case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
    return true
  default:
    return false
  }
}

func isFloat(kind reflect.Kind) bool {
  switch kind {
  case reflect.Float32, reflect.Float64:
    return true
  default:
    return false
  }
}
