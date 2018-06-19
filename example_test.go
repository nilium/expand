package expand_test

import (
	"fmt"

	"go.spiff.io/expand"
)

// The following is a simulated set of environment variables. To use your OS's environment
// variables, you'd use something like os.LookupEnv or equivalent, depending on what you do
// and do not allow expansion of.
var env = map[string]string{
	"NAME": "Example",
}

// testGetenv returns a value for the given name in env if it has a defined value.
func testGetenv(name string) (string, bool) {
	v, ok := env[name]
	return v, ok
}

func ExampleExpand_default() {
	fmt.Println(expand.Expand("Hello, ${NAME:-World}. ${NAME:+I was given your name.}${NAME:/I was not given your name.}", testGetenv))
	fmt.Println(expand.Expand("This is in ${ENVIRON:-testing}. ${ENVIRON:+This environment was given.}${ENVIRON:/This is the default environment.}", testGetenv))

	// Output:
	// Hello, Example. I was given your name.
	// This is in testing. This is the default environment.
}

func ExampleExpand_custom() {
	// parser is used to expand %(KEY) expressions in strings.
	parser := expand.Parser{
		ScopedOnly: true,
		LeadByte:   '%',
		OpenByte:   '(',
		CloseByte:  ')',
	}

	fmt.Println(parser.Expand("Hello, %(NAME:-World). %(NAME:+I was given your name.)%(NAME:/I was not given your name.)", testGetenv))
	fmt.Println(parser.Expand("This is in %(ENVIRON:-testing). %(ENVIRON:+This environment was given.)%(ENVIRON:/This is the default environment.)", testGetenv))

	// Output:
	// Hello, Example. I was given your name.
	// This is in testing. This is the default environment.
}
