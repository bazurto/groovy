package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	appName = "bz"
)

var (
	bzUserDir        string // $HOME/bz/
	bzUserConfigFile string // $HOME/bz/config
	bzUserCacheDir   string // $HOME/bz/cache/
	bzUserConfig     BzUserConfig
)

func init() {
	homeDir, _ := os.UserHomeDir()
	bzUserDir = filepath.Join(homeDir, fmt.Sprintf(".%s", appName))
	bzUserConfigFile = filepath.Join(bzUserDir, "config")
	bzUserCacheDir = filepath.Join(bzUserDir, "cache")

	// create directories if it does not exist
	if err := mkdirIfNotExists(bzUserCacheDir); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating create cache dir: %s\n", err)
		os.Exit(1)
	}

	// load user config if it exists
	if fileExists(bzUserConfigFile) {
		if err := hclLoadConfig(bzUserConfigFile, &bzUserConfig); err != nil {
			fmt.Fprintf(os.Stderr, "Error reading user config (%s): %s\n", bzUserConfigFile, err)
			os.Exit(1)
		}
	}
}

func main() {

	// get all extracted directories from dependencies
	eaList := resolveFromBzFilePath(".bz")
	var newPath []string
	for _, ea := range eaList {
		newPath = append(newPath, ea.BinDir)
	}

	//
	path := os.Getenv("PATH")
	pathParts := strings.Split(path, string([]rune{os.PathListSeparator}))
	newPath = append(newPath, pathParts...)
	os.Setenv("PATH", strings.Join(newPath, string([]rune{os.PathListSeparator})))

	if len(os.Args) > 1 {
		os.Exit(executeCommand())
	}

}

func executeCommand() int {
	prog := os.Args[1]
	args := os.Args[2:]
	cmd := exec.Command(prog, args...)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return exitError.ExitCode()
		} else {
			log.Printf("Unable to run `%s`: %s", strings.Join(os.Args[1:], " "), err)
			return exitError.ExitCode()
		}
	}
	return 0
}

type ExtractedDependency struct {
	Dir     string
	BinDir  string
	Exports map[string]string
}

func resolveFromBzFilePath(bzFilePath string) []ExtractedDependency {
	return resolveFromBzFilePathStack(bzFilePath, nil)
}
func resolveFromBzFilePathStack(bzFilePath string, stack []string) []ExtractedDependency {
	if !fileExists(bzFilePath) {
		return nil
	}

	bzFile := &Config{}
	err := hclLoadConfig(bzFilePath, bzFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
		os.Exit(1)
	}
	var edList []ExtractedDependency
	for _, dep := range bzFile.GetDeps() {
		// Circular depedency protection
		for _, s := range stack {
			if dep.CanonicalNameNoVersion() == s {
				stack = append(stack, dep.CanonicalNameNoVersion())
				fmt.Fprintf(os.Stderr, "Detected circular dependency: %s\n", strings.Join(stack, "->"))
				os.Exit(1)
			}
		}
		subStack := append(stack, dep.CanonicalNameNoVersion())
		/*
			if config.Export != nil {
				a, _ := config.Export.ExportBody.JustAttributes()
				for k, v := range a {
					fmt.Printf("### %s = %s###\n", k, v.Name)
				}
			}
		*/
		extractedDependencyDir, err := resolveDependency(dep)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		}

		subEdList := resolveFromBzFilePathStack(filepath.Join(extractedDependencyDir, ".bz"), subStack)
		edList = append(edList, subEdList...)

		// bin dir
		var binDir string
		if bzFile.Desc != nil && bzFile.Desc.BinDir != nil {
			binDir = *bzFile.Desc.BinDir
		} else {
			binDir = filepath.Join(extractedDependencyDir, "bin")
		}

		//
		ed := ExtractedDependency{
			Dir:    extractedDependencyDir,
			BinDir: binDir,
		}
		edList = append(edList, ed)
	}
	return edList
}
