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
  "net/url"
)

// URLize is similar to MakePath, but with Unicode handling
// Example:
//     uri: Vim (text editor)
//     urlize: vim-text-editor
func URLize(uri string) string {
  return URLEscape(MakePathSanitized(uri))
}

// URLEscape escapes unicode letters.
func URLEscape(uri string) string {
  // escape unicode letters
  parsedURI, err := url.Parse(uri)
  if err != nil {
    // if net/url can not parse URL it means Sanitize works incorrectly
    panic(err)
  }
  x := parsedURI.String()
  return x
}

