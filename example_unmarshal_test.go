package railing_test

import (
	"fmt"
	"image/color"
	"log"
	"net/url"

	"github.com/jszwec/railing"
)

func ExampleUnmarshal() {
	type ColorMap struct {
		ID      int
		Palette []color.RGBA
	}

	values := railing.Values{Values: url.Values{
		"ID":           {"1"},
		"Palette[][R]": {"255", "0"},
		"Palette[][G]": {"0", "255"},
		"Palette[][B]": {"0", "0"},
		"Palette[][A]": {"0", "0"},
	}}

	var colorMap ColorMap
	if err := railing.Unmarshal(values, &colorMap); err != nil {
		log.Fatal(err)
	}
	fmt.Println(colorMap)

	// Output:
	// {1 [{255 0 0 0} {0 255 0 0}]}
}

func ExampleUnmarshal_maps() {
	values := railing.Values{Values: url.Values{
		"first_array": {"1", "2"},
	}}

	intMap := make(map[string][]int)
	if err := railing.Unmarshal(values, &intMap); err != nil {
		log.Fatal(err)
	}
	fmt.Println(intMap)

	// Output:
	// map[first_array:[1 2]]
}

func ExampleUnmarshal_interface() {
	values := railing.Values{Values: url.Values{
		"person[name]": {"bob"},
	}}

	var v interface{}
	if err := railing.Unmarshal(values, &v); err != nil {
		log.Fatal(err)
	}
	// v becomes map[string]interface{}
	fmt.Println(v)

	// Output:
	// map[person:map[name:[bob]]]
}
