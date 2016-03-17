package railing

import (
	"net/url"
	"testing"
)

func TestEncode(t *testing.T) {
	fixtures := []struct {
		in       Values
		expected string
	}{
		// 0
		{
			in:       Values{url.Values{}},
			expected: "",
		},
		// 1
		{
			in: Values{url.Values{
				"foo": []string{"1"},
				"bar": []string{"2"},
			}},
			expected: "bar=2&foo=1",
		},
		// 2
		{
			in: Values{url.Values{
				"array[]":   []string{"1", "2"},
				"foo":       []string{"1"},
				"bar[name]": []string{"name"},
				"bar[id]":   []string{"id"},
			}},
			// "array[]=1&array[]=2&bar[id]=id&bar[name]=name&foo=1"
			expected: "array%5B%5D=1&array%5B%5D=2&bar%5Bid%5D=id&bar%5Bname%5D=nam" +
				"e&foo=1",
		},
		// 3
		{
			in: Values{url.Values{
				"foo[][name]": []string{"a", "b"},
				"foo[][id]":   []string{"1", "2"},
			}},
			// "foo[][id]=1&foo[][name]=a&foo[][id]=2&foo[][name]=b"
			expected: "foo%5B%5D%5Bid%5D=1&foo%5B%5D%5Bname%5D=a&foo%5B%5D%5Bid%5D=" +
				"2&foo%5B%5D%5Bname%5D=b",
		},
		// 4
		{
			in: Values{url.Values{
				"array[][a][][a]": []string{"2", "1"},
				"array[][a][][z]": []string{"z", "a"},
			}},
			//  "[a][][a]=2&array[][a][][z]=z&array[][a][][a]=1&array[][a][][z]=a"
			expected: "array%5B%5D%5Ba%5D%5B%5D%5Ba%5D=2&array%5B%5D%5Ba%5D%5B%5D%5" +
				"Bz%5D=z&array%5B%5D%5Ba%5D%5B%5D%5Ba%5D=1&array%5B%5D%5Ba%5D%5B%5D%5" +
				"Bz%5D=a",
		},
		// 5
		{
			in: Values{url.Values{
				"a":                           []string{"val"},
				"array[]":                     []string{"a", "b"},
				"z[]":                         []string{"1", "2"},
				"a_b":                         []string{"val"},
				"foo[][id]":                   []string{"1", "2"},
				"obj1[id]":                    []string{"id"},
				"obj1[obj2][id]":              []string{"id"},
				"obj1[obj2][obj3][obj4][][b]": []string{"2", "2"},
				"obj1[obj2][obj3][objs][]":    []string{"1", "2"},
				"obj1[obj2][obj3][name][]":    []string{"name", "name"},
				"obj1[obj2][obj3][obj4][][a]": []string{"1", "1"},
			}},
			expected: "a=val&a_b=val&array%5B%5D=a&array%5B%5D=b&foo%5B%5D%5Bid%5D=" +
				"1&foo%5B%5D%5Bid%5D=2&obj1%5Bid%5D=id&obj1%5Bobj2%5D%5Bid%5D=id&obj1" +
				"%5Bobj2%5D%5Bobj3%5D%5Bname%5D%5B%5D=name&obj1%5Bobj2%5D%5Bobj3%5D%5" +
				"Bname%5D%5B%5D=name&obj1%5Bobj2%5D%5Bobj3%5D%5Bobj4%5D%5B%5D%5Ba%5D=" +
				"1&obj1%5Bobj2%5D%5Bobj3%5D%5Bobj4%5D%5B%5D%5Bb%5D=2&obj1%5Bobj2%5D%5" +
				"Bobj3%5D%5Bobj4%5D%5B%5D%5Ba%5D=1&obj1%5Bobj2%5D%5Bobj3%5D%5Bobj4%5D" +
				"%5B%5D%5Bb%5D=2&obj1%5Bobj2%5D%5Bobj3%5D%5Bobjs%5D%5B%5D=1&obj1%5Bob" +
				"j2%5D%5Bobj3%5D%5Bobjs%5D%5B%5D=2&z%5B%5D=1&z%5B%5D=2",
		},
	}
	for i, fixture := range fixtures {
		if out := fixture.in.Encode(); out != fixture.expected {
			t.Errorf("expected %s; got %s (i=%d)", fixture.expected, out, i)
		}
	}
}
