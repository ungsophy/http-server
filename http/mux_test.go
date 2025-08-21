package http_test

import (
	"fmt"
	"testing"

	"github.com/codecrafters-io/http-server-starter-go/http"
)

func TestHandleFunc(t *testing.T) {
	recoverError := func(t *testing.T, aerr any) error {
		if aerr == nil {
			t.Errorf("expected panic, got nil")
		}

		err, ok := aerr.(error)
		if !ok {
			t.Errorf("expected panic to be an error, got %T", aerr)
		}

		return err
	}

	t.Run("invalid pattern", func(t *testing.T) {
		t.Run("pattern does not have http method", func(t *testing.T) {
			defer func() {
				err := recoverError(t, recover())

				expectedMsg := "pattern \"/invalid-pattern\" is invalid. " +
					"pattern should have HTTP method and path. i.e. GET /index"
				if err.Error() != expectedMsg {
					t.Errorf("expected panic message \"%v\", got \"%v\"", expectedMsg, err.Error())
				}
			}()

			mux := http.NewMux()
			mux.HandleFunc("/invalid-pattern", func(req *http.Request, resp *http.Response) {})
		})

		t.Run("pattern has invalid http method", func(t *testing.T) {
			defer func() {
				err := recoverError(t, recover())

				expectedMsg := "\"FOO\" method is invalid. " +
					"method must be one of GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS"
				if err.Error() != expectedMsg {
					t.Errorf("expected panic message \"%v\", got \"%v\"", expectedMsg, err.Error())
				}
			}()

			mux := http.NewMux()
			mux.HandleFunc("FOO /index", func(req *http.Request, resp *http.Response) {})
		})

		t.Run("path does not start with slash", func(t *testing.T) {
			defer func() {
				err := recoverError(t, recover())

				expectedMsg := "\"index\" is invalid. path must start with /"
				if err.Error() != expectedMsg {
					t.Errorf("expected panic message \"%v\", got \"%v\"", expectedMsg, err.Error())
				}
			}()

			mux := http.NewMux()
			mux.HandleFunc("GET index", func(req *http.Request, resp *http.Response) {})
		})
	})

	t.Run("register the same pattern twice", func(t *testing.T) {
		defer func() {
			err := recoverError(t, recover())

			expectedMsg := "route pattern \"GET /index\" already exists"
			if err.Error() != expectedMsg {
				t.Errorf("expected panic message \"%v\", got \"%v\"", expectedMsg, err.Error())
			}
		}()

		mux := http.NewMux()
		mux.HandleFunc("GET /index", func(req *http.Request, resp *http.Response) {})
		mux.HandleFunc("GET /index", func(req *http.Request, resp *http.Response) {})
	})

	t.Run("path params", func(t *testing.T) {
		mux := http.NewMux()
		mux.HandleFunc("GET /files/{filename}", func(req *http.Request, resp *http.Response) {
			resp.StatusCode = 200
			strBody := fmt.Sprintf("File: %s", req.Params["filename"])
			resp.Body = []byte(strBody)
		})

		filename := "test.txt"
		expectedBody := fmt.Sprintf("File: %s", filename)
		req := &http.Request{
			Method: "GET",
			Path:   fmt.Sprintf("/files/%v", filename),
		}
		resp := http.NewResponse()
		mux.HandleRequest(req, resp)

		if resp.StatusCode != 200 {
			t.Errorf("expected status code 200, got %d", resp.StatusCode)
		}

		if string(resp.Body) != expectedBody {
			t.Errorf("expected body \"%v\", got \"%s\"", expectedBody, string(resp.Body))
		}
	})
}

func TestHandleRequest(t *testing.T) {
	t.Run("executes handler that matches route", func(t *testing.T) {
		mux := http.NewMux()
		mux.HandleFunc("GET /index", func(req *http.Request, resp *http.Response) {
			resp.Body = []byte("index")
		})
		mux.HandleFunc("GET /home", func(req *http.Request, resp *http.Response) {
			resp.Body = []byte("home")
		})

		req := &http.Request{
			Method: "GET",
			Path:   "/index",
		}
		resp := http.NewResponse()
		mux.HandleRequest(req, resp)

		if string(resp.Body) != "index" {
			t.Errorf("expected body \"index\", got \"%s\"", string(resp.Body))
		}
	})

	t.Run("returns not found handler when no route matches", func(t *testing.T) {
		mux := http.NewMux()
		req := &http.Request{
			Method: "GET",
			Path:   "/not-found",
		}
		resp := http.NewResponse()
		mux.HandleRequest(req, resp)

		if resp.StatusCode != 404 {
			t.Errorf("expected status code 404, got %d", resp.StatusCode)
		}

		if string(resp.Body) != "Not found" {
			t.Errorf("expected body \"Not found\", got \"%s\"", string(resp.Body))
		}
	})

	t.Run("returns a 200 status code when response status code is not explicitly set", func(t *testing.T) {
		mux := http.NewMux()
		mux.HandleFunc("GET /index", func(req *http.Request, resp *http.Response) {})

		req := &http.Request{
			Method: "GET",
			Path:   "/index",
		}
		resp := http.NewResponse()
		mux.HandleRequest(req, resp)

		if resp.StatusCode != 200 {
			t.Errorf("expected status code 200, got %d", resp.StatusCode)
		}
	})

	t.Run("sets Connection header to close when request has Connection: close", func(t *testing.T) {
		mux := http.NewMux()
		mux.HandleFunc("GET /index", func(req *http.Request, resp *http.Response) {})

		req := &http.Request{
			Method: "GET",
			Path:   "/index",
			Headers: map[string]string{
				"Connection": "close",
			},
		}
		resp := http.NewResponse()
		mux.HandleRequest(req, resp)

		if resp.Headers["Connection"] != "close" {
			t.Errorf("expected Connection header to be 'close', got '%s'", resp.Headers["Connection"])
		}
	})
}
