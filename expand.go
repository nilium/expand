// Package expand is used to handle basic ${VAR:-DEFAULT} style text interpolations.
//
// It currently only supports ASCII variable names and customizable leaders and open/closing
// brackets (in ASCII), since you probably won't see variations on that in the wild.
//
// Accepted expansions are:
//
//      Syntax          Kind
//      $VAR            Expand the name VAR (ScopedOnly=false).
//      ${VAR}          Expand the name VAR.
//      ${VAR:-DEF}     Expand the name VAR. If VAR is undefined, expand to DEF.
//      ${VAR:+DEF}     If VAR is defined, expand to DEF, otherwise nothing.
//      ${VAR:/DEF}     If VAR is undefined, expand to DEF, otherwise nothing.
//
// Again, the '$', '{', and '}' in the above expansions are configurable. If a backslash or '$' is
// encountered before a '$' (or other leader) above, it is expanded to a literal '$' instead of
// a variable expansion.
package expand // import "go.spiff.io/expand"

// TODO(ncower): Rewrite to support unicode runes as open/lead/close?
// TODO(ncower): Support \ (backslash) as a lead character? (disables \ as escape)

import (
	"bytes"
	"strings"
	"unicode"
)

// Default characters for a Parser's LeadByte, OpenByte, and CloseByte.
const (
	DefaultLeadByte  = '$'
	DefaultOpenByte  = '{'
	DefaultCloseByte = '}'
)

// LookupFunc is any function that looks up a variable with the given name and returns its value and
// whether it's defined.
type LookupFunc func(key string) (value string, ok bool)

// GetFunc converts a function that returns a value for a string and returns a LookupFunc for it.
// The LookupFunc returns that a value is not defined if fn returns an empty string.
func GetFunc(fn func(string) string) LookupFunc {
	return func(key string) (string, bool) {
		v := fn(key)
		return v, v != ""
	}
}

// Expand expands an input string using shell-like expansions $VAR, ${VAR}, ${VAR:-if_undefined}, and
// ${VAR:+if_defined} and returns the result of the expansions. This uses default Parser settings.
//
// Cases of "$VAR" and "${VAR}" are replaced with empty strings if VAR is not defined.
// "${VAR:-if_undefined}" replaces VAR with the text "if_undefined" if VAR is not defined.
// "${VAR:+if_defined}" replaces VAR with the text "if_defined" if VAR is defined.
func Expand(in string, fn LookupFunc) string {
	return (*Parser)(nil).Expand(in, fn)
}

// Parser configures expansion options. LeadByte is the leading character needed to begin an
// expansion, defaulting to '$'. OpenByte and CloseByte refer to opening and closing braces, and
// default to '{' and '}', respectively. LeadByte, OpenByte, and CloseByte may be any non-zero ASCII
// (7-bit) character.
//
// ScopedOnly controls whether only expansions wrapped in OpenByte/CloseByte pairs are expanded
// (e.g., $FOO is the literal text "$FOO", while "${FOO}" is an expansion of the name "FOO").
//
// A nil or zero-value Parser is valid for use.
type Parser struct {
	ScopedOnly bool

	LeadByte  byte
	OpenByte  byte
	CloseByte byte
}

// Expand expands an input string using shell-like expansions $VAR, ${VAR},
// ${VAR:-if_undefined}, and ${VAR:+if_defined} and returns the
// result of the expansions.
//
// Cases of "$VAR" and "${VAR}" are replaced with empty strings if VAR is not defined.
// "${VAR:-if_undefined}" replaces VAR with the text "if_undefined" if VAR is not defined.
// "${VAR:+if_defined}" replaces VAR with the text "if_defined" if VAR is defined.
//
// If ScopedOnly is true, $VAR expansions are ignored.
//
// In the above syntax examples, the "${}" characters are set by the Parser's LeadByte, OpenByte,
// and CloseByte fields.
func (p *Parser) Expand(in string, fn LookupFunc) string {
	var buf bytes.Buffer
	buf.Grow(len(in) * 2)
	_, ex := p.parse(in, -1)
	ex.expand(&buf, fn)
	return buf.String()
}

func (p *Parser) isIDChar(c byte) bool {
	return !(c == ':' ||
		c == p.closeChar() ||
		c == p.leadChar() ||
		unicode.IsSpace(rune(c)))
}

func (p *Parser) isScopedOnly() bool {
	return p != nil && p.ScopedOnly
}

func (p *Parser) leadChar() byte {
	if p == nil || p.LeadByte&0x7f == 0 {
		return DefaultLeadByte
	}
	return p.LeadByte
}

func (p *Parser) openChar() byte {
	if p == nil || p.OpenByte&0x7f == 0 {
		return DefaultOpenByte
	}
	return p.OpenByte
}

func (p *Parser) closeChar() byte {
	if p == nil || p.CloseByte&0x7f == 0 {
		return DefaultCloseByte
	}
	return p.CloseByte
}

func (p *Parser) parse(in string, until rune) (n int, root *expansion) {
	var (
		closeChar  = p.closeChar()
		openChar   = p.openChar()
		leadChar   = p.leadChar()
		scopedOnly = p.isScopedOnly()

		i       int
		c       byte
		consume = func(ex *expansion) {
			switch {
			case len(ex.value) == 0 && ex.kind == exPrefix:
			case root == nil && ex.kind == exPrefix:
				root = ex
			case root == nil:
				root = &expansion{children: []*expansion{ex}}
			case ex.kind == exPrefix && len(root.children) == 0:
				root.value += ex.value
			default:
				root.children = append(root.children, ex)
			}
			n += len(in[:i])
			in = in[i:]
			i = 0
		}
	)

reset:
	for ; i < len(in); i++ {
		c = in[i]
		switch {
		case until != -1 && rune(c) == until:
			in = in[:i]
			goto done
		case scopedOnly && i+3 < len(in) && c == '\\' && // Treat \\LEAD{ as literal text \ followed by an expansion
			in[i+1] == '\\' && in[i+2] == leadChar && in[i+3] == openChar:
			consume(&expansion{value: in[:i]})
			in = in[i+1:]
			continue
		case scopedOnly && i+2 < len(in) && c == '\\' && // Treat \LEAD{ as literal text LEAD{
			in[i+1] == leadChar && in[i+2] == openChar:
			consume(&expansion{value: in[:i]})
			in = in[i+1:]
			continue
		case c != leadChar:
			continue
		case len(in) <= i+1:
			// consume(&expansion{value: in[:i], kind: exPrefix})
		case !scopedOnly && in[i+1] == leadChar:
			consume(&expansion{value: in[:i]})
			in = in[i+1:]
			continue
		case scopedOnly && c == leadChar && in[i+1] != openChar:
			continue
		case i > 0:
			consume(&expansion{value: in[:i], kind: exPrefix})
		}

		i++
		break
	}

	if i >= len(in) {
		goto done
	} else if in[i] == openChar {
		i++
	} else {
		head := i
		for ; i < len(in); i++ {
			if p.isIDChar(in[i]) {
				continue
			} else if i != head {
				consume(&expansion{value: in[head:i], kind: exVariable})
			}
			goto reset
		}

		consume(&expansion{value: in[head:i], kind: exVariable})
		goto done
	}

	for head := i; i < len(in); i++ {
		switch c = in[i]; c {
		case closeChar:
			tail := i
			i++
			if head != tail {
				consume(&expansion{value: in[head:tail], kind: exVariable})
			}
			goto reset

		case ':':
			var (
				n       int
				back    *expansion
				tail    = i
				kind    = exVariable
				varname string
			)
			i++
			if head == tail { // No variable name
				goto reset
			}

			varname = strings.TrimSpace(in[head:tail])
			switch in[i] {
			case '-':
			case '+':
				kind = exDefined
			case '/':
				kind = exUndefined
			default:
				goto reset
			}

			i++
			n, back = p.parse(in[i:], rune(closeChar))
			i += n
			if i < len(in) && in[i] == closeChar {
				i++
				consume(&expansion{value: varname, kind: kind, children: []*expansion{back}})
			} else {
				consume(&expansion{value: in[:tail+2], kind: exPrefix})
				consume(back)
			}

			goto reset
		}

		if !p.isIDChar(c) {
			consume(&expansion{value: in[:i]})
			goto reset
		}
	}

done:
	if len(in) > 0 {
		consume(&expansion{value: in})
	}

	if root != nil && root.value == "" && len(root.children) == 1 {
		root = root.children[0]
	}

	return n, root
}

type expansion struct {
	value    string
	kind     exKind
	children []*expansion
}

type exKind int

const (
	exPrefix    exKind = iota // Literal text
	exVariable                // $VAR ${VAR} ${VAR:-UNDEFINED TEXT}
	exDefined                 // ${VAR:+DEFINED TEXT}   (if undefined, expansion is empty)
	exUndefined               // ${VAR:?UNDEFINED TEXT} (if defined, expansion is empty)
)

func (e *expansion) expand(w *bytes.Buffer, fn LookupFunc) {
	if e == nil {
		return
	}

	switch e.kind {
	case exPrefix:
		if v := e.value; v != "" {
			w.WriteString(v)
		}
	case exVariable:
		if v, ok := fn(e.value); ok {
			if v != "" {
				w.WriteString(v)
			}
			return
		}
	case exDefined, exUndefined:
		if _, ok := fn(e.value); ok != (e.kind == exDefined) {
			return
		}
	}

	for _, e := range e.children {
		e.expand(w, fn)
	}
}
