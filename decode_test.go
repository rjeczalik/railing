package railing

import (
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

type unmarshalTest struct {
	in  url.Values
	ptr interface{}
	out interface{}
	err error
}

type all struct {
	String      string    `railing:"string"`
	Float32     float32   `railing:"float32"`
	Float64     float64   `railing:"float64"`
	Int         int       `railing:"int"`
	Int8        int8      `railing:"int8"`
	Int16       int16     `railing:"int16"`
	Int32       int32     `railing:"int32"`
	Int64       int64     `railing:"int64"`
	Uint        uint      `railing:"uint"`
	Uint8       uint8     `railing:"uint8"`
	Uint16      uint16    `railing:"uint16"`
	Uint32      uint32    `railing:"uint32"`
	Uint64      uint64    `railing:"uint64"`
	Bool        bool      `railing:"bool"`
	unexported  int       `railing:"unexported"`
	Ignore      int       `railing:"-"`
	SliceString []string  `railing:"slice_string"`
	SliceInt    []int     `railing:"slice_int"`
	SliceUint   []uint    `railing:"slice_uint"`
	SliceBool   []bool    `railing:"slice_bool"`
	SliceFloat  []float32 `railing:"slice_float"`
	SlicePInt   []*int    `railing:"slice_pint"`
}

type arrays struct {
	ArrayInt [2]int `railing:"array_int"`
}

type Embedded struct {
	Int int `railing:"int"`
}

type Embedded1 struct {
	Embedded
	unexported *all
}

type Embedded2 struct {
	*Embedded
}

type Embedded3 struct {
	Embedded
	Int int `railing:"int"`
}

type Embedded4 struct {
	*Embedded2
}

type Embedded5 struct {
	*Embedded1
}

type Embedded6 struct {
	joinedStr `railing:"embedded"`
}

type pointer struct {
	Pint *int `railing:"pint"`
}

type joinedStrParent struct {
	JoinedStr joinedStr `railing:"joined"`
}

type joinedStr struct {
	Str string
}

type I struct {
	U Unmarshaler `railing:"unmarshaler"`
}

type interfaceParent struct {
	Interface interface{} `railing:"interface"`
}

type comma struct {
	Ints []int `railing:"ints,comma"`
}

func (js *joinedStr) UnmarshalQuery(v Values) error {
	js.Str = strings.Join(v.Values["Str"], ",")
	return nil
}

func pint(n int) *int {
	return &n
}

var um Unmarshaler = &joinedStr{}

type Map map[string][]string

type M struct {
	Map
	I int `railing:"i"`
}

type foo struct {
	ID      int      `railing:"id"`
	Name    string   `railing:"name"`
	Pointer pointer  `railing:"pointer"`
	Slice   sliceInt `railing:"slice"`
}

type sliceInt []int

func (s *sliceInt) UnmarshalQuery(v Values) error {
	strs := strings.Split(strings.Join(v.Values["slice"], ","), ",")
	for _, str := range strs {
		n, err := strconv.Atoi(str)
		if err != nil {
			return err
		}
		*s = append(*s, n)
	}
	return nil
}

func (s *sliceInt) MarshalQuery() (Values, error) {
	strs := make([]string, 0, len(*s))
	for _, i := range *s {
		strs = append(strs, strconv.Itoa(i))
	}
	return Values{url.Values{
		"": []string{strings.Join(strs, ",")},
	}}, nil
}

type structSlice struct {
	Foos []foo `railing:"foo"`
}

type structArray struct {
	Foos [2]foo `railing:"foo"`
}

func TestUnmarshal(t *testing.T) {
	fixtures := []unmarshalTest{
		// 0
		{
			in: url.Values{
				"string":       []string{"string"},
				"float32":      []string{"1.0"},
				"float64":      []string{"1.0"},
				"int":          []string{"1"},
				"int8":         []string{"1"},
				"int16":        []string{"1"},
				"int32":        []string{"1"},
				"int64":        []string{"1"},
				"uint":         []string{"1"},
				"uint8":        []string{"1"},
				"uint16":       []string{"1"},
				"uint32":       []string{"1"},
				"uint64":       []string{"1"},
				"bool":         []string{"true"},
				"unexported:":  []string{"1"},
				"Ignore":       []string{"1"},
				"slice_string": []string{"string1", "string2"},
				"slice_int[]":  []string{"1", "2"},
				"slice_uint":   []string{"1", "2"},
				"slice_bool":   []string{"true", "false"},
				"slice_float":  []string{"1.0", "2.0"},
				"slice_pint":   []string{"1", "2"},
			},
			ptr: new(all),
			out: all{
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
		},
		// 1
		{
			in:  url.Values{"int": []string{"1"}},
			ptr: new(Embedded1),
			out: Embedded1{Embedded{Int: 1}, nil},
		},
		// 2
		{
			in:  url.Values{"int": []string{"1"}},
			ptr: new(Embedded2),
			out: Embedded2{&Embedded{Int: 1}},
		},
		// 3
		{
			in:  url.Values{"int": []string{"1"}},
			ptr: new(Embedded3),
			out: Embedded3{Embedded: Embedded{Int: 0}, Int: 1},
		},
		// 4
		{
			in:  url.Values{"pint": []string{"1"}},
			ptr: new(pointer),
			out: pointer{Pint: pint(1)},
		},
		// 5
		{
			in:  url.Values{"int": []string{"1"}},
			ptr: new(Embedded4),
			out: Embedded4{&Embedded2{&Embedded{Int: 1}}},
		},
		// 6
		{
			in:  url.Values{"int": []string{"1"}},
			ptr: new(Embedded5),
			out: Embedded5{&Embedded1{Embedded{Int: 1}, nil}},
		},
		// 7
		{
			in: url.Values{
				"joined[Str]": []string{"one", "two", "three"},
			},
			ptr: new(joinedStrParent),
			out: joinedStrParent{joinedStr{"one,two,three"}},
		},
		// 8
		{
			in:  url.Values{"interface": []string{"5"}},
			ptr: new(interfaceParent),
			out: interfaceParent{Interface: []string{"5"}},
		},
		// 9
		{
			in:  url.Values{"array_int": []string{"5", "4", "3", "2"}},
			ptr: new(arrays),
			out: arrays{ArrayInt: [2]int{5, 4}},
		},
		// 10
		{
			in:  url.Values{"array_int": []string{"5"}},
			ptr: new(arrays),
			out: arrays{ArrayInt: [2]int{5, 0}},
		},
		// 11
		{
			in:  url.Values{"array_int": []string{"5", "4"}},
			ptr: new(arrays),
			out: arrays{ArrayInt: [2]int{5, 4}},
		},
		// 12
		{
			in: url.Values{
				"string":      []string{"string", "string2"},
				"float32":     []string{"1.0", "1"},
				"float64":     []string{"1.0", "1"},
				"int":         []string{"1", "1"},
				"int8":        []string{"1", "1"},
				"int16":       []string{"1", "1"},
				"int32":       []string{"1", "1"},
				"int64":       []string{"1", "1"},
				"uint":        []string{"1", "1"},
				"uint8":       []string{"1", "1"},
				"uint16":      []string{"1", "1"},
				"uint32":      []string{"1", "1"},
				"uint64":      []string{"1", "1"},
				"bool":        []string{"true", "1"},
				"unexported:": []string{"1", "1"},
			},
			ptr: new(all),
			out: all{
				String:  "string",
				Float32: 1,
				Float64: 1,
				Int:     1,
				Int8:    1,
				Int16:   1,
				Int32:   1,
				Int64:   1,
				Uint:    1,
				Uint8:   1,
				Uint16:  1,
				Uint32:  1,
				Uint64:  1,
				Bool:    true,
			},
		},
		// 13
		{
			in: url.Values{
				"string":      nil,
				"float32":     nil,
				"float64":     nil,
				"int":         nil,
				"int8":        nil,
				"int16":       nil,
				"int32":       nil,
				"int64":       nil,
				"uint":        nil,
				"uint8":       nil,
				"uint16":      nil,
				"uint32":      nil,
				"uint64":      nil,
				"bool":        nil,
				"unexported:": nil,
			},
			ptr: new(all),
			out: all{},
		},
		// 14
		{
			in:  url.Values{"Str": []string{"5", "4", "2"}},
			ptr: new(Embedded6),
			out: Embedded6{joinedStr{"5,4,2"}},
		},
		// 15
		{
			in:  url.Values{"embedded": nil},
			ptr: new(Embedded6),
			out: Embedded6{joinedStr{""}},
		},
		// 16
		{
			in:  url.Values{"embedded": []string{}},
			ptr: new(Embedded6),
			out: Embedded6{joinedStr{""}},
		},
		// 17
		{
			in:  url.Values{"ids[]": []string{"1", "2", "3"}},
			ptr: new(map[string][]int),
			out: map[string][]int{"ids": {1, 2, 3}},
		},
		// 18
		{
			in: url.Values{
				"ids":         []string{"1", "2", "3"},
				"car[wheels]": []string{"4"},
				"car[color]":  []string{"red"},
			},
			ptr: new(map[string]interface{}),
			out: map[string]interface{}{
				"ids": []string{"1", "2", "3"},
				"car": map[string]interface{}{
					"wheels": []string{"4"},
					"color":  []string{"red"},
				},
			},
		},
		// 19
		{
			in:  url.Values{"ids": []string{"1", "2", "3"}},
			ptr: new(map[string]int),
			out: map[string]int{"ids": 1},
		},
		// 20
		{
			in: url.Values{
				"ids":                []string{"1", "2", "3"},
				"slice[]":            []string{"1", "2"},
				"foo[a]":             []string{"1"},
				"car[wheels]":        []string{"4"},
				"car[color]":         []string{"red"},
				"car[specs][length]": []string{"5"},
			},
			ptr: new(interface{}),
			out: map[string]interface{}{
				"ids":   []string{"1", "2", "3"},
				"slice": []string{"1", "2"},
				"foo": map[string]interface{}{
					"a": []string{"1"},
				},
				"car": map[string]interface{}{
					"wheels": []string{"4"},
					"color":  []string{"red"},
					"specs": map[string]interface{}{
						"length": []string{"5"},
					},
				},
			},
		},
		// 21
		{
			in:  url.Values{"Str": []string{"1", "2", "3"}},
			ptr: um,
			out: joinedStr{Str: "1,2,3"},
		},
		// 22
		{
			in: url.Values{
				"i":   []string{"1"},
				"lol": []string{"lol"},
			},
			ptr: new(M),
			out: M{I: 1, Map: map[string][]string{"lol": {"lol"}}},
		},
		// 23
		{
			in: url.Values{
				"foo[][id]":            []string{"1", "2"},
				"foo[][name]":          []string{"a", "b"},
				"foo[][pointer][pint]": []string{"5", "0"},
				"foo[][slice]":         []string{"1,2,3", "4,5,6"},
			},
			ptr: new(structSlice),
			out: structSlice{Foos: []foo{
				{1, "a", pointer{pint(5)}, []int{1, 2, 3}},
				{2, "b", pointer{pint(0)}, []int{4, 5, 6}}},
			},
		},
		// 24
		{
			in: url.Values{
				"foo[][id]":            []string{"1", "2", "3"},
				"foo[][name]":          []string{"a", "b", "c"},
				"foo[][pointer][pint]": []string{"5", "0", "1"},
				"foo[][slice]":         []string{"1,2,3", "4,5,6", "99"},
			},
			ptr: new(structArray),
			out: structArray{Foos: [2]foo{
				{1, "a", pointer{pint(5)}, []int{1, 2, 3}},
				{2, "b", pointer{pint(0)}, []int{4, 5, 6}}},
			},
		},
		// 25
		{
			in:  url.Values{"unmarshaler[Str]": []string{"one", "two", "three"}},
			ptr: &I{um},
			out: I{&joinedStr{"one,two,three"}},
		},
		// 26
		{
			in: url.Values{
				"string": nil,
				"int":    []string{},
				"bool":   []string{"t"},
			},
			ptr: new(all),
			out: all{Bool: true},
		},
		// 27
		{
			in: url.Values{
				"ints": []string{"1,2"},
			},
			ptr: new(comma),
			out: comma{Ints: []int{1, 2}},
		},
		// 28
		{
			in: url.Values{
				"ints": []string{"1,2", "3,4"},
			},
			ptr: new(comma),
			out: comma{Ints: []int{1, 2, 3, 4}},
		},
		//
		// errors
		//
		// 29
		{
			in:  make(url.Values),
			ptr: new([]string),
			out: ([]string)(nil),
			err: &UnmarshalTypeError{"object", reflect.TypeOf([]string{})},
		},
		// 30
		{
			in: url.Values{
				"foo[][id]":            []string{"1", "2"},
				"foo[][name]":          []string{"a", "b", "c"},
				"foo[][pointer][pint]": []string{"5", "0"},
			},
			ptr: new(structSlice),
			out: structSlice{Foos: ([]foo)(nil)},
			err: errMissingData(reflect.TypeOf([]foo{})),
		},
		// 31
		{
			in:  url.Values{"int": []string{"lol"}},
			ptr: new(all),
			out: all{},
			err: &UnmarshalTypeError{"number lol", reflect.TypeOf(1)},
		},
		// 32
		{
			in:  url.Values{"unmarshaler": []string{"lol"}},
			ptr: new(I),
			out: I{},
			err: &UnmarshalTypeError{"object", reflect.ValueOf(I{}).Field(0).Type()},
		},
		// 33
		{
			in:  url.Values{"unmarshaler[name]": []string{"lol"}},
			ptr: new(I),
			out: I{},
			err: &UnmarshalTypeError{"object", reflect.ValueOf(I{}).Field(0).Type()},
		},
	}
	for i, fixture := range fixtures {
		v := reflect.ValueOf(fixture.ptr)
		if err := Unmarshal(Values{fixture.in},
			v.Interface()); !reflect.DeepEqual(fixture.err, err) {
			t.Errorf("expected err=%v; got %v (i=%d)", fixture.err, err, i)
			continue
		}
		if !reflect.DeepEqual(v.Elem().Interface(), fixture.out) {
			t.Errorf("expected %#v; got %#v (i=%d)", fixture.out,
				v.Elem().Interface(), i)
		}
	}
}
