package main

import "fmt"

type Dep struct {
	Server  string
	Owner   string
	Repo    string
	Version string // with v
}

func (d *Dep) CanonicalNameNoVersion() string {
	return fmt.Sprintf("%s/%s/%s", d.Server, d.Owner, d.Repo)
}

func (d *Dep) String() string {
	return fmt.Sprintf("%s/%s/%s-%s", d.Server, d.Owner, d.Repo, d.Version)
}
