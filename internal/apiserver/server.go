package apiserver

import (
	"context"
	"errors"
	"github.com/cappstr/GopherSocial/internal/config"
	"github.com/cappstr/GopherSocial/internal/store"
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"
)

type ApiServer struct {
	config     *config.Config
	logger     *slog.Logger
	store      *store.Store
	jwtManager *JwtManager
}

func New(config *config.Config, logger *slog.Logger, store *store.Store, jwtManager *JwtManager) *ApiServer {
	return &ApiServer{
		config:     config,
		logger:     logger,
		store:      store,
		jwtManager: jwtManager,
	}
}

func (s *ApiServer) Start(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/health", s.healthCheckHandler)
	mux.HandleFunc("POST /v1/auth/signup", s.SignUpHandler)
	mux.HandleFunc("POST /v1/auth/signin", s.SignInHandler)
	mux.HandleFunc("POST /v1/post", s.CreatePostHandler)

	loggingMiddleware := LoggingMiddleware(s.logger)
	authMiddleware := AuthMiddleware(s.jwtManager, s.store.User)
	httpServer := &http.Server{
		Addr:    net.JoinHostPort(s.config.ApiServerHost, s.config.ApiServerAddr),
		Handler: loggingMiddleware(authMiddleware(mux)),
	}

	go func() {
		s.logger.Info("API server started", "listening on:", net.JoinHostPort(s.config.ApiServerHost,
			s.config.ApiServerAddr))
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.logger.Error("error starting http server", "error", err)
		}
	}()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			s.logger.Error("error shutting down http server", "error", err)
		}
	}()
	wg.Wait()

	return nil
}
