package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/SurTeam/TSInputMethod/internal/dictionary"
	"github.com/SurTeam/TSInputMethod/internal/engine"
	"github.com/SurTeam/TSInputMethod/internal/session"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "imed: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	eng, err := loadEngine()
	if err != nil {
		return err
	}
	manager := session.NewManager(eng)
	server, addr, err := listenIPC(manager)
	if err != nil {
		return err
	}
	fmt.Printf("imed listening on %s\n", addr)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	errCh := make(chan error, 1)
	go func() {
		errCh <- server.Serve()
	}()

	select {
	case sig := <-sigCh:
		fmt.Printf("imed shutting down on %s\n", sig)
		_ = server.Close()
		return nil
	case err := <-errCh:
		return err
	}
}

func loadEngine() (engine.Engine, error) {
	path := resolveDictPath()
	if path == "" {
		fmt.Fprintln(os.Stderr, "imed: no dictionary configured; using stub engine")
		return engine.NewStubEngine(), nil
	}
	dict, err := dictionary.LoadFile(path)
	if err != nil {
		return nil, fmt.Errorf("load dictionary %s: %w", path, err)
	}
	eng := engine.NewPinyinEngine(dict)
	fmt.Printf("imed loaded dictionary %s entries=%d scheme=%s\n", path, dict.EntryCount(), eng.Scheme())
	return eng, nil
}

func resolveDictPath() string {
	if override := os.Getenv("TSIME_DICT"); override != "" {
		return override
	}
	candidates := []string{
		filepath.Join("data", "dicts", "rime-ice.tsdict"),
		filepath.Join("testdata", "dicts", "fixture.tsdict"),
	}
	if exe, err := os.Executable(); err == nil {
		root := filepath.Dir(exe)
		candidates = append([]string{
			filepath.Join(root, "data", "dicts", "rime-ice.tsdict"),
			filepath.Join(root, "..", "data", "dicts", "rime-ice.tsdict"),
			filepath.Join(root, "..", "Resources", "data", "dicts", "rime-ice.tsdict"),
			filepath.Join(root, "..", "Resources", "dicts", "rime-ice.tsdict"),
		}, candidates...)
	}
	for _, path := range candidates {
		if st, err := os.Stat(path); err == nil && !st.IsDir() {
			return path
		}
	}
	return ""
}
