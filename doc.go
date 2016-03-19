// Package railing implements encoding and decoding rails style query parameters.
//
// See: http://guides.rubyonrails.org/action_controller_overview.html#hash-and-array-parameters.
//
// Marshal and Unmarshal functions are based on the Values type which is a
// wrapper around url.Values. The latter cannot be used because of differences
// in Encode function.
package railing
