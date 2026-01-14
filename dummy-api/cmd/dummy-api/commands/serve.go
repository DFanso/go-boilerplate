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
	"google.golang.org/grpc/credentials/insecure"

	"github.com/vidwadeseram/go-boilerplate/dummy-api/gen/dummy"
	dummypb "github.com/vidwadeseram/go-boilerplate/dummy-api/gen/grpc/dummy/pb"
	grpcserver "github.com/vidwadeseram/go-boilerplate/dummy-api/gen/grpc/dummy/server"
	httpserver "github.com/vidwadeseram/go-boilerplate/dummy-api/gen/http/dummy/server"
	"github.com/vidwadeseram/go-boilerplate/dummy-api/internal/auth"
	"github.com/vidwadeseram/go-boilerplate/dummy-api/internal/config"
	db "github.com/vidwadeseram/go-boilerplate/dummy-api/internal/db/sqlc"
	appservice "github.com/vidwadeseram/go-boilerplate/dummy-api/internal/service"
	goahttp "goa.design/goa/v3/http"
	goahttpmiddleware "goa.design/goa/v3/http/middleware"
)

func newServeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start HTTP and gRPC servers",
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

			conn, err := grpc.DialContext(ctx, cfg.IdentityGRPCTarget, grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				return fmt.Errorf("connect to identity grpc: %w", err)
			}
			defer conn.Close()

			identityClient := auth.NewClient(conn)
			queries := db.New(pool)
			svc := appservice.New(logger, queries, identityClient)

			return runServers(ctx, cfg, svc, logger)
		},
	}

	return cmd
}

func runServers(ctx context.Context, cfg *config.Config, svc dummy.Service, logger *slog.Logger) error {
	endpoints := dummy.NewEndpoints(svc)

	mux := goahttp.NewMuxer()
	errHandler := func(ctx context.Context, w http.ResponseWriter, err error) {
		logger.ErrorContext(ctx, "http response error", "error", err)
	}

	server := httpserver.New(endpoints, mux, goahttp.RequestDecoder, goahttp.ResponseEncoder, errHandler, nil, http.Dir("."))
	server.Use(goahttpmiddleware.RequestID())
	server.Mount(mux)

	httpSrv := &http.Server{Addr: cfg.HTTPAddr, Handler: mux}

	grpcSrv := grpc.NewServer()
	dummypb.RegisterDummyServer(grpcSrv, grpcserver.New(endpoints, nil))

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		logger.Info("dummy HTTP server listening", "addr", cfg.HTTPAddr)
		if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	})

	g.Go(func() error {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return httpSrv.Shutdown(shutdownCtx)
	})

	g.Go(func() error {
		lis, err := net.Listen("tcp", cfg.GRPCAddr)
		if err != nil {
			return err
		}
		logger.Info("dummy gRPC server listening", "addr", cfg.GRPCAddr)
		if err := grpcSrv.Serve(lis); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
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
