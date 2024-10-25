package request

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/echsylon/go-log"
)

type HttpServer interface {
	Handle(string, func(map[string][]string, []byte) ([]byte, int))
	Serve() error
	Stop() error
}

type server struct {
	multiplexer *http.ServeMux
	server      *http.Server
}

var KeyServerAddress = struct{}{}

func NewHttpServer(baseContext context.Context, port int) HttpServer {
	serverAddress := fmt.Sprintf("localhost:%d", port)
	multiplexer := http.NewServeMux()

	httpServer := &http.Server{
		Addr:    serverAddress,
		Handler: multiplexer,
		BaseContext: func(requestListener net.Listener) context.Context {
			serverAddressValue := requestListener.Addr().String()
			return context.WithValue(baseContext, KeyServerAddress, serverAddressValue)
		},
	}

	return &server{
		multiplexer: multiplexer,
		server:      httpServer,
	}
}

func (s *server) Handle(path string, do func(map[string][]string, []byte) ([]byte, int)) {
	regex := regexp.MustCompile(`\{([a-z0-9]+)\}`)
	keys := regex.FindAllStringSubmatch(path, -1)

	s.multiplexer.HandleFunc(path, func(response http.ResponseWriter, request *http.Request) {
		params := make(map[string][]string)
		parsePathSegments(keys, request, &params)
		parseQueryAndFormValues(request, &params)

		body, _ := io.ReadAll(request.Body)
		result, status := do(params, body)

		length := strconv.Itoa(len(result))
		response.Header().Add("Content-Type", "application/json")
		response.Header().Add("Content-Length", length)
		response.WriteHeader(status)
		response.Write(result)
	})
}

func (s *server) Serve() error {
	// Blocks until err or close.
	if err := s.server.ListenAndServe(); err != http.ErrServerClosed {
		log.Error("HTTP Server closed unexpectedly: %s\n", err.Error())
		return err
	} else {
		return nil
	}
}

func (s *server) Stop() error {
	ctx, abort := context.WithTimeout(context.Background(), 10*time.Second)
	defer abort()

	// Shutdown will block until it has gracefully closed all connections
	// notified all listeners and disposed of all internal resources. Our
	// context will, however, only allow it a set amount of time for this.
	if err := s.server.Shutdown(ctx); err != nil {
		log.Warning("HTTP Server closed harshly")
		return err
	} else {
		log.Information("HTTP Server shut down gracefully")
		return nil
	}
}

// Helper functions
func parsePathSegments(keys [][]string, request *http.Request, params *map[string][]string) {
	if params == nil {
		return
	}

	for _, match := range keys {
		key := match[1]
		value := request.PathValue(key)
		values, ok := (*params)[key]

		if !ok || len(values) == 0 {
			(*params)[key] = []string{value}
		} else {
			(*params)[key] = append(values, value)
		}
	}
}

func parseQueryAndFormValues(request *http.Request, params *map[string][]string) {
	if params == nil {
		return
	}

	if err := request.ParseForm(); err == nil {
		for key, values := range request.Form {
			if len(values) == 0 {
				continue
			}

			existingValues, ok := (*params)[key]
			if !ok || len(existingValues) == 0 {
				(*params)[key] = values
			} else {
				(*params)[key] = append(existingValues, values...)
			}
		}
	}
}
