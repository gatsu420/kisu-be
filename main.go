package main

import (
	"context"
	"errors"
	"flag"
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
	"github.com/gatsu420/kisu-be/app/llmtool/geminitool"
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
	envPath := flag.String("env-path", "", "path of env file")
	flag.Parse()
	config, err := commonconfig.NewConfig(*envPath)
	if err != nil {
		slog.Error(err.Error(),
			slog.Int(commonerr.StatusCodeKey, http.StatusInternalServerError),
			slog.Any(commonerr.ErrKey, err))
	}

	server := startServer(context.Background(), config)
	quitCh := make(chan os.Signal, 1)
	signal.Notify(quitCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("unable to serve incoming connections",
				slog.Any(commonerr.ErrKey, err),
				slog.Int(commonerr.StatusCodeKey, http.StatusInternalServerError),
			)
		}
	}()

	<-quitCh
	slog.Info("stopping http server")

	timeoutCtx, timeoutCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer timeoutCancel()

	err = server.Shutdown(timeoutCtx)
	if err != nil {
		slog.Error("http server shutdown error",
			slog.Any(commonerr.ErrKey, err),
			slog.Int(commonerr.StatusCodeKey, http.StatusInternalServerError),
		)
	}
}

func startServer(ctx context.Context, config commonconfig.Config) *http.Server {
	googleAuth := googleauthadapter.NewAdapter(config.GoogleAuthClientID, config.GoogleAuthClientSecret, config.GoogleAuthRedirectURL)
	bqRepo := bqrepo.NewRepository(config.ProjectID, googleAuth)

	pgPool, err := pgxpool.New(ctx, config.PostgresDSN)
	if err != nil {
		slog.Error("unable to create postgres connection pool",
			slog.Int(commonerr.StatusCodeKey, http.StatusInternalServerError),
			slog.Any(commonerr.ErrKey, err))
	}
	pgRepo := pgrepo.NewRepository(pgPool)

	geminiTool := geminitool.NewTool(bqRepo)
	geminiToolWiring := geminitool.NewWiring()
	registerGeminiTools(geminiToolWiring, geminiTool)

	genaiClient, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: config.GeminiApiKey,
	})
	if err != nil {
		slog.Error("unable to create genai client",
			slog.Any(commonerr.ErrKey, err),
			slog.Int(commonerr.StatusCodeKey, http.StatusInternalServerError),
		)
	}

	geminiAdapter := geminiadapter.NewAdapter(genaiClient, geminiToolWiring)
	stateRepo := staterepo.NewRepository()

	metadataUsecase := metadata.NewUsecase(pgRepo)
	answerUsecase := answer.NewUsecase(geminiAdapter)

	authHandler := authhandlerv1.NewHandler(googleAuth, metadataUsecase, stateRepo)
	answerHandler := answerhandlerv1.NewHandler(metadataUsecase, answerUsecase)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /auth/v1/get-permission", authHandler.GetPermission)
	mux.HandleFunc("GET /auth/v1/callback", authHandler.Callback)
	authMiddleware := middleware.Chain(
		middleware.RefreshToken(pgRepo, googleAuth),
	)
	mux.Handle("POST /answer/v1/tool", authMiddleware(http.HandlerFunc(answerHandler.AddTool)))
	mux.Handle("GET /answer/v1/answer", authMiddleware(http.HandlerFunc(answerHandler.GetAnswer)))

	return &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
}

func registerGeminiTools(wiring geminitool.Wiring, tool geminitool.Tool) {
	wiring.Add([]geminitool.WiringItem{
		tool.GetSeller(),
	})
}
