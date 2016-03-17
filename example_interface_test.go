package railing_test

import (
	"fmt"
	"log"
	"net/url"

	"github.com/jszwec/railing"
)

func Example_interface() {
	m := map[string]interface{}{
		"name": "Bob",
		"address": map[string]interface{}{
			"city":  "New York",
			"state": "NY",
		},
	}

	// Marshal map
	values, err := railing.Marshal(m)
	if err != nil {
		log.Fatal(err)
	}

	// Print unescaped query string.
	str, err := url.QueryUnescape(values.Encode())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(str)

	var v interface{}

	// Parse Query to create url.Values.
	urlValues, err := url.ParseQuery(values.Encode())
	if err != nil {
		log.Fatal(err)
	}

	// Unmarshal query string to a clean map
	if err := railing.Unmarshal(railing.Values{Values: urlValues},
		&v); err != nil {
		log.Fatal(err)
	}

	// v = map[address:map[state:[NY] city:[New York]] name:[Bob]]

	// Output:
	// address[city]=New York&address[state]=NY&name=Bob

}
