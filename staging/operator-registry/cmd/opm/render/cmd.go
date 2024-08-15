package render

import (
	"io"
	"log"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/operator-framework/operator-registry/alpha/action"
	"github.com/operator-framework/operator-registry/alpha/action/migrations"
	"github.com/operator-framework/operator-registry/alpha/declcfg"
	"github.com/operator-framework/operator-registry/cmd/opm/internal/util"
	"github.com/operator-framework/operator-registry/pkg/sqlite"
)

func NewCmd() *cobra.Command {
	var (
		render           action.Render
		output           string

		oldMigrateAllFlag bool
		migrateLevel      string
	)
	cmd := &cobra.Command{
		Use:   "render [index-image | bundle-image | sqlite-file]...",
		Short: "Generate a stream of file-based catalog objects from catalogs and bundles",
		Long: `Generate a stream of file-based catalog objects to stdout from the provided
catalog images, file-based catalog directories, bundle images, and sqlite
database files.

` + sqlite.DeprecationMessage,
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			render.Refs = args

			var write func(declcfg.DeclarativeConfig, io.Writer) error
			switch output {
			case "yaml":
				write = declcfg.WriteYAML
			case "json":
				write = declcfg.WriteJSON
			default:
				log.Fatalf("invalid --output value %q, expected (json|yaml)", output)
			}

			// The bundle loading impl is somewhat verbose, even on the happy path,
			// so discard all logrus default logger logs. Any important failures will be
			// returned from render.Run and logged as fatal errors.
			logrus.SetOutput(io.Discard)

			reg, err := util.CreateCLIRegistry(cmd)
			if err != nil {
				log.Fatal(err)
			}
			defer reg.Destroy()

			render.Registry = reg

			// if the deprecated flag was used, set the level explicitly to the last migration to perform all migrations
			var m *migrations.Migrations
			if oldMigrateAllFlag {
				m, err = migrations.NewMigrations(migrations.AllMigrations)
			} else if migrateLevel != "" {
				m, err = migrations.NewMigrations(migrateLevel)
			}
			if err != nil {
				log.Fatal(err)
			}
			render.Migrations = m

			cfg, err := render.Run(cmd.Context())
			if err != nil {
				log.Fatal(err)
			}

			if err := write(*cfg, os.Stdout); err != nil {
				log.Fatal(err)
			}
		},
	}
	cmd.Flags().StringVarP(&output, "output", "o", "json", "Output format of the streamed file-based catalog objects (json|yaml)")

	cmd.Flags().StringVar(&migrateLevel, "migrate-level", "", "Name of the last migration to run (default: none)\n"+migrations.HelpText())
	cmd.Flags().BoolVar(&oldMigrateAllFlag, "migrate", false, "Perform all available schema migrations on the rendered FBC")
	cmd.MarkFlagsMutuallyExclusive("migrate", "migrate-level")

	cmd.Long += "\n" + sqlite.DeprecationMessage
	return cmd
}

func nullLogger() *logrus.Entry {
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	return logrus.NewEntry(logger)
}
