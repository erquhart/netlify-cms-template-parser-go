package main

import (
  "bytes"
  "html/template"

  "github.com/gopherjs/gopherjs/js"
  "gopkg.in/russross/blackfriday.v2"
)

func main() {
  // For exporting to global/window
  js.Global.Set("goTemplateParser", map[string]interface{}{
    "compile": compile,
  })
}

func renderMarkdown(tmpl string) template.HTML {
  input := []byte(tmpl)
  output := blackfriday.Run(input)
  return template.HTML(output)
}

func compile(data *js.Object, tmpl string) string {
  var dataMap = data.Interface()
  var buf bytes.Buffer
  var t, _ = template.New("").Funcs(template.FuncMap{"renderMarkdown": renderMarkdown}).Parse(tmpl)
  t.Execute(&buf, dataMap)
  return buf.String()
}
