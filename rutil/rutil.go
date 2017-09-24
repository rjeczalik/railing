package rutil

import (
	"net/url"

	"github.com/jszwec/railing"
)

func ParseURL(link string) (railing.Values, error) {
	u, err := url.Parse(link)
	if err != nil {
		return railing.Values{}, err
	}
	return railing.Values{Values: u.Query()}, nil
}

func UnmarshalURL(link string, v interface{}) error {
	m, err := ParseURL(link)
	if err != nil {
		return err
	}
	return railing.Unmarshal(m, v)
}
