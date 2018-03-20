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
  "strings"

  "github.com/spf13/cast"
)

// Dictionary creates a map[string]interface{} from the given parameters by
// walking the parameters and treating them as key-value pairs.  The number
// of parameters must be even.
func Dictionary(values ...interface{}) (map[string]interface{}, error) {
  if len(values)%2 != 0 {
    return nil, errors.New("invalid dictionary call")
  }

  dict := make(map[string]interface{}, len(values)/2)

  for i := 0; i < len(values); i += 2 {
    key, ok := values[i].(string)
    if !ok {
      return nil, errors.New("dictionary keys must be strings")
    }
    dict[key] = values[i+1]
  }

  return dict, nil
}

// First returns the first N items in a rangeable list.
func First(limit interface{}, seq interface{}) (interface{}, error) {
  if limit == nil || seq == nil {
    return nil, errors.New("both limit and seq must be provided")
  }

  limitv, err := cast.ToIntE(limit)
  if err != nil {
    return nil, err
  }

  if limitv < 1 {
    return nil, errors.New("can't return negative/empty count of items from sequence")
  }

  seqv := reflect.ValueOf(seq)
  seqv, isNil := indirect(seqv)
  if isNil {
    return nil, errors.New("can't iterate over a nil value")
  }

  switch seqv.Kind() {
  case reflect.Array, reflect.Slice, reflect.String:
    // okay
  default:
    return nil, errors.New("can't iterate over " + reflect.ValueOf(seq).Type().String())
  }

  if limitv > seqv.Len() {
    limitv = seqv.Len()
  }

  return seqv.Slice(0, limitv).Interface(), nil
}

// In returns whether v is in the set l.  l may be an array or slice.
func In(l interface{}, v interface{}) bool {
  if l == nil || v == nil {
    return false
  }

  lv := reflect.ValueOf(l)
  vv := reflect.ValueOf(v)

  switch lv.Kind() {
  case reflect.Array, reflect.Slice:
    for i := 0; i < lv.Len(); i++ {
      lvv := lv.Index(i)
      lvv, isNil := indirect(lvv)
      if isNil {
        continue
      }
      switch lvv.Kind() {
      case reflect.String:
        if vv.Type() == lvv.Type() && vv.String() == lvv.String() {
          return true
        }
      default:
        if isNumber(vv.Kind()) && isNumber(lvv.Kind()) {
          f1, err1 := numberToFloat(vv)
          f2, err2 := numberToFloat(lvv)
          if err1 == nil && err2 == nil && f1 == f2 {
            return true
          }
        }
      }
    }
  case reflect.String:
    if vv.Type() == lv.Type() && strings.Contains(lv.String(), vv.String()) {
      return true
    }
  }
  return false
}

// Intersect returns the common elements in the given sets, l1 and l2.  l1 and
// l2 must be of the same type and may be either arrays or slices.
func Intersect(l1, l2 interface{}) (interface{}, error) {
  if l1 == nil || l2 == nil {
    return make([]interface{}, 0), nil
  }

  var ins *intersector

  l1v := reflect.ValueOf(l1)
  l2v := reflect.ValueOf(l2)

  switch l1v.Kind() {
  case reflect.Array, reflect.Slice:
    ins = &intersector{r: reflect.MakeSlice(l1v.Type(), 0, 0), seen: make(map[interface{}]bool)}
    switch l2v.Kind() {
    case reflect.Array, reflect.Slice:
      for i := 0; i < l1v.Len(); i++ {
        l1vv := l1v.Index(i)
        for j := 0; j < l2v.Len(); j++ {
          l2vv := l2v.Index(j)
          ins.handleValuePair(l1vv, l2vv)
        }
      }
      return ins.r.Interface(), nil
    default:
      return nil, errors.New("can't iterate over " + reflect.ValueOf(l2).Type().String())
    }
  default:
    return nil, errors.New("can't iterate over " + reflect.ValueOf(l1).Type().String())
  }
}

// Slice returns a slice of all passed arguments.
func Slice(args ...interface{}) []interface{} {
  return args
}

type intersector struct {
  r    reflect.Value
  seen map[interface{}]bool
}

func (i *intersector) appendIfNotSeen(v reflect.Value) {

  vi := v.Interface()
  if !i.seen[vi] {
    i.r = reflect.Append(i.r, v)
    i.seen[vi] = true
  }
}

func (i *intersector) handleValuePair(l1vv, l2vv reflect.Value) {
  switch kind := l1vv.Kind(); {
  case kind == reflect.String:
    l2t, err := toString(l2vv)
    if err == nil && l1vv.String() == l2t {
      i.appendIfNotSeen(l1vv)
    }
  case isNumber(kind):
    f1, err1 := numberToFloat(l1vv)
    f2, err2 := numberToFloat(l2vv)
    if err1 == nil && err2 == nil && f1 == f2 {
      i.appendIfNotSeen(l1vv)
    }
  case kind == reflect.Ptr, kind == reflect.Struct:
    if l1vv.Interface() == l2vv.Interface() {
      i.appendIfNotSeen(l1vv)
    }
  case kind == reflect.Interface:
    i.handleValuePair(reflect.ValueOf(l1vv.Interface()), l2vv)
  }
}
