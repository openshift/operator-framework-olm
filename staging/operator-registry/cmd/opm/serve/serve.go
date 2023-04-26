package serve

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/operator-framework/operator-registry/pkg/api"
	health "github.com/operator-framework/operator-registry/pkg/api/grpc_health_v1"
	"github.com/operator-framework/operator-registry/pkg/cache"
	"github.com/operator-framework/operator-registry/pkg/lib/dns"
	"github.com/operator-framework/operator-registry/pkg/lib/graceful"
	"github.com/operator-framework/operator-registry/pkg/lib/log"
	"github.com/operator-framework/operator-registry/pkg/server"
)

type serve struct {
	configDir             string
	cacheDir              string
	cacheOnly             bool
	cacheEnforceIntegrity bool

	port           string
	terminationLog string
	debug          bool

	logger *logrus.Entry
}

func NewCmd() *cobra.Command {
	logger := logrus.New()
	s := serve{
		logger: logrus.NewEntry(logger),
	}
	cmd := &cobra.Command{
		Use:   "serve <source_path>",
		Short: "serve declarative configs",
		Long: `This command serves declarative configs via a GRPC server.

NOTE: The declarative config directory is loaded by the serve command at
startup. Changes made to the declarative config after the this command starts
will not be reflected in the served content.
`,
		Args: cobra.ExactArgs(1),
		PreRun: func(_ *cobra.Command, args []string) {
			s.configDir = args[0]
			if s.debug {
				logger.SetLevel(logrus.DebugLevel)
			}
		},
		Run: func(cmd *cobra.Command, _ []string) {
			if !cmd.Flags().Changed("cache-enforce-integrity") {
				s.cacheEnforceIntegrity = s.cacheDir != "" && !s.cacheOnly
			}
			if err := s.run(cmd.Context()); err != nil {
				logger.Fatal(err)
			}
		},
	}

	cmd.Flags().BoolVar(&s.debug, "debug", false, "enable debug logging")
	cmd.Flags().StringVarP(&s.port, "port", "p", "50051", "port number to serve on")
	cmd.Flags().StringVarP(&s.terminationLog, "termination-log", "t", "/dev/termination-log", "path to a container termination log file")
	cmd.Flags().StringVar(&s.cacheDir, "cache-dir", "", "if set, sync and persist server cache directory")
	cmd.Flags().BoolVar(&s.cacheOnly, "cache-only", false, "sync the serve cache and exit without serving")
	cmd.Flags().BoolVar(&s.cacheEnforceIntegrity, "cache-enforce-integrity", false, "exit with error if cache is not present or has been invalidated. (default: true when --cache-dir is set and --cache-only is false, false otherwise), ")
	return cmd
}

func (s *serve) run(ctx context.Context) error {
	// Immediately set up termination log
	err := log.AddDefaultWriterHooks(s.terminationLog)
	if err != nil {
		s.logger.WithError(err).Warn("unable to set termination log path")
	}

	// Ensure there is a default nsswitch config
	if err := dns.EnsureNsswitch(); err != nil {
		s.logger.WithError(err).Warn("unable to write default nsswitch config")
	}

	s.logger = s.logger.WithFields(logrus.Fields{"configs": s.configDir, "port": s.port})

	if s.cacheDir == "" && s.cacheEnforceIntegrity {
		return fmt.Errorf("--cache-dir must be specified with --cache-enforce-integrity")
	}

	if s.cacheDir == "" {
		s.cacheDir, err = os.MkdirTemp("", "opm-serve-cache-")
		if err != nil {
			return err
		}
		defer os.RemoveAll(s.cacheDir)
	}

	store, err := cache.New(s.cacheDir)
	if err != nil {
		return err
	}
	if storeCloser, ok := store.(io.Closer); ok {
		defer storeCloser.Close()
	}
	if s.cacheEnforceIntegrity {
		if err := store.CheckIntegrity(os.DirFS(s.configDir)); err != nil {
			return err
		}
		if err := store.Load(); err != nil {
			return err
		}
	} else {
		if err := cache.LoadOrRebuild(store, os.DirFS(s.configDir)); err != nil {
			return err
		}
	}

	if s.cacheOnly {
		return nil
	}

	lis, err := net.Listen("tcp", ":"+s.port)
	if err != nil {
		return fmt.Errorf("failed to listen: %s", err)
	}

	grpcServer := grpc.NewServer()
	api.RegisterRegistryServer(grpcServer, server.NewRegistryServer(store))
	health.RegisterHealthServer(grpcServer, server.NewHealthServer())
	reflection.Register(grpcServer)
	s.logger.Info("serving registry")
	return graceful.Shutdown(s.logger, func() error {
		return grpcServer.Serve(lis)
	}, func() {
		grpcServer.GracefulStop()
	})
}
