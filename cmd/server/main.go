package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/geezyx/sudo-server/internal/actions/wol"
	"github.com/geezyx/sudo-server/internal/core/services/command"
	"github.com/geezyx/sudo-server/internal/core/services/config"
	commandhandler "github.com/geezyx/sudo-server/internal/handlers/command"
	"github.com/go-logr/logr"
	"github.com/go-logr/stdr"
)

type middleware func(http.Handler) http.Handler
type middlewares []middleware

func (mws middlewares) apply(hdlr http.Handler) http.Handler {
	if len(mws) == 0 {
		return hdlr
	}
	return mws[1:].apply(mws[0](hdlr))
}

func (c *controller) shutdown(ctx context.Context, server *http.Server) context.Context {
	ctx, done := context.WithCancel(ctx)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		defer done()

		<-quit
		signal.Stop(quit)
		close(quit)

		atomic.StoreInt64(&c.healthy, 0)
		server.ErrorLog.Printf("Server is shutting down...\n")

		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		server.SetKeepAlivesEnabled(false)
		if err := server.Shutdown(ctx); err != nil {
			server.ErrorLog.Fatalf("Could not gracefully shutdown the server: %s\n", err)
		}
	}()

	return ctx
}

type controller struct {
	logger        logr.Logger
	nextRequestID func() string
	healthy       int64
	apiKey        string
}

func main() {
	listenAddr := ":8090"
	if len(os.Args) == 2 {
		listenAddr = os.Args[1]
	}

	stdLogger := log.New(os.Stdout, "http: ", log.LstdFlags)
	logger := stdr.New(stdLogger)
	logger.Info("Server is starting...")

	apiKey := getAPIKey()

	cfg, err := config.New().Load("config.yaml")
	if err != nil {
		fmt.Println("error loading config.yaml")
		os.Exit(1)
	}

	c := &controller{
		logger:        logger,
		nextRequestID: func() string { return strconv.FormatInt(time.Now().UnixNano(), 36) },
		apiKey:        apiKey,
	}

	cmd := command.New()

	wolAction := wol.New(cfg.WOL.MACAddress)
	for _, c := range cfg.WOL.Commands {
		cmd.Add(c, wolAction.WakeUp)
	}

	h := commandhandler.NewHTTPHandler(cmd, logger)

	router := http.NewServeMux()
	router.Handle("/sudo", h)
	router.HandleFunc("/healthz", c.healthz)

	server := &http.Server{
		Addr:         listenAddr,
		Handler:      (middlewares{c.tracing, c.logging, c.authorization}).apply(router),
		ErrorLog:     stdLogger,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}
	ctx := c.shutdown(context.Background(), server)

	logger.Info("server is ready to handle requests", "listener", listenAddr)
	atomic.StoreInt64(&c.healthy, time.Now().UnixNano())

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		logger.Error(err, "could not listen", "listener", listenAddr)
	}
	<-ctx.Done()
	logger.Info("server stopped")
}

func (c *controller) healthz(w http.ResponseWriter, req *http.Request) {
	if h := atomic.LoadInt64(&c.healthy); h == 0 {
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		fmt.Fprintf(w, "uptime: %s\n", time.Since(time.Unix(0, h)))
	}
}

func (c *controller) logging(hdlr http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		defer func(start time.Time) {
			requestID := w.Header().Get("X-Request-Id")
			if requestID == "" {
				requestID = "unknown"
			}
			c.logger.Info("request", "id", requestID, "method", req.Method, "path", req.URL.Path, "remote", req.RemoteAddr, "user agent", req.UserAgent(), "duration", time.Since(start))
		}(time.Now())
		hdlr.ServeHTTP(w, req)
	})
}

func (c *controller) tracing(hdlr http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		requestID := req.Header.Get("X-Request-Id")
		if requestID == "" {
			requestID = c.nextRequestID()
		}
		w.Header().Set("X-Request-Id", requestID)
		hdlr.ServeHTTP(w, req)
	})
}

func (c *controller) authorization(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("apikey")
		if apiKey != c.apiKey {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return // don't call original handler
		}
		h.ServeHTTP(w, r)
	})
}

func getAPIKey() string {
	apiKey := flag.String("api-key", "", "api key for validating requests")
	if *apiKey == "" {
		v := os.Getenv("SUDO_API_KEY")
		apiKey = &v
	}
	if *apiKey == "" {
		fmt.Println("must provide --api-key or set SUDO_API_KEY environment variable")
		os.Exit(2)
	}
	return *apiKey
}
