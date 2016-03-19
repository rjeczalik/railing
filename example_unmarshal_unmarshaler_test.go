package railing_test

import (
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/jszwec/railing"
)

type Unmarshaler struct {
	Name string
	City string
}

func (u *Unmarshaler) UnmarshalQuery(v railing.Values) error {
	strs := strings.Split(v.Get("unmarshaler"), ":")
	if len(strs) < 2 {
		return fmt.Errorf("error")
	}
	u.Name, u.City = strs[0], strs[1]
	return nil
}

func ExampleUnmarshal_unmarshaler() {
	values := railing.Values{Values: url.Values{
		"unmarshaler": {"Bob:NY"},
	}}

	var u Unmarshaler
	if err := railing.Unmarshal(values, &u); err != nil {
		log.Fatal(err)
	}
	fmt.Println(u)

	// Output:
	// {Bob NY}
}
