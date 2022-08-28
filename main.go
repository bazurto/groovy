package main

import (
	"fmt"
	"os"
	"path/filepath"
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
	config := &Config{}
	err := hclLoadConfig(".bz", config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
		os.Exit(1)
	}

	for _, dep := range config.GetDeps() {
		fmt.Printf("githubDownloadDependency: %s\n", dep.String())
		if depFile, err := resolveDependency(dep); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		} else {
			fmt.Printf("found dependency: %s\n", depFile)
		}
	}
}
