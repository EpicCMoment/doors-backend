package dxl

import "strconv"

// DxlLiteral holds one typed parameter and knows how to render its DXL
// source literal. The variant determines the literal form: strings are
// quoted/escaped, integers are plain, booleans become DXL true/false, and
// raw strings are spliced in verbatim.
type DxlLiteral struct {
	literalKind DxlLiteralKind
	text        string
	integer     int64
	boolean     bool
}

type DxlLiteralKind int

const (
	dxlString DxlLiteralKind = iota
	dxlInteger
	dxlBoolean
	dxlRaw
)

func String(s string) DxlLiteral { return DxlLiteral{literalKind: dxlString, text: s} }
func Int(n int64) DxlLiteral     { return DxlLiteral{literalKind: dxlInteger, integer: n} }
func Bool(b bool) DxlLiteral     { return DxlLiteral{literalKind: dxlBoolean, boolean: b} }

// Raw splices pre-validated DXL source directly into the script.
func Raw(src string) DxlLiteral { return DxlLiteral{literalKind: dxlRaw, text: src} }

// Literal renders the DXL source form of the value.
func (l DxlLiteral) Literal() string {
	switch l.literalKind {
	case dxlRaw:
		return l.text
	case dxlString:
		return strconv.Quote(l.text) // quoted, with backslash escapes
	case dxlInteger:
		return strconv.FormatInt(l.integer, 10)
	case dxlBoolean:
		if l.boolean {
			return "true"
		}
		return "false"
	}
	return ""
}
