package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/deckhouse/module-sdk/internal/controller"
)

func NewCMD(controller *controller.HookController) *CMD {
	return &CMD{
		controller: controller,
	}
}

type CMD struct {
	controller *controller.HookController
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func (c *CMD) Execute() {
	rootCmd := c.rootCmd()
	rootCmd.AddCommand(c.hooksCmd())

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

// rootCmd represents the base command when called without any subcommands
func (c *CMD) rootCmd() *cobra.Command {
	return &cobra.Command{
		Use:   filepath.Base(os.Args[0]),
		Short: "Binary with module hooks inside",
		Long: `Inside this binary contains list of 
precompiled hooks to use with corresponding module.`,
	}
}

func (c *CMD) hooksCmd() *cobra.Command {
	hooksCmd := &cobra.Command{
		Use:   "hook",
		Short: "Working with hooks",
		Long:  `Tools for working with nested hooks`,
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

	hooksCmd.AddCommand(&cobra.Command{
		Use:   "dump",
		Short: "Dump hooks configs",
		Long:  `Dump list of hooks configs in config.json file`,
		RunE: func(_ *cobra.Command, _ []string) error {
			err := c.controller.WriteHookConfigsInFile()
			if err != nil {
				return fmt.Errorf("can not write configs to file: %w", err)
			}

			fmt.Println("dump successfully")

			return nil
		},
	})

	hooksCmd.AddCommand(&cobra.Command{
		Use:   "run",
		Short: "Running hook",
		Long:  `Run hook from binary registry`,
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			idxRaw := args[0]
			idx, err := strconv.Atoi(idxRaw)
			if err != nil {
				return fmt.Errorf("argument '%s' is not integer", idxRaw)
			}

			err = c.controller.RunHook(idx)
			if err != nil {
				return fmt.Errorf("run hook error: %w", err)
			}

			return nil
		},
	})

	return hooksCmd
}
