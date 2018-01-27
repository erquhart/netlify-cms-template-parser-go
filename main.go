package main

import (
  "bytes"
  "fmt"
  "html/template"

  "github.com/gopherjs/gopherjs/js"
)

// Starting point for compiling JS code
func main() {
  js.Global.Set("goTemplateParser", map[string]interface{}{
    "compile": compile,
  })
}

func compile(data *js.Object, tmpl string) string {
  var dataMap = data.Interface()
  var buf bytes.Buffer
  var t, parseErr = template.New("").Parse(tmpl)
  if parseErr != nil {
    fmt.Println("parsing template:", parseErr)
  }
  execErr := t.Execute(&buf, dataMap)
  if execErr != nil {
    fmt.Println("executing template:", execErr)
  }
  return buf.String()
}
