package railing_test

import (
	"fmt"
	"log"
	"net/url"

	"github.com/jszwec/railing"
)

func Example_maps() {
	intMap := map[string]int{
		"first":  1,
		"second": 2,
	}

	// Marshal map
	values, err := railing.Marshal(intMap)
	if err != nil {
		log.Fatal(err)
	}

	// Print unescaped query string.
	str, err := url.QueryUnescape(values.Encode())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(str)

	// clear map
	intMap = make(map[string]int)

	// Parse Query to create url.Values.
	urlValues, err := url.ParseQuery(values.Encode())
	if err != nil {
		log.Fatal(err)
	}

	// Unmarshal query string to a clean map
	if err := railing.Unmarshal(railing.Values{Values: urlValues},
		&intMap); err != nil {
		log.Fatal(err)
	}
	// intsMaps = map[first:1 second:2]

	//
	//
	// lets create map[string][]int now
	//
	//

	intsMap := map[string][]int{
		"first_array":  []int{1, 2},
		"second_array": []int{3, 4},
	}

	// Marshal map
	values, err = railing.Marshal(intsMap)
	if err != nil {
		log.Fatal(err)
	}

	// Print unescaped query string.
	str, err = url.QueryUnescape(values.Encode())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(str)

	// clear map
	intsMap = make(map[string][]int)

	// Parse Query to create url.Values.
	urlValues, err = url.ParseQuery(values.Encode())
	if err != nil {
		log.Fatal(err)
	}

	// Unmarshal query string to a clean map
	if err := railing.Unmarshal(railing.Values{Values: urlValues},
		&intsMap); err != nil {
		log.Fatal(err)
	}

	// intsMaps = map[first_array:[1 2] second_array:[3 4]]

	// Output:
	// first=1&second=2
	// first_array[]=1&first_array[]=2&second_array[]=3&second_array[]=4
}
