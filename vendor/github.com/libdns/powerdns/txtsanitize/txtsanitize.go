package txtsanitize

import (
	"bytes"
	"strings"
)

// TXTSanitize - attempts to make sure that the return value is enclosed in
// double quotes, and any embedded double quotes are escaped properly with `\`.
// It honors pre-escaped sequences inside, in that it ignores all but those
// preceding an embedded double quote. The return value is idempotent, meaning
// you can feed the output back into TXTSanitize and you should get the same
// result as the initial call.  This is important so that we don't re-escape
// things such as in the case when reading a TXT record from an API call and
// sending the payload back in a subsequent update call.
func TXTSanitize(in string) string {
	quoted, contents := false, []byte(in)
	if len(in) >= 2 {
		quoted = contents[0] == '"' && contents[len(in)-1] == '"'
	}
	if quoted {
		contents = contents[1 : len(in)-1]
	}
	escaped := 0
	var bldr, out strings.Builder

	for ind := 0; ind < len(contents); ind++ {
		tInd := bytes.IndexByte(contents[ind:], '"')
		if tInd == -1 {
			bldr.Write(contents[ind:])
			break
		}
		bldr.Write(contents[ind : ind+tInd])
		ind += tInd

		// look for \", but be aware of \\" is not escaped, but \\\" is
		escCt := 0
		for j := ind - 1; j >= 0 && contents[j] == '\\'; j-- {
			escCt++
		}
		if escCt%2 == 0 {
			// add escape
			escaped++
			bldr.WriteByte('\\')
		}
		bldr.WriteByte('"')
	}

	// This tries to catch the situation where we have something like:
	// "foo" and other stuff "bar"
	// and make sure we get:
	// "\"foo\" and other stuff \"bar\""
	// which is likely what we want, instead of:
	// "foo\" and other stuff \"bar"
	out.WriteByte('"')
	if quoted && escaped > 0 && escaped%2 == 0 {
		out.WriteString(`\"`)
		out.WriteString(bldr.String())
		out.WriteString(`\"`)
	} else {
		out.WriteString(bldr.String())
	}
	out.WriteByte('"')
	return out.String()
}
