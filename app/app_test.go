package app_test

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptrace"
	"os"
	"testing"

	"github.com/codecrafters-io/http-server-starter-go/app"
)

func TestHandlers(t *testing.T) {
	cfg := &app.Config{
		Directory: "./../testdata",
		Port:      8181,
		Logger:    slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	testApp := app.NewApp(cfg)

	go func() {
		err := testApp.Start()
		if err != nil && !errors.Is(err, net.ErrClosed) {
			t.Errorf("failed to start ap	p: %v", err)
		}
	}()

	// waiting for HTTP server to properly created before sending requests
	<-testApp.HTTPServerCreated

	t.Cleanup(func() {
		err := testApp.Stop()
		if err != nil {
			t.Errorf("failed to stop app: %v", err)
		}
	})

	testCases := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
		expectedBody   []byte
	}{
		{
			name:           "GET /not-found",
			method:         "GET",
			path:           "/not-found",
			expectedStatus: http.StatusNotFound,
			expectedBody:   []byte("Not found"),
		},
		{
			name:           "GET /home",
			method:         "GET",
			path:           "/",
			expectedStatus: http.StatusOK,
			expectedBody:   []byte{},
		},
		{
			name:           "GET /user-agent",
			method:         "GET",
			path:           "/user-agent",
			expectedStatus: http.StatusOK,
			expectedBody:   []byte("Go-http-client/1.1"),
		},
		{
			name:           "GET /echo/foo",
			method:         "GET",
			path:           "/echo/foo",
			expectedStatus: http.StatusOK,
			expectedBody:   []byte("foo"),
		},
		{
			name:           "GET /echo/bar",
			method:         "GET",
			path:           "/echo/bar",
			expectedStatus: http.StatusOK,
			expectedBody:   []byte("bar"),
		},
		{
			name:           "GET /files/test",
			method:         "GET",
			path:           "/files/test",
			expectedStatus: http.StatusOK,
			expectedBody:   []byte("abc"),
		},
		{
			name:           "GET /files/not-found",
			method:         "GET",
			path:           "/files/not-found",
			expectedStatus: http.StatusNotFound,
			expectedBody:   []byte(""),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(tt *testing.T) {
			req := request{
				method: tc.method,
				url:    fmt.Sprintf("http://localhost:%d%v", cfg.Port, tc.path),
				body:   nil,
			}
			resp, err := sendRequest(context.Background(), req)
			if err != nil {
				tt.Errorf("failed to send request: %v", err)
			}

			if tc.expectedStatus != resp.status {
				t.Fatalf("unexpected %v status code but got %v", tc.expectedStatus, resp.status)
			}

			if !bytes.Equal(resp.body, tc.expectedBody) {
				tt.Fatalf("unexpected response body: got %q, want %q", resp.body, tc.expectedBody)
			}
		})
	}

	t.Run("POST /files/new-file", func(tt *testing.T) {
		filename := "hello-world"

		err := os.Remove(fmt.Sprintf("./../testdata/%v", filename))
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			tt.Fatalf("failed to remove file: %v", err)
		}

		ctx := context.Background()
		body := []byte("Hello, World!")
		url := fmt.Sprintf("http://localhost:%d/files/%v", cfg.Port, filename)

		req1 := request{
			method: http.MethodPost,
			url:    url,
			body:   bytes.NewBuffer(body),
		}
		resp1, err := sendRequest(ctx, req1)
		if err != nil {
			tt.Fatalf("failed to send request: %v", err)
		}

		if resp1.status != http.StatusCreated {
			tt.Fatalf("unexpected status code: got %v, want %v", resp1.status, http.StatusCreated)
		}

		req2 := request{
			method: http.MethodGet,
			url:    url,
		}
		resp2, err := sendRequest(ctx, req2)
		if err != nil {
			tt.Fatalf("failed to send request: %v", err)
		}

		if resp2.status != http.StatusOK {
			tt.Fatalf("unexpected status code: got %v, want %v", resp2.status, http.StatusOK)
		}

		if !bytes.Equal(resp2.body, body) {
			tt.Fatalf("unexpected response body: got %q, want %q", resp2.body, body)
		}

		err = os.Remove(fmt.Sprintf("./../testdata/%v", filename))
		if err != nil {
			tt.Fatalf("failed to remove file: %v", err)
		}
	})

	t.Run("Test re-use connection", func(tt *testing.T) {
		// making sure that current connection is properly closed before running actual test
		req := request{
			method: http.MethodGet,
			url:    fmt.Sprintf("http://localhost:%d/echo/reset", cfg.Port),
			headers: map[string][]string{
				"Connection": {"close"},
			},
		}

		resp, err := sendRequest(context.Background(), req)
		if err != nil {
			tt.Fatalf("failed to send request: %v", err)
		}

		if resp.status != http.StatusOK {
			tt.Fatalf("unexpected status code: got %v, want %v", resp.status, http.StatusOK)
		}

		reusedCollection := make([]bool, 0)
		clientTrace := httptrace.ClientTrace{
			GotConn: func(info httptrace.GotConnInfo) {
				reusedCollection = append(reusedCollection, info.Reused)
			},
		}
		ctx := httptrace.WithClientTrace(context.Background(), &clientTrace)

		for i := 0; i <= 2; i++ {
			url := fmt.Sprintf("http://localhost:%d/echo/%v", cfg.Port, i)
			req := request{
				method: http.MethodGet,
				url:    url,
			}

			// telling server to close connection after responding to the request
			if i == 1 {
				req.headers = map[string][]string{
					"Connection": {"close"},
				}
			}

			resp, err := sendRequest(ctx, req)
			if err != nil {
				tt.Fatalf("failed to send request: %v", err)
			}

			if resp.status != http.StatusOK {
				tt.Fatalf("unexpected status code: got %v, want %v", resp.status, http.StatusOK)
			}

			expectedBody := fmt.Sprintf("%v", i)
			if string(resp.body) != expectedBody {
				tt.Errorf("unexpected response body: got %q, want %q", resp.body, expectedBody)
			}

			if i == 1 {
				if !reusedCollection[i] {
					tt.Errorf("expected connection to be reused, but it was not")
				}
			} else {
				if reusedCollection[i] {
					tt.Errorf("expected connection to not be reused, but it was")
				}
			}
		}
	})

	t.Run("Test gzip", func(tt *testing.T) {
		str := "foobar123"
		req := request{
			method: http.MethodGet,
			url:    fmt.Sprintf("http://localhost:%d/echo/%s", cfg.Port, str),
			headers: map[string][]string{
				"Accept-Encoding": {"gzip"},
			},
		}

		resp, err := sendRequest(context.Background(), req)
		if err != nil {
			tt.Fatalf("failed to send request: %v", err)
		}

		if resp.status != http.StatusOK {
			tt.Fatalf("unexpected status code: got %v, want %v", resp.status, http.StatusOK)
		}

		gzipReader, err := gzip.NewReader(bytes.NewReader(resp.body))
		if err != nil {
			tt.Fatalf("failed to create gzip reader: %v", err)
		}
		defer gzipReader.Close()

		decompressedBody, err := io.ReadAll(gzipReader)
		if err != nil {
			tt.Fatalf("failed to read decompressed body: %v", err)
		}

		if string(decompressedBody) != str {
			tt.Errorf("unexpected response body: got %q, want %q", decompressedBody, str)
		}
	})
}

type request struct {
	method  string
	url     string
	body    io.Reader
	headers map[string][]string
}

type response struct {
	status  int
	body    []byte
	headers map[string][]string
}

func sendRequest(ctx context.Context, req request) (*response, error) {
	r, err := http.NewRequestWithContext(ctx, req.method, req.url, req.body)
	r.Header = req.headers
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	return &response{
		status:  resp.StatusCode,
		body:    respBody,
		headers: resp.Header,
	}, nil
}
