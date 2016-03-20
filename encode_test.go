package railing

import (
	"net/url"
	"reflect"
	"strings"
	"testing"
)

type marshalTest []struct {
	in  interface{}
	out url.Values
	err error
}

type omit struct {
	Int int `railing:",omitempty"`
}

func (j *joinedStr) MarshalQuery() (Values, error) {
	return Values{url.Values{"Str": strings.Split(j.Str, ",")}}, nil
}

var nilInterface interface{}
var nilpointer *all
var nilIntPointer *int

var marshaler Marshaler = &joinedStr{Str: "str"}

var dpointer = &pointer{pint(5)}
var ddpointer = &dpointer

type Interface interface{}

func TestMarshal(t *testing.T) {
	fixtures := marshalTest{
		// 0
		{
			in: all{
				String:      "string",
				Float32:     1,
				Float64:     1,
				Int:         1,
				Int8:        1,
				Int16:       1,
				Int32:       1,
				Int64:       1,
				Uint:        1,
				Uint8:       1,
				Uint16:      1,
				Uint32:      1,
				Uint64:      1,
				Bool:        true,
				SliceString: []string{"string1", "string2"},
				SliceInt:    []int{1, 2},
				SliceUint:   []uint{1, 2},
				SliceBool:   []bool{true, false},
				SliceFloat:  []float32{1.0, 2.0},
				SlicePInt:   []*int{pint(1), pint(2)},
			},
			out: url.Values{
				"string":         []string{"string"},
				"float32":        []string{"1"},
				"float64":        []string{"1"},
				"int":            []string{"1"},
				"int8":           []string{"1"},
				"int16":          []string{"1"},
				"int32":          []string{"1"},
				"int64":          []string{"1"},
				"uint":           []string{"1"},
				"uint8":          []string{"1"},
				"uint16":         []string{"1"},
				"uint32":         []string{"1"},
				"uint64":         []string{"1"},
				"bool":           []string{"true"},
				"slice_string[]": []string{"string1", "string2"},
				"slice_int[]":    []string{"1", "2"},
				"slice_uint[]":   []string{"1", "2"},
				"slice_bool[]":   []string{"true", "false"},
				"slice_float[]":  []string{"1", "2"},
				"slice_pint[]":   []string{"1", "2"},
			},
			err: nil,
		},
		// 1
		{
			in: struct {
				Slice []string `railing:"slice,comma"`
			}{[]string{"1", "2"}},
			out: url.Values{
				"slice": []string{"1,2"},
			},
			err: nil,
		},
		// 2
		{
			in:  omit{},
			out: make(url.Values),
		},
		// 3
		{
			in: omit{1},
			out: url.Values{
				"Int": []string{"1"},
			},
		},
		// 4
		{
			in: joinedStrParent{joinedStr{Str: "foo"}},
			out: url.Values{
				"joined[Str]": []string{"foo"},
			},
			err: nil,
		},
		// 5
		{
			in: Embedded1{Embedded: Embedded{Int: 1}},
			out: url.Values{
				"int": []string{"1"},
			},
			err: nil,
		},
		// 6
		{
			in: Embedded5{Embedded1: &Embedded1{Embedded: Embedded{Int: 1}}},
			out: url.Values{
				"int": []string{"1"},
			},
			err: nil,
		},
		// 7
		{
			in:  Embedded5{},
			out: make(url.Values),
			err: nil,
		},
		// 8, MarshalQuery() will not work on embedded struct
		{
			in: Embedded6{joinedStr{"1,2"}},
			out: url.Values{
				"embedded[Str]": []string{"1,2"},
			},
		},
		// 9
		{
			in: struct {
				Interface
			}{},
			out: make(url.Values),
		},
		// 10
		{
			in: struct {
				Interface
			}{joinedStr{"1,2"}},
			out: url.Values{
				"Str": []string{"1,2"},
			},
		},
		// 11
		{
			in: struct {
				Interface `railing:"interface"`
			}{joinedStr{"1,2"}},
			out: url.Values{
				"interface[Str]": []string{"1,2"},
			},
		},
		// 12
		{
			in: Embedded3{Embedded{1}, 2},
			out: url.Values{
				"int": []string{"2"},
			},
			err: nil,
		},
		// 13
		{
			in: struct {
				Foos []foo `railing:"foos"`
			}{
				[]foo{
					{
						ID:      1,
						Name:    "one",
						Pointer: pointer{pint(2)},
						Slice:   sliceInt{1, 2, 3},
					},
					{
						ID:      2,
						Name:    "two",
						Pointer: pointer{pint(3)},
						Slice:   sliceInt{4, 5, 6},
					},
				}},
			out: url.Values{
				"foos[][id]":            []string{"1", "2"},
				"foos[][name]":          []string{"one", "two"},
				"foos[][pointer][pint]": []string{"2", "3"},
				"foos[][slice]":         []string{"1,2,3", "4,5,6"},
			},
			err: nil,
		},
		// 14
		{
			in: map[string]string{
				"key": "val",
			},
			out: url.Values{
				"key": []string{"val"},
			},
			err: nil,
		},
		// 15
		{
			in: map[string][]int{
				"ints": {1, 2, 3},
			},
			out: url.Values{
				"ints[]": []string{"1", "2", "3"},
			},
			err: nil,
		},
		// 16
		{
			in: map[string][3]int{
				"ints": [...]int{1, 2, 3},
			},
			out: url.Values{
				"ints[]": []string{"1", "2", "3"},
			},
			err: nil,
		},
		// 17
		{
			in: map[string]interface{}{
				"key": "value",
				"obj": map[string]interface{}{
					"id":    5,
					"slice": []int{1, 2},
				},
				"foo":        []interface{}{1, 2, "foo"},
				"marshaler":  []joinedStr{{"1,2"}, {"a,b"}},
				"marshalerp": []*joinedStr{{"1,2"}, {"a,b"}},
			},
			out: url.Values{
				"key":               []string{"value"},
				"obj[id]":           []string{"5"},
				"obj[slice][]":      []string{"1", "2"},
				"foo[]":             []string{"1", "2", "foo"},
				"marshaler[][Str]":  []string{"1,2", "a,b"},
				"marshalerp[][Str]": []string{"1,2", "a,b"},
			},
		},
		// 18
		{
			in:  pointer{},
			out: make(url.Values),
			err: nil,
		},
		// 19
		{
			in:  nilpointer,
			out: make(url.Values),
			err: nil,
		},
		// 20
		{
			in: map[string][]*int{
				"ints":  {nilIntPointer, pint(1)},
				"nil[]": nil,
			},
			out: url.Values{"ints[]": []string{"1"}},
			err: nil,
		},
		// 21
		{
			in: &joinedStr{Str: "1,2,3"},
			out: url.Values{
				"Str": []string{"1", "2", "3"},
			},
			err: nil,
		},
		// 22
		{
			in: &joinedStrParent{JoinedStr: joinedStr{Str: "1,2,3"}},
			out: url.Values{
				"joined[Str]": []string{"1", "2", "3"},
			},
			err: nil,
		},
		// 23
		{
			in: marshaler,
			out: url.Values{
				"Str": []string{"str"},
			},
			err: nil,
		},
		// 24
		{
			in: struct {
				P **pointer
			}{P: ddpointer},
			out: url.Values{
				"P[pint]": []string{"5"},
			},
		},
		// 25
		{
			in: struct {
				A []omit
			}{A: []omit{{1}, {2}}},
			out: url.Values{
				"A[][Int]": []string{"1", "2"},
			},
		},
		// 26
		{
			in: struct {
				Slice []int `railing:"slice,comma"`
			}{[]int{1, 2}},
			out: url.Values{
				"slice": []string{"1,2"},
			},
		},
		// 27
		{
			in:  nil,
			out: make(url.Values),
		},
		// 28
		{
			in:  nilInterface,
			out: make(url.Values),
		},
		//
		// errors
		//
		// 29
		{
			in:  []string{"slice"},
			err: &UnsupportedTypeError{reflect.TypeOf([]string{})},
		},
		// 30
		{
			in: struct {
				Ch chan struct{}
			}{},
			err: &UnsupportedTypeError{reflect.TypeOf(make(chan struct{}))},
		},
	}
	for i, fixture := range fixtures {
		out, err := Marshal(fixture.in)
		if !reflect.DeepEqual(fixture.err, err) {
			t.Errorf("expected err=%v; got %v (i=%d)", fixture.err, err, i)
			continue
		}
		if !reflect.DeepEqual(out.Values, fixture.out) {
			t.Errorf("expected %#v; got %#v (i=%d)", fixture.out, out.Values, i)
		}
	}
}
