package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/SurTeam/TSInputMethod/internal/dictionary"
	"github.com/SurTeam/TSInputMethod/internal/emoji"
	"github.com/SurTeam/TSInputMethod/internal/engine"
	"github.com/SurTeam/TSInputMethod/internal/ranking"
	"github.com/SurTeam/TSInputMethod/internal/session"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "tsimed: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	eng, err := loadEngine()
	if err != nil {
		return err
	}
	manager := session.NewManager(eng)
	if pe, ok := eng.(*engine.PinyinEngine); ok {
		if pe.UserDictionary() != nil {
			observer := dictionary.NewLearningObserver(pe.UserDictionary(), 0, true)
			manager.SetPhraseObserver(observer)
			manager.SetUserDictVersion(pe.UserDictionary().Version())
		}
		manager.SetFeedbackSink(session.EngineFeedbackSink{Target: pe})
		if model := pe.PersonalModel(); model != nil {
			if path, err := ranking.DefaultPersonalModelPath(); err == nil {
				if err := model.LoadFile(path); err != nil {
					fmt.Fprintf(os.Stderr, "tsimed: personal model: %v\n", err)
				} else {
					fmt.Printf("tsimed loaded personal model %s updates=%d\n", path, model.UpdateCount())
				}
			}
		}
		pe.ApplySettings()
	}
	server, addr, err := listenIPC(manager)
	if err != nil {
		return err
	}
	fmt.Printf("tsimed listening on %s\n", addr)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	errCh := make(chan error, 1)
	go func() {
		errCh <- server.Serve()
	}()

	select {
	case sig := <-sigCh:
		fmt.Printf("tsimed shutting down on %s\n", sig)
		_ = server.Close()
		return nil
	case err := <-errCh:
		return err
	}
}

func loadEngine() (engine.Engine, error) {
	path := resolveDictPath()
	if path == "" {
		fmt.Fprintln(os.Stderr, "tsimed: no dictionary configured; using stub engine")
		return engine.NewStubEngine(), nil
	}
	dict, err := dictionary.LoadFile(path)
	if err != nil {
		return nil, fmt.Errorf("load dictionary %s: %w", path, err)
	}
	eng := engine.NewPinyinEngine(dict)
	if user, err := dictionary.OpenUserDictionary(""); err == nil {
		eng.SetUserDictionary(user)
		fmt.Printf("tsimed loaded user dictionary %s entries=%d\n", user.Path(), user.EntryCount())
	} else {
		fmt.Fprintf(os.Stderr, "tsimed: user dictionary: %v\n", err)
	}
	if emojiPath := resolveEmojiPath(path); emojiPath != "" {
		if table, err := emoji.LoadFile(emojiPath); err == nil {
			eng.SetEmojiTable(table)
			fmt.Printf("tsimed loaded emoji table %s keys=%d\n", emojiPath, table.Size())
		} else {
			fmt.Fprintf(os.Stderr, "tsimed: emoji table %s: %v\n", emojiPath, err)
		}
	}
	fmt.Printf("tsimed loaded dictionary %s entries=%d scheme=%s\n", path, dict.EntryCount(), eng.Scheme())
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

func resolveEmojiPath(dictPath string) string {
	if override := os.Getenv("TSIME_EMOJI"); override != "" {
		return override
	}
	candidates := []string{
		filepath.Join("data", "dicts", "emoji.txt"),
		filepath.Join("third_party", "rime-ice", "opencc", "emoji.txt"),
	}
	if dictPath != "" {
		dir := filepath.Dir(dictPath)
		candidates = append([]string{
			filepath.Join(dir, "emoji.txt"),
			filepath.Join(dir, "opencc", "emoji.txt"),
		}, candidates...)
	}
	if exe, err := os.Executable(); err == nil {
		root := filepath.Dir(exe)
		candidates = append([]string{
			filepath.Join(root, "data", "dicts", "emoji.txt"),
			filepath.Join(root, "..", "data", "dicts", "emoji.txt"),
			filepath.Join(root, "..", "Resources", "data", "dicts", "emoji.txt"),
			filepath.Join(root, "..", "Resources", "dicts", "emoji.txt"),
		}, candidates...)
	}
	for _, path := range candidates {
		if st, err := os.Stat(path); err == nil && !st.IsDir() {
			return path
		}
	}
	return ""
}
