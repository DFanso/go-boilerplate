package commands

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"

	identitypb "github.com/vidwadeseram/go-boilerplate/identity-api/gen/grpc/identity/pb"
	grpcserver "github.com/vidwadeseram/go-boilerplate/identity-api/gen/grpc/identity/server"
	httpserver "github.com/vidwadeseram/go-boilerplate/identity-api/gen/http/identity/server"
	"github.com/vidwadeseram/go-boilerplate/identity-api/gen/identity"
	"github.com/vidwadeseram/go-boilerplate/identity-api/internal/config"
	db "github.com/vidwadeseram/go-boilerplate/identity-api/internal/db/sqlc"
	"github.com/vidwadeseram/go-boilerplate/identity-api/internal/security"
	appservice "github.com/vidwadeseram/go-boilerplate/identity-api/internal/service"
	goahttp "goa.design/goa/v3/http"
	goahttpmiddleware "goa.design/goa/v3/http/middleware"
)

func newServeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the HTTP and gRPC servers",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			cfg, err := config.Load()
			if err != nil {
				return err
			}

			logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

			pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
			if err != nil {
				return fmt.Errorf("connect to database: %w", err)
			}
			defer pool.Close()

			queries := db.New(pool)
			tokens := security.NewTokenManager(cfg.JWTSecret, time.Hour)
			svc := appservice.New(logger, queries, tokens)

			return runServers(ctx, cfg, svc, logger)
		},
	}

	return cmd
}

func runServers(ctx context.Context, cfg *config.Config, svc identity.Service, logger *slog.Logger) error {
	endpoints := identity.NewEndpoints(svc)

	hErrHandler := func(ctx context.Context, w http.ResponseWriter, err error) {
		logger.ErrorContext(ctx, "http response error", "error", err)
	}

	mux := goahttp.NewMuxer()
	httpSrv := httpserver.New(endpoints, mux, goahttp.RequestDecoder, goahttp.ResponseEncoder, hErrHandler, nil, http.Dir("."))
	httpSrv.Use(goahttpmiddleware.RequestID())
	httpSrv.Mount(mux)

	httpServer := &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: mux,
	}

	grpcSrv := grpc.NewServer()
	identitypb.RegisterIdentityServer(grpcSrv, grpcserver.New(endpoints, nil))

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		logger.Info("identity HTTP server listening", "addr", cfg.HTTPAddr)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	})

	g.Go(func() error {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return httpServer.Shutdown(shutdownCtx)
	})

	g.Go(func() error {
		lis, err := net.Listen("tcp", cfg.GRPCAddr)
		if err != nil {
			return err
		}
		logger.Info("identity gRPC server listening", "addr", cfg.GRPCAddr)
		if err := grpcSrv.Serve(lis); err != nil {
			return err
		}
		return nil
	})

	g.Go(func() error {
		<-ctx.Done()
		grpcSrv.GracefulStop()
		return nil
	})

	return g.Wait()
}
