package railing

import (
	"net/url"
	"reflect"
	"testing"
)

type T struct {
	X string `railing:"x"`
	EmbeddedT
	Foo foo `railing:"foo"`
	A   A   `railing:"a"`
}

type EmbeddedT struct {
	Int int     `railing:"int"`
	X   float32 `railing:"x"`
}

type A struct {
	Int int       `railing:"int"`
	A   EmbeddedA `railing:"a"`
}

type EmbeddedA struct {
	Int         int       `railing:"int"`
	Foos        []foo     `railing:"foos"`
	Slice       []float64 `railing:"slice"`
	SliceJoined []int     `railing:"ints,comma"`
}

func TestQuery(t *testing.T) {
	expected := &T{
		X: "x",
		EmbeddedT: EmbeddedT{
			Int: 2,
		},
		Foo: foo{
			ID:      1,
			Name:    "foo",
			Pointer: pointer{pint(5)},
			Slice:   sliceInt{1, 2, 3},
		},
		A: A{
			Int: 5,
			A: EmbeddedA{
				Int: 5,
				Foos: []foo{
					{
						ID:      2,
						Name:    "foo1",
						Pointer: pointer{pint(2)},
						Slice:   sliceInt{1, 2},
					},
					{
						ID:      2,
						Name:    "foo3",
						Pointer: pointer{pint(3)},
						Slice:   sliceInt{2},
					},
				},
				Slice:       []float64{1.1, 2.2},
				SliceJoined: []int{1, 2, 3},
			},
		},
	}
	query := "a%5Ba%5D%5Bfoos%5D%5B%5D%5Bid%5D=2&a%5Ba%5D%5Bfoos%5D%5B%5D%5Bnam" +
		"e%5D=foo1&a%5Ba%5D%5Bfoos%5D%5B%5D%5Bpointer%5D%5Bpint%5D=2&a%5Ba%5D%5Bf" +
		"oos%5D%5B%5D%5Bslice%5D=1%2C2&a%5Ba%5D%5Bfoos%5D%5B%5D%5Bid%5D=2&a%5Ba%5" +
		"D%5Bfoos%5D%5B%5D%5Bname%5D=foo3&a%5Ba%5D%5Bfoos%5D%5B%5D%5Bpointer%5D%5" +
		"Bpint%5D=3&a%5Ba%5D%5Bfoos%5D%5B%5D%5Bslice%5D=2&a%5Ba%5D%5Bint%5D=5&a%5" +
		"Ba%5D%5Bints%5D=1%2C2%2C3&a%5Ba%5D%5Bslice%5D%5B%5D=1.1&a%5Ba%5D%5Bslice" +
		"%5D%5B%5D=2.2&a%5Bint%5D=5&foo%5Bid%5D=1&foo%5Bname%5D=foo&foo%5Bpointer" +
		"%5D%5Bpint%5D=5&foo%5Bslice%5D=1%2C2%2C3&int=2&x=x"
	m, err := url.ParseQuery(query)
	if err != nil {
		t.Fatalf("expected err=nil; got %v", err)
	}
	var tt T
	if err := Unmarshal(Values{m}, &tt); err != nil {
		t.Fatalf("expected err=nil; got %v", err)
	}
	if !reflect.DeepEqual(expected, &tt) {
		t.Errorf("expected %v; got %v", expected, &tt)
	}
	val, err := Marshal(&tt)
	if err != nil {
		t.Fatalf("expected err=nil; got %v", err)
	}
	if encoded := val.Encode(); encoded != query {
		t.Errorf("expected %s; got %s", query, encoded)
	}
}
