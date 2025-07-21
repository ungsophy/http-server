package http

type Mux struct{}

func (sm *Mux) HandleFunc(pattern string, handler func(*Request, *Response)) {

}
