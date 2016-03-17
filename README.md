railing [![GoDoc](https://godoc.org/github.com/jszwec/railing?status.svg)](http://godoc.org/github.com/jszwec/railing) [![Build Status](https://travis-ci.org/jszwec/railing.svg?branch=master)](https://travis-ci.org/jszwec/railing)
============

Package railing implements encoding and decoding rails style query parameters.
Marshal and Unmarshal functions are based on the Values type which is a wrapper
around url.Values. For details look at the example or GoDoc.

Installation
------------

    go get github.com/jszwec/railing


Example
-----

```
package main

import (
  "fmt"
  "log"
  "net/url"

  "github.com/jszwec/railing"
)

type RGB struct {
  R uint8
  G uint8
  B uint8
}

type Color struct {
  ID   int
  Name string
  RGB  RGB
}

type Colors struct {
  Colors []Color
}

func main() {
  red := Color{
    ID:   1,
    Name: "red",
    RGB:  RGB{255, 0, 0},
  }

  blue := Color{
    ID:   2,
    Name: "blue",
    RGB:  RGB{0, 0, 255},
  }

  colors := Colors{[]Color{red, blue}}

  // Marshal colors.
  values, err := railing.Marshal(&colors)
  if err != nil {
    log.Fatal(err)
  }

  // Print unescaped created query string.
  str, err := url.QueryUnescape(values.Encode())
  if err != nil {
    log.Fatal(err)
  }
  fmt.Println(str)

  // Parse Query to create url.Values.
  urlValues, err := url.ParseQuery(values.Encode())
  if err != nil {
    log.Fatal(err)
  }

  // Unmarshal Values to newColors var.
  var newColors Colors
  if err := railing.Unmarshal(railing.Values{Values: urlValues},
    &newColors); err != nil {
    log.Fatal(err)
  }
  fmt.Println(newColors)
  // Output:
  // Colors[][ID]=1&Colors[][Name]=red&Colors[][RGB][B]=0&Colors[][RGB][G]=0&Colors[][RGB][R]=255&Colors[][ID]=2&Colors[][Name]=blue&Colors[][RGB][B]=255&Colors[][RGB][G]=0&Colors[][RGB][R]=0
  // {[{1 red {255 0 0}} {2 blue {0 0 255}}]}
}
```

Bugs
-----

If you encounter one, post an issue.
