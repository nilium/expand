expand
===
[![GoDoc](https://godoc.org/go.spiff.io/expand?status.svg)](https://godoc.org/go.spiff.io/expand)
[![Build Status](https://travis-ci.org/nilium/expand.svg?branch=master)](https://travis-ci.org/nilium/expand)
[![codecov](https://codecov.io/gh/nilium/expand/branch/master/graph/badge.svg)](https://codecov.io/gh/nilium/expand)

	$ go get -u go.spiff.io/expand

The expand package is used to handle basic `${VAR:-DEFAULT}` style text
interpolations.

It currently only supports ASCII variable names and customizable leaders and
open/closing brackets (in ASCII), since you probably won't see variations on
that in the wild.

Accepted expansions are:

| Syntax         | Kind
| -------------- | ----
| `$VAR`         | Expand the name VAR (allowed if `Parser.ScopedOnly=false`).
| `${VAR}`       | Expand the name VAR.
| `${VAR:-DEF}`  | Expand the name VAR. If VAR is undefined, expand to DEF.
| `${VAR:+DEF}`  | If VAR is defined, expand to DEF, otherwise nothing.
| `${VAR:/DEF}`  | If VAR is undefined, expand to DEF, otherwise nothing.

The '$', '{', and '}' in the above expansions are configurable. If a backslash
or '$' is encountered before a '$' (or other leader) above, it is expanded to
a literal '$' instead of a variable expansion.

This differs from [os.Expand](https://godoc.org/os#Expand) in that it allows the
use of shell expansions like `${V:-}`, `${V:+}`, and a `${V:/}` (not found in
any shells I know, but useful for when you want to include some text only if
V was defined).

Further expansions can be nested inside of the default/defined/undefined
operators' values.


Example
---
The following example demonstrates a simple use of the package to expand
environment variables in a string, using the syntax `%(NAME)` (while blocking
`%NAME` as a form):

```go
package main

import (
	"fmt"
	"os"

	"go.spiff.io/expand"
)

func main() {
	// parser is used to expand %(KEY) expressions in strings.
	parser := expand.Parser{
		ScopedOnly: true,
		LeadByte:   '%',
		OpenByte:   '(',
		CloseByte:  ')',
	}

	fmt.Println(parser.Expand("Hello, %(NAME:-World).", os.LookupEnv))
	fmt.Println(parser.Expand("This is in %(ENVIRON:-testing).", os.LookupEnv))
}
```

If you want normal `${V}` expansions, you can use the simpler `expand.Expand`
function:

```go
package main

import (
	"fmt"
	"os"

	"go.spiff.io/expand"
)

func main() {
	fmt.Println(parser.Expand("Hello, ${NAME:-World}.", os.LookupEnv))
	fmt.Println(parser.Expand("This is in ${ENVIRON:-testing}.", os.LookupEnv))
}
```


License
---
expand is licensed under the BSD 2-clause license.
It can be read in the LICENSE file.
