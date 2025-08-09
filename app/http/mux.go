package http

import (
	"fmt"
	"net/url"
	"regexp"
	"slices"
	"strings"
)

var (
	methods           = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	pathVariableRegex = regexp.MustCompile(`^{(.+)}$`)
)

type Handler func(*Request, *Response)

type Mux struct {
	handlers map[string]Handler
}

func NewMux() *Mux {
	return &Mux{
		handlers: make(map[string]Handler),
	}
}

func (mux *Mux) HandleReqeust(req *Request, resp *Response) {
	pattern, handler := mux.findHandler(req)
	if len(pattern) == 0 {
		fmt.Println("cannot find handler")
	}

	handler(req, resp)

	if resp.StatusCode == 0 {
		resp.StatusCode = 200
	}

	connection := req.Headers["Connection"]
	if connection == "close" {
		resp.Headers["Connection"] = "close"
	}
}

func (mux *Mux) HandleFunc(pattern string, handler Handler) {
	patternErr := mux.validatePattern(pattern)
	if patternErr != nil {
		panic(pattern)
	}

	_, exists := mux.handlers[pattern]
	if exists {
		panic(fmt.Errorf("route pattern %s already exists", pattern))
	}

	mux.handlers[pattern] = handler
}

func (mux *Mux) findHandler(req *Request) (string, Handler) {
	for pattern, handler := range mux.handlers {
		pattern, params := mux.extractParams(req, pattern)
		if pattern != "" {
			req.Params = params
			return pattern, handler
		}
	}

	return "", notFoundHandler
}

func (mux *Mux) validatePattern(pattern string) error {
	comps := strings.Split(pattern, " ")
	if len(comps) != 2 {
		return fmt.Errorf("pattern is invalid. pattern should have HTTP method and path. i.e. GET /index")
	}

	method := comps[0]
	path := comps[1]

	if !slices.Contains(methods, method) {
		return fmt.Errorf(`"%v" method is invalid. method must be one of %v`, method, strings.Join(methods, ", "))
	}

	if string(path[0]) != "/" {
		return fmt.Errorf(`"%v" is invalid. path must start with /`, path)
	}

	_, parseErr := url.Parse(path)
	if parseErr != nil {
		return fmt.Errorf("path is invalid: %w", parseErr)
	}

	return nil
}

func (mux *Mux) extractParams(req *Request, pattern string) (string, map[string]string) {
	comps := strings.Split(pattern, " ")
	method := comps[0]

	if !strings.EqualFold(method, req.Method) {
		return "", nil
	}

	pathItems := strings.Split(strings.Trim(comps[1], "/"), "/")
	reqPathItems := strings.Split(strings.Trim(req.Path, "/"), "/")
	if len(pathItems) != len(reqPathItems) {
		return "", nil
	}

	var params = make(map[string]string)
	for i, pathItem := range pathItems {
		matches := pathVariableRegex.FindStringSubmatch(pathItem)
		if len(matches) == 2 {
			params[matches[1]] = reqPathItems[i]
		} else if pathItem != reqPathItems[i] {
			return "", nil
		}
	}

	return pattern, params
}

func notFoundHandler(req *Request, resp *Response) {
	resp.StatusCode = 404
	resp.Headers["Content-Type"] = "text/plain"
	resp.Body = []byte("Not found")
}
