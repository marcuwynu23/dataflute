package serve

import (
	"context"
	"os"
	"os/signal"

	"github.com/marcuwynu23/dataflute/internal/config"
	"github.com/spf13/cobra"
)

var (
	dataDir  string
	filePath string
	endpoint string
	host     string
	port     int
	pretty   bool
)

// rootCmd is the root command of the dataflute CLI.
var rootCmd = &cobra.Command{
	Use:   "dataflute",
	Short: "Expose local data files as read-only REST APIs",
	Long: `dataflute serves local data files (JSON, YAML, CSV, ...) as
read-only REST APIs. Use "serve" to start the HTTP server.`,
}

// serveCmd starts the HTTP server.
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the HTTP server",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		cfg := &config.Config{
			DataDir:  dataDir,
			FilePath: filePath,
			Endpoint: endpoint,
			Host:     host,
			Port:     port,
			Pretty:   pretty,
		}
		if err := cfg.Validate(); err != nil {
			return err
		}
		return run(cmd.Context(), cfg)
	},
}

// Execute runs the root command and blocks until it exits.
func Execute() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	rootCmd.AddCommand(serveCmd)
	return rootCmd.ExecuteContext(ctx)
}

func init() {
	flags := serveCmd.Flags()
	flags.StringVarP(&dataDir, "data", "d", "", "Directory containing supported data files")
	flags.StringVarP(&filePath, "file", "f", "", "Serve a single data file")
	flags.StringVarP(&endpoint, "endpoint", "e", "", "Endpoint for the single file (required when using --file)")
	flags.StringVarP(&host, "host", "H", "0.0.0.0", "Host to bind the HTTP server to")
	flags.IntVarP(&port, "port", "p", 8080, "Port to bind the HTTP server to")
	flags.BoolVarP(&pretty, "pretty", "P", false, "Pretty-print JSON responses")
}
