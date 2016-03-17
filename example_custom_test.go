package railing_test

import (
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"

	"github.com/jszwec/railing"
)

type Person struct {
	ID   int
	Name string
	City string
}

func (p *Person) MarshalQuery() (railing.Values, error) {
	return railing.Values{Values: url.Values{
		"person": []string{fmt.Sprintf("%d:%s:%s", p.ID, p.Name, p.City)},
	}}, nil
}

func (p *Person) UnmarshalQuery(v railing.Values) error {
	str := v.Get("person")
	strs := strings.Split(str, ":")
	if len(strs) < 3 {
		return fmt.Errorf("error")
	}
	n, err := strconv.Atoi(strs[0])
	if err != nil {
		return err
	}
	p.ID = n
	p.Name = strs[1]
	p.City = strs[2]
	return nil
}

func Example_custom() {
	person := Person{1, "Bob", "NY"}
	values, err := railing.Marshal(&person)
	if err != nil {
		log.Fatal(err)
	}
	str, err := url.QueryUnescape(values.Encode())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(str)

	var newPerson Person

	// Parse Query to create url.Values.
	urlValues, err := url.ParseQuery(values.Encode())
	if err != nil {
		log.Fatal(err)
	}

	// Unmarshal query string to a clean map
	if err := railing.Unmarshal(railing.Values{Values: urlValues},
		&newPerson); err != nil {
		log.Fatal(err)
	}
	fmt.Println(person)

	// Output:
	// person=1:Bob:NY
	// {1 Bob NY}
}
