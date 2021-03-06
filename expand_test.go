package expand // import "go.spiff.io/expand"

import (
	"runtime"
	"strconv"
	"testing"
)

func TestExpansions(t *testing.T) {
	lookup := func(key string) (value string, ok bool) {
		switch key {
		case "empty":
			return "", true
		case "none":
			return "", false
		default:
			return "text", true
		}
		return "text", key != "none"
	}

	type testCase struct {
		Name string
		In   string
		Want string
	}

	cases := []testCase{
		{
			"Empty",
			"",
			"",
		},
		{
			"AllDollarSigns",
			"$$$$$$$$$",
			"$$$$$",
		},
		{
			"RequireLeadByte",
			"#{key}${key}",
			"#{key}text",
		},
		{
			"Literal",
			"Literal",
			"Literal",
		},
		{
			"Delimiter",
			"${}{}",
			"${}{}",
		},
		{
			"EscapeDelimiterBraced",
			"$${key}",
			"${key}",
		},
		{
			"ComplexTrailingLiteral",
			"${key}literal",
			"textliteral",
		},
		{
			"ComplexLeadingLiteral",
			"literal${key}",
			"literaltext",
		},
		{
			"ComplexPaddedLiteral",
			"literal${key}literal",
			"literaltextliteral",
		},
		{
			"ComplexPaddedLiteralMulti",
			"${key:-foo}literal${key}literal${key:+bar}${key:/baz}${none:/qux}",
			"textliteraltextliteralbarqux",
		},
		{
			"EscapeDelimiterBracedUndefined",
			"$${none:-undefined}${none:-undefined}",
			"${none:-undefined}undefined",
		},
		{
			"EscapeDelimiterBracedDefined",
			"$${key:+defined}${key:+defined}",
			"${key:+defined}defined",
		},
		{
			"ExpandUndefinedOnly",
			"${none-not a variable}",
			"not a variable",
		},
		{
			"ExpandDefinedOnly",
			"${key+x} ${empty+empty var} ${none+y}",
			"x empty var ",
		},
		{
			"NestedInterpolations",
			"$$${key:+defined: ${key} ${none:-undefined}}$$",
			"$defined: text undefined$",
		},
		{
			"NestedInterpolations",
			"$$${key:-undefined: ${key} ${none:-undefined}}$$",
			"$text$",
		},
		{
			"ExpansionTypes",
			"literal $key $none ${key} ${none} ${key:-undefined} ${none:-undefined} ${key:+defined} ${none:+defined}",
			"literal text  text  text undefined defined ",
		},
		{
			"IncompleteExpansionNone",
			"${key}${key:${none:-alice}",
			"text${key:alice",
		},
		{
			"IncompleteExpansionUndefined",
			"${key}${key:-${none:-bob}",
			"text${key:-bob",
		},
		{
			"IncompleteExpansionDefined",
			"${key}${key:+${none:-carol}",
			"text${key:+carol",
		},
		{
			"OnlyVariable",
			"$variable",
			"text",
		},
		{
			"NoVariable",
			"${}${ }${:}${:+}${:-}$ ",
			"${}${ }${:}${:+}${:-}$ ",
		},
		{
			"AllPrefixes",
			"Foo${ }Bar",
			"Foo${ }Bar",
		},
		{
			"UnusualCharacters",
			"${!@#%^&*()[]{.,/<>?;'\"|-+=~`é}}",
			"text}",
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.Name, func(t *testing.T) {
			if got := Expand(c.In, lookup); got != c.Want {
				t.Fatalf("Expand(%q) = %q; want %q", c.In, got, c.Want)
			}
		})
	}
}

func TestScopedOnlyExpansions(t *testing.T) {
	lookup := func(key string) (value string, ok bool) {
		return "text", key != "none"
	}

	type testCase struct {
		Name string
		In   string
		Want string
	}

	cases := []testCase{
		{
			"Empty",
			"",
			"",
		},
		{
			"RequireLeadByte",
			"#(key)$(key)",
			"text$(key)",
		},
		{
			"AllHashSigns",
			"#########",
			"#########",
		},
		{
			"Literal",
			"Literal",
			"Literal",
		},
		{
			"Delimiter",
			"#()()",
			"#()()",
		},
		{
			"EscapeDelimiterBraced",
			"\\#(key)",
			"#(key)",
		},
		{
			"ComplexTrailingLiteral",
			"#(key)literal",
			"textliteral",
		},
		{
			"ComplexLeadingLiteral",
			"literal#(key)",
			"literaltext",
		},
		{
			"ComplexPaddedLiteral",
			"literal#(key)literal",
			"literaltextliteral",
		},
		{
			"ComplexPaddedLiteralMulti",
			"#(key:-foo)literal#(key)literal#(key:+bar)#(key:/baz)#(none:/qux)",
			"textliteraltextliteralbarqux",
		},
		{
			"EscapeDelimiterBracedUndefined",
			"\\f#(none:-undefined)#(none:-undefined)",
			"\\fundefinedundefined",
		},
		{
			"EscapeDelimiterBracedUndefined",
			"\\#(none:-undefined)#(none:-undefined)",
			"#(none:-undefined)undefined",
		},
		{
			"EscapeDelimiterBracedDefined",
			"\\#(key:+defined)#(key:+defined)",
			"#(key:+defined)defined",
		},
		{
			"EscapedNestedInterpolations",
			"#\\#(key:+defined: #(key) #(none:-undefined))\\\\#(key:+ defined)##",
			"##(key:+defined: text undefined)\\ defined##",
		},
		{
			"NestedInterpolations_1",
			"##(key:+defined: #(key) #(none:-undefined))##",
			"#defined: text undefined##",
		},
		{
			"NestedInterpolations_2",
			"###(key:-undefined: #(key) #(none:-undefined))##",
			"##text##",
		},
		{
			"ExpansionTypes",
			"literal #key #none #(key) #(none) #(key:-undefined) #(none:-undefined) #(key:+defined) #(none:+defined)",
			"literal #key #none text  text undefined defined ",
		},
		{
			"IncompleteExpansionNone",
			"#(key)#(key:#(none:-alice)",
			"text#(key:alice",
		},
		{
			"IncompleteExpansionUndefined",
			"#(key)#(key:-#(none:-bob)",
			"text#(key:-bob",
		},
		{
			"IncompleteExpansionDefined",
			"#(key)#(key:+#(none:-carol)",
			"text#(key:+carol",
		},
		{
			"OnlyVariable",
			"#variable",
			"#variable",
		},
		{
			"NoVariable",
			"#()#( )#(:)#(:+)#(:-)# ",
			"#()#( )#(:)#(:+)#(:-)# ",
		},
		{
			"AllPrefixes",
			"Foo#( )Bar",
			"Foo#( )Bar",
		},
		{
			"UnusualCharacters",
			"#(!@$%^&*{}[](.,/<>?;'\"|-+=~`é))",
			"text)",
		},
	}

	p := Parser{
		LeadByte:   '#',
		OpenByte:   '(',
		CloseByte:  ')',
		ScopedOnly: true,
	}
	for _, c := range cases {
		c := c
		t.Run(c.Name, func(t *testing.T) {
			if got := p.Expand(c.In, lookup); got != c.Want {
				t.Fatalf("p.Expand(%q, ...) = %q; want %q", c.In, got, c.Want)
			}
		})
	}
}

func TestUnicodeExpand(t *testing.T) {
	const want = "你好"
	const wantKey = "မင်္ဂလာပါ"
	const input = "${" + wantKey + "}"
	gotKey := false

	fn := func(key string) (string, bool) {
		if gotKey = wantKey == key; gotKey {
			return want, true
		}
		return "", false
	}

	got := Expand(input, fn)
	if got != want {
		t.Fatalf("Expand(%q) = %q; want %q", input, got, want)
	}
}

func TestGetFunc(t *testing.T) {
	want, wantOK := "value", true
	got, gotOK := GetFunc(func(string) string { return "value" })("key")
	if got != want || gotOK != wantOK {
		t.Errorf("GetFunc(-> value) = %q, %t; want %q, %t", got, gotOK, want, wantOK)
	}

	want, wantOK = "", false
	got, gotOK = GetFunc(func(string) string { return "" })("key")
	if got != want || gotOK != wantOK {
		t.Errorf("GetFunc(-> value) = %q, %t; want %q, %t", got, gotOK, want, wantOK)
	}
}

func TestCrashExpansion(t *testing.T) {
	// The following cases were generated with go-fuzz and produce panics.
	// In all cases, the following strings should expand to themselves.
	cases := []string{
		"${0:",
		"${0:+${0:",
		"${0:+${0:+${0:+${0:+",
		"${0:+${0:+${0:+${0:+${0:+",
		"${0:-${0:-${0:-",
		"${0:-${0:-${0:-${0:",
		"${0:-${0:-${0:-${0:-${0:",
		"${0:-${0:-${0:-${0:-${0:-${0:",
		"${0:-${0:-${0:-${0:-${0:-${0:-${0:-${0:-${0:-",
		"${0:-${0:-${0:-${0:-${0:-${0:-${0:-${0:-${0:-${0:",
		"${0:-${0:-${0:-${0:-${0:-${0:-${0:-${0:-${0:-${0:-",
		"${0:-${0:-${0:-${0:-${0:-${0:-${0:-${0:-${0:-${0:-${0:",
		"${0:-${0:-${0:-${0:-${0:-${0:-${0:-${0:-${0:-${0:-${0:-${0:",
		"${0:-${0:-${0:/${0:-${0:-${0:-",
		"${0:-${0:-${0:/${0:-${0:-${0:-${0:",
		"${0:-${0:-${0:/${0:-${0:-${0:-${0:-",
		"${0:-${0:-${0:/${0:-${0:-${0:-${0:/${0:",
		"${0:-${0:-${0:/${0:-${0:-${0:-${0:/${0:-",
		"${0:-${0:-${0:/${0:-${0:-${0:-${0:/${0:-${0:",
		"${0:/",
		"${0:/${0:/",
		"${0:/${0:/${0:",
	}

	nop := func(string) (string, bool) { return "", false }

	for i, in := range cases {
		in := in
		t.Run(strconv.Itoa(i+1), func(t *testing.T) {
			defer func() {
				if rc := recover(); rc != nil {
					t.Errorf("panic: %v\n%s", rc, stack())
				}
			}()
			if got := Expand(in, nop); got != in {
				t.Errorf("Expand(%q) = %q; want %q", in, got, in)
			}
		})
	}
}

// stack returns up to 1024 bytes of the calling goroutine's stack trace using runtime.Stack.
func stack() []byte {
	var buf [1024]byte
	return buf[:runtime.Stack(buf[:], false)]
}
