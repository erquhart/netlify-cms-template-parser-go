package main

import (
  "bytes"
  "html/template"

  "github.com/erquhart/netlify-cms-template-parser-go/hugo/config"
  "github.com/erquhart/netlify-cms-template-parser-go/hugo/helpers"
  "github.com/erquhart/netlify-cms-template-parser-go/hugo/tpl/collections"
  "github.com/erquhart/netlify-cms-template-parser-go/hugo/tpl/urls"
  "github.com/gopherjs/gopherjs/js"
  "gopkg.in/russross/blackfriday.v2"
)

func main() {
  // For exporting to global/window
  js.Global.Set("goTemplateParser", map[string]interface{}{
    "compile": compile,
  })
  js.Module.Get("exports").Set("goTemplateParser", map[string]interface{}{
    "compile": compile,
  })
}

/*
type PathSpec struct {
  disablePathToLower bool
  removePathAccents  bool
}

var PathSpec := helpers.NewPathSpec(&config.Provider{
  disablePathToLower: false,
  removePathAccents: false,
})
*/

func renderMarkdown(tmpl string) template.HTML {
  input := []byte(tmpl)
  output := blackfriday.Run(input)
  return template.HTML(output)
}

func compile(data *js.Object, tmpl string) string {
  var dataMap = data.Interface()
  var buf bytes.Buffer
  var t, _ = template.New("").Funcs(template.FuncMap{
    "renderMarkdown": renderMarkdown,
    "urlize": urls.URLize,
    "first": collections.First,
    "where": collections.Where,
  }).Parse(tmpl)
  t.Execute(&buf, dataMap)
  return buf.String()
}
