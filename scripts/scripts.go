package scripts

import (
	"embed"
	"fmt"
)

//go:embed *.dxl
var fs embed.FS

// preamble is prepended to every utility script. It provides JSON helpers so
// each script only has to emit a final JSON string expression (the value
// returned by eval_).
//
// NOTE: written against the DOORS 9.7 DXL Reference; the escaping helper is
// the piece most likely to need a tweak on a real install — validate it first
// there with a tiny eval_ sanity check.
const preamble = `
// ----- JSON helpers -----
string jsonEscape(string s) {
	string out = ""
	int i
	string ch
	for (i = 0; i < length(s); i++) {
		ch = s[i]
		if (ch == "\\")      out = out + "\\\\"
		else if (ch == "\"") out = out + "\\\""
		else if (ch == "\n") out = out + "\\n"
		else if (ch == "\t") out = out + "\\t"
		else if (ch == "\r") { /* drop */ }
		else                 out = out + ch
	}
	return out
}

string jstr(string s) {
	return "\"" + jsonEscape(s) + "\""
}

string intToStr(int n) {
	if (n == 0) {
		return "0"
	}
	string digits = "0123456789"
	string out = ""
	int v = n
	if (v < 0) {
		return "-" + intToStr(0 - v)
	}
	while (v > 0) {
		int d = v % 10
		out = digits[d] + out
		v = v / 10
	}
	return out
}

// error envelope builder
string jerr(string msg) {
	return "{\"status\":\"error\",\"message\":" + jstr(msg) + "}"
}
// --------------------------
`

// Load returns the full template text (preamble + utility body) for the named script.
func Load(name string) (string, error) {
	body, err := fs.ReadFile(name + ".dxl")
	if err != nil {
		return "", fmt.Errorf("load script %q: %w", name, err)
	}
	return preamble + "\n" + string(body), nil
}

// Names lists the available utility script names (without .dxl extension).
func Names() []string {
	entries, err := fs.ReadDir(".")
	if err != nil {
		return nil
	}
	var names []string
	for _, e := range entries {
		n := e.Name()
		if len(n) > 4 && n[len(n)-4:] == ".dxl" {
			names = append(names, n[:len(n)-4])
		}
	}
	return names
}
