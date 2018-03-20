package main

import (
  "bytes"
  "html/template"

  "github.com/erquhart/netlify-cms-template-parser-go/hugo/helpers"
  "github.com/erquhart/netlify-cms-template-parser-go/hugo/hugolib"
  "github.com/erquhart/netlify-cms-template-parser-go/hugo/tpl/collections"
  "github.com/erquhart/netlify-cms-template-parser-go/hugo/tpl/encoding"
  "github.com/erquhart/netlify-cms-template-parser-go/hugo/tpl/safe"
  "github.com/gopherjs/gopherjs/js"
  "gopkg.in/russross/blackfriday.v2"
)

func main() {
  // For exporting to global/window
  js.Global.Set("goTemplateParser", map[string]interface{}{
    "compile": compile,
    "scratch": hugolib.NewScratch(),
  })
  js.Module.Get("exports").Set("goTemplateParser", map[string]interface{}{
    "compile": compile,
    "scratch": hugolib.NewScratch(),
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
  var t, _ = template.New("").Funcs(template.FuncMap{
    "dict": collections.Dictionary,
    "first": collections.First,
    "jsonify": encoding.Jsonify,
    "markdownify": renderMarkdown,
    "safeJS": safe.JS,
    "slice": collections.Slice,
    "urlize": helpers.URLize,
    "where": collections.Where,
  }).Parse(tmpl)
  t.Execute(&buf, dataMap)
  return buf.String()
}
