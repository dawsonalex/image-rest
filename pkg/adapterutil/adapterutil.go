package adapterutil

import (
	"net/http"
)

// Adapter type is a function that takes a http.HandlerFunc as input
// and returns  a http.HandlerFunc. To be used when chaining middleware.
type Adapter func(http.HandlerFunc) http.HandlerFunc

// Adapt takes an http.HandlerFunc and applied all of the Adapters to it in order.
// Useful for chaining handler functions together.
func Adapt(f http.HandlerFunc, middlewares ...Adapter) http.HandlerFunc {
	for _, m := range middlewares {
		f = m(f)
	}
	return f
}
