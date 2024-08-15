package migrations

import (
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/operator-framework/operator-registry/alpha/declcfg"
)

type MigrationToken string

const (
	invalidMigration string = ""
	NoMigrations     string = "none"
	AllMigrations    string = "all"
)

type Migration interface {
	Token() MigrationToken
	Help() string
	Migrate(*declcfg.DeclarativeConfig) error
}

func newMigration(token string, help string, fn func(config *declcfg.DeclarativeConfig) error) Migration {
	return &simpleMigration{token: MigrationToken(token), help: help, fn: fn}
}

type simpleMigration struct {
	token MigrationToken
	help  string
	fn    func(*declcfg.DeclarativeConfig) error
}

func (s simpleMigration) Token() MigrationToken {
	return s.token
}

func (s simpleMigration) Migrate(config *declcfg.DeclarativeConfig) error {
	return s.fn(config)
}

func (s simpleMigration) Help() string {
	return s.help
}

type Migrations struct {
	Migrations []Migration
}

// CloneAllMigrations returns a shallow copy of allMigrations
// since slices.Clone is not available in the current golang version
func CloneAllMigrations() []Migration {
	return append(allMigrations[:0:0], allMigrations...)
}

// allMigrations represents the migration catalog
// the order of these migrations is important
var allMigrations = []Migration{
	newMigration(NoMigrations, "do nothing", func(_ *declcfg.DeclarativeConfig) error { return nil }),
	newMigration("bundle-object-to-csv-metadata", `migrates bundles' "olm.bundle.object" to "olm.csv.metadata"`, bundleObjectToCSVMetadata),
}

func NewMigrations(name string) (*Migrations, error) {
	if name == AllMigrations {
		return &Migrations{Migrations: CloneAllMigrations()}, nil
	}

	migrations := CloneAllMigrations()

	found := false
	keep := migrations[:0]
	for _, migration := range migrations {
		keep = append(keep, migration)
		if migration.Token() == MigrationToken(name) {
			found = true
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("unknown migration level %q", name)
	}
	return &Migrations{Migrations: keep}, nil
}

func HelpText() string {
	var help strings.Builder
	help.WriteString("\nThe migrator will run all migrations up to and including the selected level.\n\n")
	help.WriteString("Available migrators:\n")
	if len(allMigrations) == 0 {
		help.WriteString("   (no migrations available in this version)\n")
	}

	tabber := tabwriter.NewWriter(&help, 0, 0, 1, ' ', 0)
	for _, migration := range allMigrations {
		fmt.Fprintf(tabber, "  - %s\t: %s\n", migration.Token(), migration.Help())
	}
	tabber.Flush()
	return help.String()
}

func (m *Migrations) Migrate(config *declcfg.DeclarativeConfig) error {
	for _, migration := range m.Migrations {
		if err := migration.Migrate(config); err != nil {
			return err
		}
	}
	return nil
}
