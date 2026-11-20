package main

import (
	"fmt"
	"os"

	"github.com/SurTeam/TSInputMethod/internal/ranking"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "tsime-personal: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		printUsage()
		return fmt.Errorf("missing command")
	}
	path := ""
	i := 0
	if len(args) >= 2 && args[0] == "--path" {
		path = args[1]
		i = 2
	}
	if i >= len(args) {
		printUsage()
		return fmt.Errorf("missing command")
	}
	cmd := args[i]
	rest := args[i+1:]
	switch cmd {
	case "path":
		defaultPath, err := ranking.DefaultPersonalModelPath()
		if err != nil {
			return err
		}
		fmt.Println(defaultPath)
		return nil
	case "clear-weights":
		filePath, err := resolvePath(path, rest)
		if err != nil {
			return err
		}
		model, err := ranking.OpenPersonalModel(filePath, ranking.DefaultSGDConfig())
		if err != nil {
			return err
		}
		model.ResetPersonalization()
		if err := model.Save(); err != nil {
			return err
		}
		fmt.Printf("cleared weights path=%s updates=%d weights=%d\n", model.Path(), model.UpdateCount(), model.WeightCount())
		return nil
	case "status":
		filePath, err := resolvePath(path, rest)
		if err != nil {
			return err
		}
		model, err := ranking.OpenPersonalModel(filePath, ranking.DefaultSGDConfig())
		if err != nil {
			return err
		}
		fmt.Printf("path=%s updates=%d weights=%d\n", model.Path(), model.UpdateCount(), model.WeightCount())
		return nil
	default:
		printUsage()
		return fmt.Errorf("unknown command %q", cmd)
	}
}

func resolvePath(flagPath string, rest []string) (string, error) {
	if flagPath != "" {
		return flagPath, nil
	}
	for i := 0; i+1 < len(rest); i++ {
		if rest[i] == "--path" {
			return rest[i+1], nil
		}
	}
	return ranking.DefaultPersonalModelPath()
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "usage:\n")
	fmt.Fprintf(os.Stderr, "  tsime-personal path\n")
	fmt.Fprintf(os.Stderr, "  tsime-personal clear-weights [--path FILE]\n")
	fmt.Fprintf(os.Stderr, "  tsime-personal status [--path FILE]\n")
}
