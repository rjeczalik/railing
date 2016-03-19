package railing_test

import (
	"fmt"
	"log"
	"net/url"

	"github.com/jszwec/railing"
)

type Marshaler struct {
	ID   int
	Name string
	City string
}

func (p *Marshaler) MarshalQuery() (railing.Values, error) {
	return railing.Values{Values: url.Values{
		"marshaler": []string{fmt.Sprintf("%d:%s:%s", p.ID, p.Name, p.City)},
	}}, nil
}

func ExampleMarshal_marshaler() {
	marshaler := Marshaler{1, "Bob", "NY"}
	values, err := railing.Marshal(&marshaler)
	if err != nil {
		log.Fatal(err)
	}
	str, err := url.QueryUnescape(values.Encode())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(str)

	// Output:
	// marshaler=1:Bob:NY
}
