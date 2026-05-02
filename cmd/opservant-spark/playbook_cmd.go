package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"syscall"

	"github.com/s4e-io/opservant-spark/internal/agent"
	"github.com/s4e-io/opservant-spark/internal/config"
	"github.com/s4e-io/opservant-spark/internal/initializer"
	"github.com/s4e-io/opservant-spark/internal/logger"
)

func cmdPlaybook(args []string) {
	fs := flag.NewFlagSet("playbook", flag.ExitOnError)
	configFlag := fs.String("config", "config.yaml", "Path to config file")
	playbookDirFlag := fs.String("dir", "", "Directory of playbook files to run")
	fs.Parse(args)

	if *playbookDirFlag == "" && fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "Error: playbook file or --dir is required")
		fmt.Fprintln(os.Stderr, "\nUsage: spark playbook [--config <path>] <file>")
		fmt.Fprintln(os.Stderr, "       spark playbook [--config <path>] --dir <dir>")
		os.Exit(1)
	}
	if *playbookDirFlag != "" && fs.NArg() > 0 {
		fmt.Fprintln(os.Stderr, "Error: --dir and a playbook file are mutually exclusive")
		fmt.Fprintln(os.Stderr, "\nUsage: spark playbook [--config <path>] <file>")
		fmt.Fprintln(os.Stderr, "       spark playbook [--config <path>] --dir <dir>")
		os.Exit(1)
	}
	if *playbookDirFlag == "" && fs.NArg() > 1 {
		fmt.Fprintf(os.Stderr, "Error: unexpected argument %q\n\n", fs.Arg(1))
		printHelp()
		os.Exit(1)
	}

	cfg, err := config.Load(*configFlag)
	if err != nil {
		fatalf("failed to load config: %v", err)
	}
	if err := cfg.Validate(); err != nil {
		fatalf("invalid config: %v", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	orch := initializer.NewOrchestrator(
		initializer.WithConfigPath(*configFlag),
		initializer.WithConfig(cfg),
		initializer.WithLogLevel(cfg.Logging.Level),
	)
	if err := orch.Initialize(ctx); err != nil {
		fatalf("initialization failed: %v", err)
	}
	defer orch.Cleanup()

	state := orch.GetState()
	if *playbookDirFlag != "" {
		playbookFiles, err := playbookFilesInDir(*playbookDirFlag)
		if err != nil {
			fatalf("failed to read playbook directory: %v", err)
		}
		if len(playbookFiles) == 0 {
			fatalf("no playbook files found in directory: %s", *playbookDirFlag)
		}

		var failed bool
		for _, playbookFile := range playbookFiles {
			select {
			case <-ctx.Done():
				os.Exit(1)
			default:
			}
			if err := processPlaybook(ctx, playbookFile, state.GetConfig(), state.GetLogger()); err != nil {
				if !errors.Is(ctx.Err(), context.Canceled) {
					failed = true
					state.GetLogger().Error("Playbook processing failed: %s: %v", playbookFile, err)
				}
			}
		}
		if failed {
			os.Exit(1)
		}
		return
	}

	playbookFile := fs.Arg(0)
	if err := processPlaybook(ctx, playbookFile, state.GetConfig(), state.GetLogger()); err != nil {
		if !errors.Is(ctx.Err(), context.Canceled) {
			state.GetLogger().Error("Playbook processing failed: %s: %v", playbookFile, err)
			os.Exit(1)
		}
	}
}

func processPlaybook(ctx context.Context, playbookFile string, cfg *config.Config, lgr *logger.Log) error {
	a := agent.New(cfg, lgr)
	return a.ExecutePlaybook(ctx, playbookFile)
}

func playbookFilesInDir(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if filepath.Ext(entry.Name()) == ".json" {
			files = append(files, filepath.Join(dir, entry.Name()))
		}
	}

	sort.Strings(files)
	return files, nil
}
