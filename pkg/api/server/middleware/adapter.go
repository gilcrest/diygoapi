package middleware

import (
	"net/http"
)

// Adapter type (it gets its name from the adapter pattern — also known as the
// decorator pattern) above is a function that both takes in and returns an
// http.Handler. This is the essence of the wrapper; we will pass in an existing
// http.Handler, the Adapter will adapt it, and return a new (probably wrapped)
// http.Handler for us to use in its place. So far this is not much different
// from just wrapping http.HandlerFunc types, however, now, we can instead write
//  functions that themselves return an Adapter.
type Adapter func(http.Handler) http.Handler

// Adapt function takes the handler you want to adapt, and a list of our
// Adapter types. The result of our Notify function is an acceptable Adapter.
// Our Adapt function will simply iterate over all adapters,
// calling them one by one (in reverse order) in a chained manner,
// returning the result of the first adapter. - Mat Ryer @matryer
func Adapt(h http.Handler, adapters ...Adapter) http.Handler {
	for _, adapter := range adapters {
		h = adapter(h)
	}
	return h
}
