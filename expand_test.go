package expand // import "go.spiff.io/expand"

import "testing"

func TestExpansions(t *testing.T) {
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
