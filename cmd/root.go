package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/spf13/cobra"
)

var (
	logLevel   string
	since      string
	until      string
)

func use_commit_times(path string) error {
	repo, err := git.PlainOpen(path)
	if err != nil {
		return err
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return err
	}
	workdir := worktree.Filesystem.Root()

	filemap, err := ls_files(workdir)
	if err != nil {
		return err
	}
	
	// Parse time parameters
	var sinceTime, untilTime *time.Time
	if since != "" {
		t, err := time.Parse("2006-01-02", since)
		if err != nil {
			return fmt.Errorf("invalid since date format: %w", err)
		}
		sinceTime = &t
	}
	if until != "" {
		t, err := time.Parse("2006-01-02", until)
		if err != nil {
			return fmt.Errorf("invalid until date format: %w", err)
		}
		untilTime = &t
	}
	
	return use_commit_times_walk(workdir, filemap, sinceTime, untilTime)
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
		return use_commit_times(".")
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
	rootCmd.Flags().StringVar(&logLevel, "log-level", "error", "Log level (debug, info, warn, error).")
	rootCmd.Flags().StringVar(&since, "since", "", "Only consider commits after this date (e.g., '2023-01-02').")
	rootCmd.Flags().StringVar(&until, "until", "", "Only consider commits before this date (e.g., '2023-01-02').")
}
