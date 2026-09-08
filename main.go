package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gatsu420/kisu-be/app/adapter/geminiadapter"
	"github.com/gatsu420/kisu-be/app/adapter/googleauthadapter"
	answerhandlerv1 "github.com/gatsu420/kisu-be/app/handler/answer/v1"
	authhandlerv1 "github.com/gatsu420/kisu-be/app/handler/auth/v1"
	"github.com/gatsu420/kisu-be/app/middleware"
	"github.com/gatsu420/kisu-be/app/repository/bqrepo"
	"github.com/gatsu420/kisu-be/app/repository/pgrepo"
	"github.com/gatsu420/kisu-be/app/repository/staterepo"
	"github.com/gatsu420/kisu-be/app/usecase/answer"
	"github.com/gatsu420/kisu-be/app/usecase/metadata"
	"github.com/gatsu420/kisu-be/common/commonconfig"
	"github.com/gatsu420/kisu-be/common/commonerr"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/genai"
)

func main() {
	err := runServer()
	if err != nil {
		slog.Error(err.Error(),
			slog.Int(commonerr.StatusCodeKey, http.StatusInternalServerError),
			slog.Any(commonerr.ErrKey, err))
		os.Exit(1)
	}
}

func runServer() error {
	envPath := flag.String("env-path", "", "path of env file")
	flag.Parse()

	config, err := commonconfig.NewConfig(*envPath)
	if err != nil {
		return fmt.Errorf("unable to load config: %w", err)
	}

	server, err := createServer(context.Background(), config)
	if err != nil {
		return err
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- server.ListenAndServe()
	}()

	quitCh := make(chan os.Signal, 1)
	signal.Notify(quitCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		if !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("unable to serve incoming connections: %w", err)
		}

		return nil
	case <-quitCh:
		slog.Info("stopping http server")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = server.Shutdown(shutdownCtx)
	if err != nil {
		return fmt.Errorf("http server shutdown: %w", err)
	}

	return nil
}

func createServer(ctx context.Context, config commonconfig.Config) (*http.Server, error) {
	googleAuth := googleauthadapter.NewAdapter(
		config.GoogleAuthClientID,
		config.GoogleAuthClientSecret,
		config.GoogleAuthRedirectURL,
	)

	bqRepo := bqrepo.NewRepository(config.ProjectID, googleAuth)
	pgPool, err := pgxpool.New(ctx, config.PostgresDSN)
	if err != nil {
		return nil, fmt.Errorf("unable to create postgres connection pool: %w", err)
	}
	pgRepo := pgrepo.NewRepository(pgPool)
	stateRepo := staterepo.NewRepository()

	genaiClient, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: config.GeminiApiKey,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create genai client: %w", err)
	}

	metadataUsecase := metadata.NewUsecase(pgRepo, bqRepo)
	geminiAdapter := geminiadapter.NewAdapter(genaiClient, metadataUsecase)
	answerUsecase := answer.NewUsecase(geminiAdapter)

	authHandler := authhandlerv1.NewHandler(googleAuth, metadataUsecase, stateRepo)
	answerHandler := answerhandlerv1.NewHandler(metadataUsecase, answerUsecase)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /auth/v1/get-permission", authHandler.GetPermission)
	mux.HandleFunc("GET /auth/v1/callback", authHandler.Callback)

	refreshToken := middleware.RefreshToken(pgRepo, googleAuth)
	mux.Handle("POST /answer/v1/tool", refreshToken(http.HandlerFunc(answerHandler.AddTool)))
	mux.Handle("GET /answer/v1/answer", refreshToken(http.HandlerFunc(answerHandler.GetAnswer)))

	return &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}, nil
}
