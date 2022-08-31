package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/amzn/ion-go/ion"
)

/*
	{
		server: {
			"github.com": {
				token: "..."
			}
		}
	}
*/
type BzUserConfig struct {
	Server map[string]BzUserConfigServerAttrs `ion:"server"`
}
type BzUserConfigServerAttrs struct {
	Token string `ion:"token"`
}

type Config struct {
	BinDir string            `ion:"binDir"`
	Deps   []string          `ion:"deps"`
	Export map[string]string `ion:"env"`
}

func (c *Config) GetDeps() []*Dep {
	var deps []*Dep
	for _, d := range c.Deps {

		dep, err := ParseDepString(d)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: %s", err)
			continue
		}
		deps = append(deps, dep)
	}
	return deps
}

func ParseDepString(depStr string) (*Dep, error) {
	parts := strings.Split(depStr, "/") // github.com/owner/repo-v1.2.3 -> github.com,owner,repo-v1.2.3
	if len(parts) != 3 {
		return nil, fmt.Errorf("unable to parse dependency '%s': invalid format", depStr)
	}

	server := parts[0]
	if server == "" {
		return nil, fmt.Errorf("unable to parse dependency '%s': server name is required", depStr)
	}

	owner := parts[1]
	if owner == "" {
		return nil, fmt.Errorf("unable to parse dependency '%s': owner name is required", depStr)
	}
	repoPlusVersion := parts[2]
	parts2 := strings.Split(repoPlusVersion, "-")
	l := len(parts2)
	if l < 2 {
		return nil, fmt.Errorf("unable to parse dependency '%s': no version", depStr)
	}

	repo := strings.Join(parts2[0:l-1], "-")
	if repo == "" {
		return nil, fmt.Errorf("unable to parse dependency '%s': repo name is required", depStr)
	}

	version := parts2[l-1]
	if strings.HasPrefix(version, "V") {
		return nil, fmt.Errorf("unable to parse dependency '%s': version prefix should be 'v' not 'V'", depStr)
	}

	if !strings.HasPrefix(version, "v") {
		return nil, fmt.Errorf("unable to parse dependency '%s': could not find a version prefixed with 'v'", depStr)
	}

	if len(version) < 2 {
		return nil, fmt.Errorf("unable to parse dependency '%s': must have a version number", depStr)
	}

	return &Dep{
		Server:  server,
		Owner:   owner,
		Repo:    repo,
		Version: version,
	}, nil
}

func hclLoadConfig(f string, cfg any) error {
	b, err := ioutil.ReadFile(f)
	if err != nil {
		return fmt.Errorf("Error reading file %s: %s\n", f, err)
	}
	return ion.Unmarshal(b, cfg)

	/*
		var ctx *hcl.EvalContext //nil

		var file *hcl.File
		var diags hcl.Diagnostics

		file, diags = hclsyntax.ParseConfig(b, f, hcl.Pos{Line: 1, Column: 1})
		if diags.HasErrors() {
			return diags
		}

		diags = gohcl.DecodeBody(file.Body, ctx, cfg)
		if diags.HasErrors() {
			return diags
		}

	*/
}

func (o *BzUserConfig) GetServerToken(serverName string) string {
	var token string
	for serverName, attr := range o.Server {
		if strings.ToLower(serverName) == strings.ToLower(serverName) {
			token = attr.Token
		}
	}
	return token
}
