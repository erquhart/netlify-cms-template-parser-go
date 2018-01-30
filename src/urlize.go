package main

import (
  "strings"
  "unicode"
  "net/url"

  "golang.org/x/text/transform"
  "golang.org/x/text/unicode/norm"
)

// URLize is similar to MakePath, but with Unicode handling
// Example:
//     uri: Vim (text editor)
//     urlize: vim-text-editor
func (p *PathSpec) URLize(uri string) string {
  return p.URLEscape(p.MakePathSanitized(uri))
}

// URLEscape escapes unicode letters.
func (p *PathSpec) URLEscape(uri string) string {
  // escape unicode letters
  parsedURI, err := url.Parse(uri)
  if err != nil {
    // if net/url can not parse URL it means Sanitize works incorrectly
    panic(err)
  }
  x := parsedURI.String()
  return x
}

// MakePathSanitized creates a Unicode-sanitized string, with the spaces replaced
func (p *PathSpec) MakePathSanitized(s string) string {
  if p.disablePathToLower {
    return p.MakePath(s)
  }
  return strings.ToLower(p.MakePath(s))
}

// MakePath takes a string with any characters and replace it
// so the string could be used in a path.
// It does so by creating a Unicode-sanitized string, with the spaces replaced,
// whilst preserving the original casing of the string.
// E.g. Social Media -> Social-Media
func (p *PathSpec) MakePath(s string) string {
  return p.UnicodeSanitize(strings.Replace(strings.TrimSpace(s), " ", "-", -1))
}

// From https://golang.org/src/net/url/url.go
func ishex(c rune) bool {
  switch {
  case '0' <= c && c <= '9':
    return true
  case 'a' <= c && c <= 'f':
    return true
  case 'A' <= c && c <= 'F':
    return true
  }
  return false
}

func isMn(r rune) bool {
  return unicode.Is(unicode.Mn, r) // Mn: nonspacing marks
}

// UnicodeSanitize sanitizes string to be used in Hugo URL's, allowing only
// a predefined set of special Unicode characters.
// If RemovePathAccents configuration flag is enabled, Uniccode accents
// are also removed.
func (p *PathSpec) UnicodeSanitize(s string) string {
  source := []rune(s)
  target := make([]rune, 0, len(source))

  for i, r := range source {
    if r == '%' && i+2 < len(source) && ishex(source[i+1]) && ishex(source[i+2]) {
      target = append(target, r)
    } else if unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsMark(r) || r == '.' || r == '/' || r == '\\' || r == '_' || r == '-' || r == '#' || r == '+' || r == '~' {
      target = append(target, r)
    }
  }

  var result string

  if p.removePathAccents {
    // remove accents - see https://blog.golang.org/normalization
    t := transform.Chain(norm.NFD, transform.RemoveFunc(isMn), norm.NFC)
    result, _, _ = transform.String(t, string(target))
  } else {
    result = string(target)
  }

  return result
}
