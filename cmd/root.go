package cmd

import (
	"fmt"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/spf13/cobra"
)

var (
	cfgFile    string
	progress   bool
	logLevel   string
	since      string
	until      string
)

func use_commit_times(path string) error {
	repo, err := git.PlainOpen(path)
	if err != nil {
		return err
	}

	filemap, err := ls_files(repo)
	if err != nil {
		return err
	}
	err = use_commit_times_log_walk(repo, filemap, nil, nil)
	if err != nil {
		return err
	}
	return nil
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "git-use-commit-times",
	Short: "set commit time timestamp",
	Long: `git-use-commit-times
	Set the file timestamp to the commit datetime.
	`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	RunE: func(cmd *cobra.Command, args []string) error {
		// Initialize logger with log level
		SetLogLevel(logLevel)

		// Execute appropriate function based on flags
		err := use_commit_times(".")
		if err != nil {
			return err
		}
		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.git-use-commit-times.yaml)")
	rootCmd.Flags().BoolVarP(&progress, "progress", "p", false, "Show progressbar.")
	rootCmd.Flags().StringVar(&logLevel, "log-level", "error", "Log level (debug, info, warn, error).")
	rootCmd.Flags().StringVar(&since, "since", "", "Only consider commits after this date (e.g., '2023-01-02').")
	rootCmd.Flags().StringVar(&until, "until", "", "Only consider commits before this date (e.g., '2023-01-02').")
}
