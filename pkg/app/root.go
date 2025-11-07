package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/deckhouse/deckhouse/pkg/log"

	"github.com/deckhouse/module-sdk/internal/controller"
)

func newCMD(controller *controller.HookController, logger *log.Logger) *cmd {
	return &cmd{
		controller: controller,
		logger:     logger,
	}
}

type cmd struct {
	controller *controller.HookController
	logger     *log.Logger
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func (c *cmd) Execute() {
	rootCmd := c.buildCommand()

	err := rootCmd.Execute()
	if err != nil {
		c.logger.Error("failed to execute root command", "error", err)
		os.Exit(1)
	}
}

// buildCommand creates the complete command structure with all subcommands
func (c *cmd) buildCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   filepath.Base(os.Args[0]),
		Short: "Binary with module hooks inside",
		Long: `Inside this binary contains list of 
precompiled hooks to use with corresponding module.`,
	}
	rootCmd.AddCommand(c.hooksCmd())
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	// SilenceUsage and SilenceErrors prevent cobra from outputting
	// "Error: ..." and "Usage: ..." which would break JSON output parsing
	rootCmd.SilenceUsage = true
	rootCmd.SilenceErrors = true
	return rootCmd
}

// NOTE: all of the RunE errors and usage errors are silenced by SilenceUsage and SilenceErrors, to not cause "Invalid character 'E'" error
func (c *cmd) hooksCmd() *cobra.Command {
	hooksCmd := &cobra.Command{
		Use:     "hooks",
		Aliases: []string{"hook"},
		Short:   "Working with hooks",
		Long:    `Command for working with nested hooks`,
	}

	hooksCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "Listing hooks",
		Long:  `Get list of hooks from binary registry`,
		Run: func(_ *cobra.Command, _ []string) {
			hmetas := c.controller.ListHooksMeta()

			fmt.Printf("Found %d items:\n", len(hmetas))

			for idx, meta := range hmetas {
				fmt.Printf("%d - %s\n", idx, meta.Name)
			}
		},
	})

	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Print hooks configs",
		Long:  `Print list of hooks configs in json format`,
		RunE: func(_ *cobra.Command, _ []string) error {
			err := c.controller.PrintHookConfigs()
			if err != nil {
				c.logger.Error("can not print configs", "error", err)
				return fmt.Errorf("can not print configs: %w", err)
			}

			return nil
		},
	}
	hooksCmd.AddCommand(configCmd)

	dumpCmd := &cobra.Command{
		Use:    "dump",
		Short:  "Dump hooks configs",
		Long:   `Dump list of hooks configs in config.json file`,
		Hidden: true,
		RunE: func(_ *cobra.Command, _ []string) error {
			err := c.controller.WriteHookConfigsInFile()
			if err != nil {
				c.logger.Error("can not write configs to file", "error", err)
				return fmt.Errorf("can not write configs to file: %w", err)
			}

			fmt.Println("dump successfully")

			return nil
		},
	}
	hooksCmd.AddCommand(dumpCmd)

	runCmd := &cobra.Command{
		Use:    "run",
		Short:  "Running hook",
		Long:   `Run hook from binary registry`,
		Hidden: true,
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) != 1 {
				c.logger.Error("invalid number of arguments", "expected", 1, "received", len(args))

				return fmt.Errorf("invalid number of arguments: expected 1, received %d", len(args))
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			idxRaw := args[0]
			idx, err := strconv.Atoi(idxRaw)
			if err != nil {
				c.logger.Error("invalid argument", "argument", idxRaw, "error", err)

				return fmt.Errorf("invalid argument: %w", err)
			}

			err = c.controller.RunHook(ctx, idx)
			if err != nil {
				c.logger.Warn("hook shutdown", "error", err)
				return fmt.Errorf("hook shutdown: %w", err)
			}

			return nil
		},
	}
	hooksCmd.AddCommand(runCmd)

	readyCmd := &cobra.Command{
		Use:    "ready",
		Short:  "Check readiness",
		Long:   `Run readiness hook for module`,
		Hidden: true,
		Run: func(cmd *cobra.Command, _ []string) {
			ctx := cmd.Context()

			err := c.controller.RunReadiness(ctx)
			if err != nil {
				c.logger.Warn("readiness hook shutdown", "error", err)
				os.Exit(1)
			}
		},
	}
	hooksCmd.AddCommand(readyCmd)

	return hooksCmd
}
