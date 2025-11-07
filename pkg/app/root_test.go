package app

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/deckhouse/deckhouse/pkg/log"

	"github.com/deckhouse/module-sdk/internal/controller"
)

func Test_HooksRun_NoErrorOrUsageOutput(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		description string
	}{
		{
			name:        "missing argument",
			args:        []string{"hooks", "run"},
			description: "should not output Error or Usage when argument is missing",
		},
		{
			name:        "invalid argument format",
			args:        []string{"hooks", "run", "not-a-number"},
			description: "should not output Error or Usage when argument is not a number",
		},
		{
			name:        "invalid hook index",
			args:        []string{"hooks", "run", "999"},
			description: "should not output Error or Usage when hook index is invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout and stderr
			var stdout, stderr bytes.Buffer

			// Create a mock controller
			cfg := &controller.Config{
				ModuleName: "test-module",
				HookConfig: &controller.HookConfig{
					BindingContextPath:    "in/binding_context.json",
					ValuesPath:            "in/values_path.json",
					ConfigValuesPath:      "in/config_values_path.json",
					HookConfigPath:        "out/hook_config.json",
					MetricsPath:           "out/metrics.json",
					KubernetesPath:        "out/kubernetes.json",
					ValuesJSONPath:        "out/values.json",
					ConfigValuesJSONPath:  "out/config_values.json",
					CreateFilesByYourself: true,
				},
				LogLevel: log.LevelFatal,
			}

			// Create logger that writes to buffer instead of stderr
			logBuf := bytes.Buffer{}
			logger := log.Default()
			logger.SetOutput(&logBuf)

			hookController := controller.NewHookController(cfg, logger)

			// Create command structure with our test logger
			cmd := newCMD(hookController, logger)

			// Build complete command structure
			rootCmd := cmd.buildCommand()
			rootCmd.SetOut(&stdout)
			rootCmd.SetErr(&stderr)

			// Set up test arguments
			rootCmd.SetArgs(tt.args)

			// Execute command
			_ = rootCmd.Execute()

			// Combine outputs (stdout, stderr, and log output)
			output := stdout.String() + stderr.String() + logBuf.String()
			// Check that Error: (from cobra) is not in output
			// We check for "Error:" with capital E, which is how cobra outputs errors
			// JSON logs may contain "error" field, but not "Error:" prefix from cobra
			assert.NotContains(t, output, "Error:", tt.description+": should not contain 'Error:' from cobra")

			// Check that Usage: (from cobra) is not in output
			// We check for "Usage:" with capital U, which is how cobra outputs usage
			assert.NotContains(t, output, "Usage:", tt.description+": should not contain 'Usage:' from cobra")
		})
	}
}

func Test_HooksConfig_NoErrorOrUsageOutput(t *testing.T) {
	// Capture stdout and stderr
	var stdout, stderr bytes.Buffer

	// Create a mock controller
	cfg := &controller.Config{
		ModuleName: "test-module",
		HookConfig: &controller.HookConfig{
			BindingContextPath:    "in/binding_context.json",
			ValuesPath:            "in/values_path.json",
			ConfigValuesPath:      "in/config_values_path.json",
			HookConfigPath:        "out/hook_config.json",
			MetricsPath:           "out/metrics.json",
			KubernetesPath:        "out/kubernetes.json",
			ValuesJSONPath:        "out/values.json",
			ConfigValuesJSONPath:  "out/config_values.json",
			CreateFilesByYourself: true,
		},
		LogLevel: log.LevelFatal,
	}

	// Create logger that writes to buffer instead of stderr
	logBuf := bytes.Buffer{}
	logger := log.Default()
	logger.SetOutput(&logBuf)

	hookController := controller.NewHookController(cfg, logger)

	// Create command structure with our test logger
	cmd := newCMD(hookController, logger)

	// Build complete command structure
	rootCmd := cmd.buildCommand()
	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)

	// Set up test arguments
	rootCmd.SetArgs([]string{"hooks", "config"})

	// Execute command
	err := rootCmd.Execute()

	// Combine outputs (stdout, stderr, and log output)
	output := stdout.String() + stderr.String() + logBuf.String()

	// Check that Error: (from cobra) is not in output
	// We check for "Error:" with capital E, which is how cobra outputs errors
	// JSON logs may contain "error" field, but not "Error:" prefix from cobra
	assert.NotContains(t, output, "Error:", "config command should not contain 'Error:' from cobra")

	// Check that Usage: (from cobra) is not in output
	// We check for "Usage:" with capital U, which is how cobra outputs usage
	assert.NotContains(t, output, "Usage:", "config command should not contain 'Usage:' from cobra")

	// Check that output is valid JSON (if there are hooks and no error)
	// Only check stdout for JSON, not stderr or logs
	if err == nil && len(stdout.String()) > 0 {
		require.True(t, strings.HasPrefix(strings.TrimSpace(stdout.String()), "{"), "config command should output valid JSON")
	}
}
