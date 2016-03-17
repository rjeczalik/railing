package railing

import (
	"net/url"
	"reflect"
	"regexp"
	"strings"
)

// reTopKey returns the top key in match[1] and the rest in match[2]
//
// it matches for example:
//  - foo[][id]  -> [1] = foo, [2] = [][id], [3] = [], [4] = [id]
//  - foo[id]    -> [1] = foo, [2] = [id],   [3] = ,   [4] = [id]
//  - foo        -> [1] = foo, [2] = ,       [3] = ,   [4] =
var reTopKey = regexp.MustCompile(`([^\[]+)((\[\])?(.*))`)

// reObject helps to detect complex rails style query params objects in
// url.Values map.
//
// it matches for example:
//  - foo[][id]      -> match[1] = foo, match[2] = id
//  - foo[id]        -> match[1] = foo, match[2] = id
//  - foo[bar][name] -> match[1] = foo, match[2] = bar, match[3] = [name]
//
var reObject = regexp.MustCompile(`([^\[]+)\[(?:\]\[)?([^\]]+)\](.*)`)

// findValues searches the url.Values for the given tag. It returns either
// a subMap or url.Values's value - []string. This depends on the tag. If its
// a nested struct, or an array.
//
// url.Values{
//  "normal":         []string{},
//  "array[]":        []string{},
//  "nested[object]": []string{},
// }
//
// The third case - "nested[object]" will return a subMap of all keys which
// start with "nested".
func findValues(m url.Values, tag string) (url.Values, []string) {
	if v, ok := m[tag]; ok {
		return nil, v
	}
	if v, ok := m[tag+"[]"]; ok {
		return nil, v
	}
	if submap := subMap(m, strings.TrimSuffix(tag, "[]")); len(submap) > 0 {
		return submap, nil
	}
	return nil, nil
}

// subMap returns a sub map of m which contains every key which starts with the
// given key arg.
//
// subMap(m, "nested")
//
// url.Values{
//  "foo":          []string{}  ->  url.Values{
//  "nested[id]":   []string{}, ->    "id":   []string{},
//  "nested[name]": []string{}, ->    "name": []string{},
// }                            ->  }
//
func subMap(m url.Values, key string) url.Values {
	subm := make(url.Values)
	for k, v := range m {
		match := reObject.FindStringSubmatch(k)
		if match == nil || match[1] != key {
			continue
		}
		subm[match[2]+match[3]] = v
	}
	return subm
}

// tag describes 'railing' tag and it's options for the given field.
//
// name      - is the tag's first argument or field name.
//
// omitEmpty - is true when the tag contains 'omitempty' option, it means that
//             the field will not be shown in encoded Values if it's value is
//             empty.
//
// comma     - is when the tag contains 'comma' option, it means that values
//             which are slices or arrays will be joined by ',' character.
//
// ignore    - is when the tag string is '-'. Such fields are going to be
//             ignored.
//
// empty     - is an internal field, it says if the tag is empty or not.
type tag struct {
	name      string
	omitEmpty bool
	comma     bool
	ignore    bool
	empty     bool
}

func parseTag(field reflect.StructField) (t tag) {
	tags := strings.Split(field.Tag.Get("railing"), ",")
	if len(tags) == 1 && tags[0] == "" {
		t.empty = true
		t.name = field.Name
		return
	}
	switch tags[0] {
	case "-":
		t.ignore = true
		return
	case "":
		t.name = field.Name
	default:
		t.name = tags[0]
	}
	for _, tagOpt := range tags[1:] {
		switch tagOpt {
		case "omitempty":
			t.omitEmpty = true
		case "comma":
			t.comma = true
		}
	}
	return
}
