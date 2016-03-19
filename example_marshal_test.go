package railing_test

import (
	"fmt"
	"image/color"
	"log"
	"net/url"

	"github.com/jszwec/railing"
)

func ExampleMarshal() {
	type ColorMap struct {
		ID      int
		Palette []color.RGBA
	}

	cm := ColorMap{
		ID: 1,
		Palette: []color.RGBA{
			{255, 0, 0, 0},
			{0, 255, 0, 0},
		},
	}

	v, err := railing.Marshal(cm)
	if err != nil {
		log.Fatal(err)
	}

	// Print unescaped created query string.
	unescaped, err := url.QueryUnescape(v.Encode())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(unescaped)

	// Output:
	// ID=1&Palette[][A]=0&Palette[][B]=0&Palette[][G]=0&Palette[][R]=255&Palette[][A]=0&Palette[][B]=0&Palette[][G]=255&Palette[][R]=0
}

func ExampleMarshal_maps() {
	intMap := map[string][]int{
		"first_array":  []int{1, 2},
		"second_array": []int{3, 4},
	}

	v, err := railing.Marshal(intMap)
	if err != nil {
		log.Fatal(err)
	}

	// Print unescaped query string.
	unescaped, err := url.QueryUnescape(v.Encode())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(unescaped)

	// Output:
	// first_array[]=1&first_array[]=2&second_array[]=3&second_array[]=4
}

func ExampleMarshal_mapStringInterface() {
	m := map[string]interface{}{
		"name": "Bob",
		"address": map[string]interface{}{
			"city":  "New York",
			"state": "NY",
		},
	}

	v, err := railing.Marshal(m)
	if err != nil {
		log.Fatal(err)
	}

	// Print unescaped query string.
	unescaped, err := url.QueryUnescape(v.Encode())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(unescaped)

	// Output:
	// address[city]=New York&address[state]=NY&name=Bob
}
