railing [![GoDoc](https://godoc.org/github.com/jszwec/railing?status.svg)](http://godoc.org/github.com/jszwec/railing) [![Build Status](https://travis-ci.org/jszwec/railing.svg?branch=master)](https://travis-ci.org/jszwec/railing) [![Go Report Card](https://goreportcard.com/badge/github.com/jszwec/railing)](https://goreportcard.com/report/github.com/jszwec/railing)
============

Package railing implements encoding and decoding rails style query parameters -
http://guides.rubyonrails.org/action_controller_overview.html#hash-and-array-parameters.
Marshal and Unmarshal functions are based on the Values type which is a wrapper
around url.Values. For details look at the example or GoDoc.

Installation
------------

    go get github.com/jszwec/railing


Example
-----

Lets assume we use rails and we have a following controller. Create method
will just print out params hash.

```ruby
class ColorMapsController < ApplicationController
  def create
    render plain: params.except(:controller, :action).inspect
  end
end
```

Now lets run this little Go program.

```go
package main

import (
	"fmt"
	"image/color"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/jszwec/railing"
)

type ColorMap struct {
	ID      int
	Palette []color.RGBA
}

func main() {
	cm := &ColorMap{
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

	resp, err := http.Post("http://127.0.0.1:3000/color_maps?"+v.Encode(), "", nil)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s", b)
}
```
Output:
```
{"ID"=>"1", "Palette"=>[{"A"=>"0", "B"=>"0", "G"=>"0", "R"=>"255"}, {"A"=>"0", "B"=>"0", "G"=>"255", "R"=>"0"}]}
```

Bugs
-----

If you encounter one, post an issue.
