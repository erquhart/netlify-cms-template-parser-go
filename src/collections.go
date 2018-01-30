package main

import (
  "errors"
  "reflect"
  "strings"
  "time"

  "github.com/spf13/cast"
)

var (
  zero reflect.Value
  errorType = reflect.TypeOf((*error)(nil)).Elem()
  timeType  = reflect.TypeOf((*time.Time)(nil)).Elem()
)

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


// Where returns a filtered subset of a given data type.
func Where(seq, key interface{}, args ...interface{}) (interface{}, error) {
  seqv, isNil := indirect(reflect.ValueOf(seq))
  if isNil {
    return nil, errors.New("can't iterate over a nil value of type " + reflect.ValueOf(seq).Type().String())
  }

  mv, op, err := parseWhereArgs(args...)
  if err != nil {
    return nil, err
  }

  var path []string
  kv := reflect.ValueOf(key)
  if kv.Kind() == reflect.String {
    path = strings.Split(strings.Trim(kv.String(), "."), ".")
  }

  switch seqv.Kind() {
  case reflect.Array, reflect.Slice:
    return checkWhereArray(seqv, kv, mv, path, op)
  case reflect.Map:
    return checkWhereMap(seqv, kv, mv, path, op)
  default:
    return nil, errors.New("can't iterate over seq")
  }
}

// parseWhereArgs parses the end arguments to the where function.  Return a
// match value and an operator, if one is defined.
func parseWhereArgs(args ...interface{}) (mv reflect.Value, op string, err error) {
  switch len(args) {
  case 1:
    mv = reflect.ValueOf(args[0])
  case 2:
    var ok bool
    if op, ok = args[0].(string); !ok {
      err = errors.New("operator argument must be string type")
      return
    }
    op = strings.TrimSpace(strings.ToLower(op))
    mv = reflect.ValueOf(args[1])
  default:
    err = errors.New("can't evaluate the array by no match argument or more than or equal to two arguments")
  }
  return
}

// checkWhereArray handles the where-matching logic when the seqv value is an
// Array or Slice.
func checkWhereArray(seqv, kv, mv reflect.Value, path []string, op string) (interface{}, error) {
  rv := reflect.MakeSlice(seqv.Type(), 0, 0)
  for i := 0; i < seqv.Len(); i++ {
    var vvv reflect.Value
    rvv := seqv.Index(i)
    if kv.Kind() == reflect.String {
      vvv = rvv
      for _, elemName := range path {
        var err error
        vvv, err = evaluateSubElem(vvv, elemName)
        if err != nil {
          return nil, err
        }
      }
    } else {
      vv, _ := indirect(rvv)
      if vv.Kind() == reflect.Map && kv.Type().AssignableTo(vv.Type().Key()) {
        vvv = vv.MapIndex(kv)
      }
    }

    if ok, err := checkCondition(vvv, mv, op); ok {
      rv = reflect.Append(rv, rvv)
    } else if err != nil {
      return nil, err
    }
  }
  return rv.Interface(), nil
}

// checkWhereMap handles the where-matching logic when the seqv value is a Map.
func checkWhereMap(seqv, kv, mv reflect.Value, path []string, op string) (interface{}, error) {
  rv := reflect.MakeMap(seqv.Type())
  keys := seqv.MapKeys()
  for _, k := range keys {
    elemv := seqv.MapIndex(k)
    switch elemv.Kind() {
    case reflect.Array, reflect.Slice:
      r, err := checkWhereArray(elemv, kv, mv, path, op)
      if err != nil {
        return nil, err
      }

      switch rr := reflect.ValueOf(r); rr.Kind() {
      case reflect.Slice:
        if rr.Len() > 0 {
          rv.SetMapIndex(k, elemv)
        }
      }
    case reflect.Interface:
      elemvv, isNil := indirect(elemv)
      if isNil {
        continue
      }

      switch elemvv.Kind() {
      case reflect.Array, reflect.Slice:
        r, err := checkWhereArray(elemvv, kv, mv, path, op)
        if err != nil {
          return nil, err
        }

        switch rr := reflect.ValueOf(r); rr.Kind() {
        case reflect.Slice:
          if rr.Len() > 0 {
            rv.SetMapIndex(k, elemv)
          }
        }
      }
    }
  }
  return rv.Interface(), nil
}

func evaluateSubElem(obj reflect.Value, elemName string) (reflect.Value, error) {
  if !obj.IsValid() {
    return zero, errors.New("can't evaluate an invalid value")
  }
  obj, isNil := indirect(obj)

  // first, check whether obj has a method. In this case, obj is
  // an interface, a struct or its pointer. If obj is a struct,
  // to check all T and *T method, use obj pointer type Value
  objPtr := obj
  if objPtr.Kind() != reflect.Interface && objPtr.CanAddr() {
    objPtr = objPtr.Addr()
  }
  mt, ok := objPtr.Type().MethodByName(elemName)
  if ok {
    if mt.PkgPath != "" {
      return zero, errors.New(elemName + " is an unexported method of type ")
    }
    // struct pointer has one receiver argument and interface doesn't have an argument
    if mt.Type.NumIn() > 1 || mt.Type.NumOut() == 0 || mt.Type.NumOut() > 2 {
      return zero, errors.New(elemName + " is the right method type, but doesn't satisfy requirements")
    }
    if mt.Type.NumOut() == 1 && mt.Type.Out(0).Implements(errorType) {
      return zero, errors.New(elemName + " is the right method type, but doesn't satisfy requirements")
    }
    if mt.Type.NumOut() == 2 && !mt.Type.Out(1).Implements(errorType) {
      return zero, errors.New(elemName + " is the right method type, but doesn't satisfy requirements")
    }
    res := objPtr.Method(mt.Index).Call([]reflect.Value{})
    if len(res) == 2 && !res[1].IsNil() {
      return zero, errors.New("error at calling a method " + elemName)
    }
    return res[0], nil
  }

  // elemName isn't a method so next start to check whether it is
  // a struct field or a map value. In both cases, it mustn't be
  // a nil value
  if isNil {
    return zero, errors.New("can't evaluate nil pointer by a struct field or map key name " + elemName)
  }
  switch obj.Kind() {
  case reflect.Struct:
    ft, ok := obj.Type().FieldByName(elemName)
    if ok {
      if ft.PkgPath != "" && !ft.Anonymous {
        return zero, errors.New(elemName + " is an unexported struct field of the right type")
      }
      return obj.FieldByIndex(ft.Index), nil
    }
    return zero, errors.New(elemName + " isn't a struct field of the right type")
  case reflect.Map:
    kv := reflect.ValueOf(elemName)
    if kv.Type().AssignableTo(obj.Type().Key()) {
      return obj.MapIndex(kv), nil
    }
    return zero, errors.New(elemName + " isn't a map key of the right type")
  }
  return zero, errors.New(elemName + " is neither a struct field, a method nor a map element of the right type")
}

func checkCondition(v, mv reflect.Value, op string) (bool, error) {
  v, vIsNil := indirect(v)
  if !v.IsValid() {
    vIsNil = true
  }

  mv, mvIsNil := indirect(mv)
  if !mv.IsValid() {
    mvIsNil = true
  }
  if vIsNil || mvIsNil {
    switch op {
    case "", "=", "==", "eq":
      return vIsNil == mvIsNil, nil
    case "!=", "<>", "ne":
      return vIsNil != mvIsNil, nil
    }
    return false, nil
  }

  if v.Kind() == reflect.Bool && mv.Kind() == reflect.Bool {
    switch op {
    case "", "=", "==", "eq":
      return v.Bool() == mv.Bool(), nil
    case "!=", "<>", "ne":
      return v.Bool() != mv.Bool(), nil
    }
    return false, nil
  }

  var ivp, imvp *int64
  var svp, smvp *string
  var slv, slmv interface{}
  var ima []int64
  var sma []string
  if mv.Type() == v.Type() {
    switch v.Kind() {
    case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
      iv := v.Int()
      ivp = &iv
      imv := mv.Int()
      imvp = &imv
    case reflect.String:
      sv := v.String()
      svp = &sv
      smv := mv.String()
      smvp = &smv
    case reflect.Struct:
      switch v.Type() {
      case timeType:
        iv := toTimeUnix(v)
        ivp = &iv
        imv := toTimeUnix(mv)
        imvp = &imv
      }
    case reflect.Array, reflect.Slice:
      slv = v.Interface()
      slmv = mv.Interface()
    }
  } else {
    if mv.Kind() != reflect.Array && mv.Kind() != reflect.Slice {
      return false, nil
    }

    if mv.Len() == 0 {
      return false, nil
    }

    if v.Kind() != reflect.Interface && mv.Type().Elem().Kind() != reflect.Interface && mv.Type().Elem() != v.Type() && v.Kind() != reflect.Array && v.Kind() != reflect.Slice {
      return false, nil
    }
    switch v.Kind() {
    case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
      iv := v.Int()
      ivp = &iv
      for i := 0; i < mv.Len(); i++ {
        if anInt, err := toInt(mv.Index(i)); err == nil {
          ima = append(ima, anInt)
        }
      }
    case reflect.String:
      sv := v.String()
      svp = &sv
      for i := 0; i < mv.Len(); i++ {
        if aString, err := toString(mv.Index(i)); err == nil {
          sma = append(sma, aString)
        }
      }
    case reflect.Struct:
      switch v.Type() {
      case timeType:
        iv := toTimeUnix(v)
        ivp = &iv
        for i := 0; i < mv.Len(); i++ {
          ima = append(ima, toTimeUnix(mv.Index(i)))
        }
      }
    case reflect.Array, reflect.Slice:
      slv = v.Interface()
      slmv = mv.Interface()
    }
  }

  switch op {
  case "", "=", "==", "eq":
    if ivp != nil && imvp != nil {
      return *ivp == *imvp, nil
    } else if svp != nil && smvp != nil {
      return *svp == *smvp, nil
    }
  case "!=", "<>", "ne":
    if ivp != nil && imvp != nil {
      return *ivp != *imvp, nil
    } else if svp != nil && smvp != nil {
      return *svp != *smvp, nil
    }
  case ">=", "ge":
    if ivp != nil && imvp != nil {
      return *ivp >= *imvp, nil
    } else if svp != nil && smvp != nil {
      return *svp >= *smvp, nil
    }
  case ">", "gt":
    if ivp != nil && imvp != nil {
      return *ivp > *imvp, nil
    } else if svp != nil && smvp != nil {
      return *svp > *smvp, nil
    }
  case "<=", "le":
    if ivp != nil && imvp != nil {
      return *ivp <= *imvp, nil
    } else if svp != nil && smvp != nil {
      return *svp <= *smvp, nil
    }
  case "<", "lt":
    if ivp != nil && imvp != nil {
      return *ivp < *imvp, nil
    } else if svp != nil && smvp != nil {
      return *svp < *smvp, nil
    }
  case "in", "not in":
    var r bool
    if ivp != nil && len(ima) > 0 {
      r = In(ima, *ivp)
    } else if svp != nil {
      if len(sma) > 0 {
        r = In(sma, *svp)
      } else if smvp != nil {
        r = In(*smvp, *svp)
      }
    } else {
      return false, nil
    }
    if op == "not in" {
      return !r, nil
    }
    return r, nil
  case "intersect":
    r, err := Intersect(slv, slmv)
    if err != nil {
      return false, err
    }

    if reflect.TypeOf(r).Kind() == reflect.Slice {
      s := reflect.ValueOf(r)

      if s.Len() > 0 {
        return true, nil
      }
      return false, nil
    }
    return false, errors.New("invalid intersect values")
  default:
    return false, errors.New("no such operator")
  }
  return false, nil
}

// toInt returns the int value if possible, -1 if not.
func toInt(v reflect.Value) (int64, error) {
  switch v.Kind() {
  case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
    return v.Int(), nil
  case reflect.Interface:
    return toInt(v.Elem())
  }
  return -1, errors.New("unable to convert value to int")
}

// toString returns the string value if possible, "" if not.
func toString(v reflect.Value) (string, error) {
  switch v.Kind() {
  case reflect.String:
    return v.String(), nil
  case reflect.Interface:
    return toString(v.Elem())
  }
  return "", errors.New("unable to convert value to string")
}

func toTimeUnix(v reflect.Value) int64 {
  if v.Kind() == reflect.Interface {
    return toTimeUnix(v.Elem())
  }
  if v.Type() != timeType {
    panic("coding error: argument must be time.Time type reflect Value")
  }
  return v.MethodByName("Unix").Call([]reflect.Value{})[0].Int()
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
