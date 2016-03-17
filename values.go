package railing

import (
	"bytes"
	"fmt"
	"net/url"
	"sort"
	"strings"
)

const leftBracket = "%5B"
const rightBracket = "%5D"
const brackets = leftBracket + rightBracket

// Values wraps url.Values. It provides it's own Encode function in order to
// build a query string compatible with rack's parser.
type Values struct {
	url.Values
}

// Encode encodes the values into “URL encoded” form. Keys are sorted by key.
// However, in case of object array, every object element is sorted within.
//
// objects[][id]=1&objects[][name]=name1&objects[][id]=2&objects[][name]=name2
func (v *Values) Encode() string {
	if v == nil || v.Values == nil {
		return ""
	}
	var buf bytes.Buffer
	v.encode("", v.Values, &buf)
	return buf.String()
}

func (v *Values) encode(topPrefix string, m url.Values, buf *bytes.Buffer) {
	for _, k := range v.keys(m) {
		prefix := url.QueryEscape(strings.TrimSuffix(k, "[]"))
		if topPrefix != "" {
			prefix = fmt.Sprintf("%s%s%s%s",
				topPrefix, leftBracket, prefix, rightBracket)
		}
		subm, vals := findValues(m, k)
		switch {
		case vals != nil:
			v.encodeFlat(prefix, strings.HasSuffix(k, "[]"), vals, buf)
		case strings.HasSuffix(k, "[]"):
			v.encodeArray(prefix, subm, buf)
		default: // subm != nil
			v.encode(prefix, subm, buf)
		}
	}
}

func (v *Values) keys(m url.Values) []string {
	set := make(map[string]struct{})
	for k := range m {
		match := reTopKey.FindStringSubmatch(k)
		if match != nil {
			set[match[1]+match[3]] = struct{}{}
			continue
		}
		set[k] = struct{}{}
	}
	keys := make(sort.StringSlice, 0, len(m))
	for k := range set {
		keys = append(keys, k)
	}
	keys.Sort()
	return keys
}

func (v *Values) encodeFlat(prefix string, suffix bool, vals []string,
	buf *bytes.Buffer) {
	for _, v := range vals {
		if buf.Len() > 0 {
			buf.WriteByte('&')
		}
		buf.WriteString(prefix)
		if suffix {
			buf.WriteString(brackets)
		}
		buf.WriteString("=" + url.QueryEscape(v))
	}
}

func (v *Values) encodeArray(prefix string, m url.Values, buf *bytes.Buffer) {
	keys := make(sort.StringSlice, 0, len(m))
	l := 0
	for k, vals := range m {
		keys = append(keys, k)
		if len(vals) > l {
			l = len(vals)
		}
	}
	keys.Sort()
	objs := make([][]string, l)
	for _, k := range keys {
		for i, val := range m[k] {
			match := reTopKey.FindStringSubmatch(k)
			objs[i] = append(objs[i], fmt.Sprintf("%s%s%s%s%s%s=%s",
				prefix,
				brackets,
				leftBracket,
				url.QueryEscape(match[1]),
				rightBracket,
				url.QueryEscape(match[2]),
				url.QueryEscape(val)))
		}
	}
	for i := range objs {
		if buf.Len() > 0 {
			buf.WriteByte('&')
		}
		buf.WriteString(strings.Join(objs[i], "&"))
	}
}
