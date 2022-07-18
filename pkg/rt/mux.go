package rt

import "net/http"

// Mux is just an interface alternative to *http.ServeMux
type Mux interface {
	HandleFunc(pattern string, handleFunc func(http.ResponseWriter, *http.Request))
}
