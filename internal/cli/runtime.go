package cli

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/howznguyen/knowns/internal/runtimeinstall"
	"github.com/howznguyen/knowns/internal/runtimequeue"
	"github.com/howznguyen/knowns/internal/search"
	"github.com/howznguyen/knowns/internal/storage"
	"github.com/spf13/cobra"
)

var runtimeInternalCmd = &cobra.Command{
	Use:    "__runtime",
	Short:  "Internal shared runtime",
	Hidden: true,
}

var runtimeRunCmd = &cobra.Command{
	Use:    "run",
	Short:  "Run the internal shared runtime",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runtimequeue.RunDaemon(cmd.Context(), search.ExecuteRuntimeJob, startRuntimeWatcher)
	},
}

var runtimeDaemonStatusCmd = &cobra.Command{
	Use:    "status",
	Short:  "Show shared runtime daemon status",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		status, err := runtimequeue.LoadStatus()
		if err != nil {
			return err
		}
		if isJSON(cmd) || isPlain(cmd) {
			printJSON(status)
			return nil
		}
		fmt.Printf("Runtime running: %v\n", status.Running)
		fmt.Printf("PID: %d\n", status.PID)
		fmt.Printf("Clients: %d\n", len(status.Clients))
		fmt.Printf("Projects: %d\n", len(status.Project))
		return nil
	},
}

var runtimeCmd = &cobra.Command{
	Use:   "runtime",
	Short: "Install and inspect runtime adapters",
}

var runtimeInstallCmd = &cobra.Command{
	Use:   "install <runtime>",
	Short: "Install a runtime memory adapter",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := runtimeinstall.DefaultOptions()
		if err := runtimeinstall.Install(args[0], opts); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Installed %s runtime adapter.\n", args[0])
		return nil
	},
}

var runtimeUninstallCmd = &cobra.Command{
	Use:   "uninstall <runtime>",
	Short: "Remove a runtime memory adapter",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := runtimeinstall.DefaultOptions()
		if err := runtimeinstall.Uninstall(args[0], opts); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Removed %s runtime adapter.\n", args[0])
		return nil
	},
}

var runtimeStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show supported runtime adapter status",
	RunE: func(cmd *cobra.Command, args []string) error {
		statuses, err := runtimeinstall.StatusAll(runtimeinstall.DefaultOptions())
		if err != nil {
			return err
		}
		runtimeinstall.SortStatuses(statuses)
		if isJSON(cmd) {
			printJSON(statuses)
			return nil
		}
		if isPlain(cmd) {
			for _, status := range statuses {
				fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\t%s\tavailable=%v\n", status.Runtime, status.HookKind, status.State, status.Available)
			}
			return nil
		}
		for _, status := range statuses {
			fmt.Fprintf(cmd.OutOrStdout(), "%s\n", status.DisplayName)
			fmt.Fprintf(cmd.OutOrStdout(), "  Kind: %s\n", status.HookKind)
			fmt.Fprintf(cmd.OutOrStdout(), "  State: %s\n", status.State)
			fmt.Fprintf(cmd.OutOrStdout(), "  Available: %v\n", status.Available)
			if len(status.Details) > 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "  Details: %s\n", strings.Join(status.Details, "; "))
			}
		}
		return nil
	},
}

func startRuntimeWatcher(ctx context.Context, storeRoot string) error {
	store := storage.NewStore(storeRoot)
	return StartCodeWatcher(ctx, store, filepath.Dir(storeRoot), watchDebounceMs)
}

func init() {
	runtimeInternalCmd.AddCommand(runtimeRunCmd)
	runtimeInternalCmd.AddCommand(runtimeDaemonStatusCmd)
	runtimeCmd.AddCommand(runtimeInstallCmd)
	runtimeCmd.AddCommand(runtimeUninstallCmd)
	runtimeCmd.AddCommand(runtimeStatusCmd)
	rootCmd.AddCommand(runtimeInternalCmd)
	rootCmd.AddCommand(runtimeCmd)
}
