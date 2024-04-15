package cmd

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/tsukinoko-kun/serve/internal/handler"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run serve to start a webserver hosting the content of a directory",
	Long: "Serve is a simple webserver that serves the content of a directory. " +
		"By default, it serves the current directory on port 8080.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		verbose, _ := cmd.Flags().GetBool("verbose")
		if verbose {
			log.SetLevel(log.DebugLevel)
		} else {
			log.SetLevel(log.InfoLevel)
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// get directory path
		var dir string
		switch len(args) {
		case 0:
			dir = "."
		case 1:
			dir = args[0]
		default:
			return fmt.Errorf("too many arguments")
		}

		// check if the directory exists
		fi, err := os.Stat(dir)
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("irectory does not exist: %q", dir)
			}
			return errors.Join(fmt.Errorf("failed to check if directory exists: %q", dir), err)
		}

		// check if the path is a directory
		if !fi.IsDir() {
			return fmt.Errorf("path is not a directory: %q", dir)
		}

		// get absolute path
		absPath, err := filepath.Abs(dir)
		if err != nil {
			return errors.Join(fmt.Errorf("failed to get absolute path: %q", dir), err)
		}

		// get md flag
		mdCompile, err := cmd.Flags().GetBool("md")
		if err != nil {
			return errors.Join(fmt.Errorf("failed to get md flag"), err)
		}

		// create file server
		handler.Init(absPath, mdCompile)

		// handle all requests with the handler
		http.Handle("/", http.HandlerFunc(handler.Handle))

		// Start the HTTP server
		port, _ := cmd.Flags().GetInt("port")
		listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err != nil {
			return errors.Join(fmt.Errorf("failed to listen on port %d", port), err)
		}
		fmt.Printf("Listening on http://%s\n", listener.Addr())

		go func() {
			err := http.Serve(listener, nil)
			if err != nil {
				log.Fatal(err)
			}
		}()

		// await termination (e.g. via SIGINT)
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		signal := <-c
		fmt.Println("")
		log.Infof("Received signal %s, shutting down...", signal)
		// close the listener
		if err := listener.Close(); err != nil {
			return errors.Join(fmt.Errorf("failed to close listener"), err)
		}

		return nil
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().Bool("md", false, "Compile Markdown files to HTML")
	rootCmd.Flags().IntP("port", "p", 0, "Port to listen on")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Verbose output")
}
