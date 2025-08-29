package app

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/codecrafters-io/http-server-starter-go/http"
)

type Config struct {
	Directory string
	Logger    *slog.Logger
	Port      int
}

type App struct {
	mux    *http.Mux
	server *http.Server

	Config            *Config
	HTTPServerCreated chan bool
}

func NewApp(config *Config) *App {
	app := &App{
		Config:            config,
		HTTPServerCreated: make(chan bool, 1),
	}

	mux := http.NewMux(config.Logger)
	mux.HandleFunc("GET /", app.homeHandler)
	mux.HandleFunc("GET /user-agent", app.getUserAgentHandler)
	mux.HandleFunc("GET /echo/{str}", app.echoHandler)
	mux.HandleFunc("GET /files/{filename}", app.readFileHandler)
	mux.HandleFunc("POST /files/{filename}", app.createFileHandler)

	server, err := http.NewServer(fmt.Sprintf(":%v", config.Port), mux, config.Logger)
	if err != nil {
		config.Logger.Error("cannot create HTTP server", "error", err)
		os.Exit(1)
	}

	app.mux = mux
	app.server = server

	return app
}

func (a *App) Start() error {
	go func() {
		a.HTTPServerCreated <- <-a.server.Created
	}()

	err := a.server.Start()
	if err != nil {
		return fmt.Errorf("cannot start HTTP server: %w", err)
	}

	return nil
}

func (a *App) Stop() error {
	err := a.server.Stop()
	if err != nil {
		return fmt.Errorf("cannot stop HTTP server: %w", err)
	}

	return nil
}

func (a *App) homeHandler(req *http.Request, resp *http.Response) {
	resp.StatusCode = 200
}

func (a *App) getUserAgentHandler(req *http.Request, resp *http.Response) {
	resp.StatusCode = 200
	resp.Headers["Content-Type"] = "text/plain"
	resp.Body = []byte(req.Headers["User-Agent"])
}

func (a *App) echoHandler(req *http.Request, resp *http.Response) {
	resp.StatusCode = 200
	resp.Headers["Content-Type"] = "text/plain"
	resp.Body = []byte(req.Params["str"])

	if strings.Contains(req.Headers["Accept-Encoding"], "gzip") {
		body, err := gzipCompress(resp.Body)
		if err != nil {
			a.Config.Logger.Error("cannot gzip response", "error", err, "str", req.Params["str"])
			resp.StatusCode = 500
			resp.Body = []byte("cannot gzip")
			return
		}

		resp.Headers["Content-Encoding"] = "gzip"
		resp.Body = body
	}
}

func (a *App) readFileHandler(req *http.Request, resp *http.Response) {
	filepath := filepath.Join(a.Config.Directory, req.Params["filename"])
	file, openErr := os.Open(filepath)
	if openErr != nil {
		a.Config.Logger.Warn("cannot open file", "error", openErr)
		resp.StatusCode = 404
		return
	}
	defer file.Close()

	body, readErr := io.ReadAll(file)
	if readErr != nil {
		a.Config.Logger.Error("error reading file", "error", readErr, "filepath", filepath)
		resp.StatusCode = 500
		resp.Body = []byte("cannot read from file")
		return
	}

	resp.StatusCode = 200
	resp.Headers["Content-Type"] = "application/octet-stream"
	resp.Body = body
}

func (a *App) createFileHandler(req *http.Request, resp *http.Response) {
	filepath := filepath.Join(a.Config.Directory, req.Params["filename"])
	file, createErr := os.Create(filepath)
	if createErr != nil {
		a.Config.Logger.Error("error creating file", "error", createErr)
		resp.StatusCode = 500
		resp.Body = []byte("cannot create file")
		return
	}
	defer file.Close()

	_, writeErr := file.Write(req.Body)
	if writeErr != nil {
		a.Config.Logger.Error("error writing file", "error", writeErr)
		resp.StatusCode = 500
		resp.Body = []byte("cannot write to file")
		return
	}

	resp.StatusCode = 201
}

func gzipCompress(data []byte) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	zw := gzip.NewWriter(buf)

	_, writeErr := zw.Write(data)
	if writeErr != nil {
		return nil, writeErr
	}
	zw.Flush()
	// Close gzip writer before reading from buffer
	// to make sure that gzip footer is written to the buffer
	zw.Close()

	return buf.Bytes(), nil
}
