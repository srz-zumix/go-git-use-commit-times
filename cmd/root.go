/*
Copyright © 2020 srz_zumix <https://github.com/srz-zumix>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	git "github.com/srz-zumix/git-use-commit-times/xgit"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var cfgFile string

func use_commit_times(path string, verbose bool, progress bool) error {
	repo, err := git.OpenRepository(path)
	if err != nil {
		return err
	}
	defer repo.Free()

	// fmt.Println(repo)
	filemap, err := ls_files(repo)
	if err != nil {
		return err
	}
	// fmt.Println(strings.Join(files, "\n"))
	err = use_commit_times_rev_walk(repo, filemap, verbose, progress)
	// err = use_commit_times_log_walk(repo, filemap, progress)
	if err != nil {
		return err
	}
	return nil
}

func use_commit_times_bylog(path string, since string, verbose bool, progress bool) error {
	repo, err := git.OpenRepository(path)
	if err != nil {
		return err
	}
	defer repo.Free()

	// fmt.Println(repo)
	filemap, err := ls_files(repo)
	if err != nil {
		return err
	}
	// fmt.Println(strings.Join(files, "\n"))
	// err = use_commit_times_rev_walk(repo, filemap, progress)
	err = use_commit_times_log_walk(repo, filemap, since, verbose, progress)
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
	Run: func(cmd *cobra.Command, args []string) {
		progress, err := cmd.Flags().GetBool("progress")
		if err != nil {
			log.Fatal(err)
		}
		verbose, err := cmd.Flags().GetBool("verbose")
		if err != nil {
			log.Fatal(err)
		}
		use_libgit2, err := cmd.Flags().GetBool("libgit-walk")
		if err != nil {
			log.Fatal(err)
		}
		since, err := cmd.Flags().GetString("since")
		if err != nil {
			log.Fatal(err)
		}
		if since == "" {
			since, err = cmd.Flags().GetString("after")
			if err != nil {
				log.Fatal(err)
			}
		}

		if use_libgit2 {
			err = use_commit_times(".", verbose, progress)
		} else {
			err = use_commit_times_bylog(".", since, verbose, progress)
		}
		if err != nil {
			log.Fatal(err)
		}
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
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.git-use-commit-times.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("progress", "p", false, "Show progressbar.")
	rootCmd.Flags().BoolP("libgit-walk", "l", false, "Get commit timestamp by libgit2 rev walk.")
	rootCmd.Flags().BoolP("verbose", "v", false, "Verbose.")
	rootCmd.Flags().String("since", "", "Commits more recent than a specific date.")
	rootCmd.Flags().String("after", "", "Commits more recent than a specific date.")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".git-use-commit-times" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".git-use-commit-times")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
