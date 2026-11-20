package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/SurTeam/TSInputMethod/internal/dictionary"
	"github.com/SurTeam/TSInputMethod/internal/pinyin"
)

func main() {
	dictionary.SetValidSyllableFunc(pinyin.IsValidSyllable)
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "tsime-userdict: %v\n", err)
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
	user, err := dictionary.OpenUserDictionary(path)
	if err != nil {
		return err
	}
	switch cmd {
	case "add":
		return cmdAdd(user, rest)
	case "remove":
		return cmdRemove(user, rest)
	case "update":
		return cmdUpdate(user, rest)
	case "lookup":
		return cmdLookup(user, rest)
	case "list":
		return cmdList(user)
	case "clear":
		return cmdClear(user, rest)
	case "reload":
		if err := user.ReloadLocked(); err != nil {
			return err
		}
		fmt.Printf("reloaded %s entries=%d version=%d\n", user.Path(), user.EntryCount(), user.Version())
		return nil
	case "path":
		fmt.Println(user.Path())
		return nil
	default:
		printUsage()
		return fmt.Errorf("unknown command %q", cmd)
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, `usage:
  tsime-userdict [--path FILE] add <text> <syllable...>
  tsime-userdict [--path FILE] remove <text> <syllable...>
  tsime-userdict [--path FILE] update <text> <syllable...> [--weight N] [--enable|--disable]
  tsime-userdict [--path FILE] lookup <text> <syllable...>
  tsime-userdict [--path FILE] list
  tsime-userdict [--path FILE] clear [learned|all]
  tsime-userdict [--path FILE] reload
  tsime-userdict [--path FILE] path`)
}

func cmdAdd(user *dictionary.UserDictionary, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("add requires text and syllables")
	}
	text := args[0]
	syllables := args[1:]
	if err := user.ReloadLocked(); err != nil {
		return err
	}
	entry, err := user.AddEntry(text, syllables, dictionary.UserSourceManual, 0)
	if err != nil {
		return err
	}
	if err := user.SaveLocked(); err != nil {
		return err
	}
	fmt.Printf("added %s [%s] weight=%d\n", entry.Text, strings.Join(entry.Syllables, "'"), entry.Weight)
	return nil
}

func cmdRemove(user *dictionary.UserDictionary, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("remove requires text and syllables")
	}
	if err := user.ReloadLocked(); err != nil {
		return err
	}
	if err := user.RemoveEntry(args[0], args[1:]); err != nil {
		return err
	}
	return user.SaveLocked()
}

func cmdUpdate(user *dictionary.UserDictionary, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("update requires text and syllables")
	}
	text := args[0]
	syllables := make([]string, 0, len(args)-1)
	var weight uint32
	var enabled *bool
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--weight":
			if i+1 >= len(args) {
				return fmt.Errorf("--weight needs a value")
			}
			n, err := strconv.ParseUint(args[i+1], 10, 32)
			if err != nil {
				return err
			}
			weight = uint32(n)
			i++
		case "--enable":
			v := true
			enabled = &v
		case "--disable":
			v := false
			enabled = &v
		default:
			syllables = append(syllables, args[i])
		}
	}
	if len(syllables) == 0 {
		return fmt.Errorf("update requires syllables")
	}
	if err := user.ReloadLocked(); err != nil {
		return err
	}
	entry, err := user.UpdateEntry(text, syllables, weight, enabled)
	if err != nil {
		return err
	}
	if err := user.SaveLocked(); err != nil {
		return err
	}
	fmt.Printf("updated %s [%s] weight=%d enabled=%v\n", entry.Text, strings.Join(entry.Syllables, "'"), entry.Weight, entry.Enabled)
	return nil
}

func cmdLookup(user *dictionary.UserDictionary, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("lookup requires text and syllables")
	}
	if err := user.ReloadLocked(); err != nil {
		return err
	}
	entry, ok := user.LookupEntry(args[0], args[1:])
	if !ok {
		return fmt.Errorf("not found")
	}
	fmt.Printf("%s\t%s\tweight=%d\tusage=%d\tsource=%s\tenabled=%v\n",
		entry.Text, strings.Join(entry.Syllables, "'"), entry.Weight, entry.UsageCount, entry.Source, entry.Enabled)
	return nil
}

func cmdList(user *dictionary.UserDictionary) error {
	if err := user.ReloadLocked(); err != nil {
		return err
	}
	for _, entry := range user.ListEntries() {
		fmt.Printf("%s\t%s\tweight=%d\tusage=%d\tsource=%s\tenabled=%v\n",
			entry.Text, strings.Join(entry.Syllables, "'"), entry.Weight, entry.UsageCount, entry.Source, entry.Enabled)
	}
	return nil
}

func cmdClear(user *dictionary.UserDictionary, args []string) error {
	mode := "learned"
	if len(args) > 0 {
		mode = args[0]
	}
	if err := user.ReloadLocked(); err != nil {
		return err
	}
	var n int
	switch mode {
	case "learned":
		n = user.ClearLearnedEntries()
	case "all":
		n = user.ClearAllEntries()
	default:
		return fmt.Errorf("clear expects learned|all")
	}
	if err := user.SaveLocked(); err != nil {
		return err
	}
	fmt.Printf("cleared %d entries\n", n)
	return nil
}
