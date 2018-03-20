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

package helpers

import (
  "strings"
  "unicode"
)

// MakePath takes a string with any characters and replace it
// so the string could be used in a path.
// It does so by creating a Unicode-sanitized string, with the spaces replaced,
// whilst preserving the original casing of the string.
// E.g. Social Media -> Social-Media
func MakePath(s string) string {
  return UnicodeSanitize(strings.Replace(strings.TrimSpace(s), " ", "-", -1))
}

// MakePathSanitized creates a Unicode-sanitized string, with the spaces replaced
func MakePathSanitized(s string) string {
  return strings.ToLower(MakePath(s))
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

// UnicodeSanitize sanitizes string to be used in Hugo URL's, allowing only
// a predefined set of special Unicode characters.
// If RemovePathAccents configuration flag is enabled, Uniccode accents
// are also removed.
func UnicodeSanitize(s string) string {
  source := []rune(s)
  target := make([]rune, 0, len(source))

  for i, r := range source {
    if r == '%' && i+2 < len(source) && ishex(source[i+1]) && ishex(source[i+2]) {
      target = append(target, r)
    } else if unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsMark(r) || r == '.' || r == '/' || r == '\\' || r == '_' || r == '-' || r == '#' || r == '+' || r == '~' {
      target = append(target, r)
    }
  }

  return string(target)
}
