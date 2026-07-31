package serve

import (
	"context"
	"os"
	"os/signal"

	"github.com/marcuwynu23/dataserve/internal/config"
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

// rootCmd is the root command of the dataserve CLI.
var rootCmd = &cobra.Command{
	Use:   "dataserve",
	Short: "Expose local data files as read-only REST APIs",
	Long: `dataserve serves local data files (JSON, YAML, CSV, ...) as
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
	flags.StringVar(&dataDir, "data", "", "Directory containing supported data files")
	flags.StringVar(&filePath, "file", "", "Serve a single data file")
	flags.StringVar(&endpoint, "endpoint", "", "Endpoint for the single file (required when using --file)")
	flags.StringVar(&host, "host", "0.0.0.0", "Host to bind the HTTP server to")
	flags.IntVar(&port, "port", 8080, "Port to bind the HTTP server to")
	flags.BoolVar(&pretty, "pretty", false, "Pretty-print JSON responses")
}
