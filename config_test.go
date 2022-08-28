package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseDepStr(t *testing.T) {
	dep, err := ParseDepString("github.com/owner/repo-v1.2.3")
	assert.Nil(t, err)
	assert.NotNil(t, dep)
	assert.Equal(t, "github.com", dep.Server, "they should be equal")
	assert.Equal(t, "owner", dep.Owner, "they should be equal")
	assert.Equal(t, "repo", dep.Repo, "they should be equal")
	assert.Equal(t, "v1.2.3", dep.Version, "they should be equal")
	assert.Equal(t, "github.com/owner/repo-v1.2.3", dep.String())
}

func TestParseDepStrWithMachineInfo(t *testing.T) {
	dep, err := ParseDepString("github.com/bazurto/openjdk-linux-amd64-v9.0.4")
	assert.Nil(t, err)
	assert.NotNil(t, dep)
	assert.Equal(t, "github.com", dep.Server, "they should be equal")
	assert.Equal(t, "bazurto", dep.Owner, "they should be equal")
	assert.Equal(t, "openjdk-linux-amd64", dep.Repo, "they should be equal")
	assert.Equal(t, "v9.0.4", dep.Version, "they should be equal")
	assert.Equal(t, "github.com/bazurto/openjdk-linux-amd64-v9.0.4", dep.String())
}

func TestParseDepStrEmptyString(t *testing.T) {
	dep, err := ParseDepString("")
	assert.NotNil(t, err)
	assert.Nil(t, dep)
	assert.Contains(t, err.Error(), "invalid format")
}

func TestParseDepStrEmptyServer(t *testing.T) {
	dep, err := ParseDepString("/owner/repo-v1.2.3")
	assert.NotNil(t, err)
	assert.Nil(t, dep)
	assert.Contains(t, err.Error(), "server name is required")
}

func TestParseDepStrNoServer(t *testing.T) {
	dep, err := ParseDepString("owner/repo-v1.2.3")
	assert.NotNil(t, err)
	assert.Nil(t, dep)
	assert.Contains(t, err.Error(), "invalid format")
}

func TestParseDepStrEmptyOwner(t *testing.T) {
	dep, err := ParseDepString("github.com//repo-v1.2.3")
	assert.NotNil(t, err)
	assert.Nil(t, dep)
	assert.Contains(t, err.Error(), "owner name is required")
}

func TestParseDepStrEmptyRepo(t *testing.T) {
	dep, err := ParseDepString("github.com/owner/-v1.2.3")
	assert.NotNil(t, err)
	assert.Nil(t, dep)
	assert.Contains(t, err.Error(), "repo name is required")
}

func TestParseDepStrEmptyVersionNumber(t *testing.T) {
	dep, err := ParseDepString("github.com/owner/repo-v")
	assert.NotNil(t, err)
	assert.Nil(t, dep)
	assert.Contains(t, err.Error(), "must have a version number")
}

func TestParseDepStrMissingVersion(t *testing.T) {
	dep, err := ParseDepString("github.com/owner/repo")
	assert.NotNil(t, err)
	assert.Nil(t, dep)
	assert.Contains(t, err.Error(), "no version")
}

func TestParseDepStrMustHaveLittleV(t *testing.T) {
	dep, err := ParseDepString("github.com/owner/repo-V1.2.3")
	assert.NotNil(t, err)
	assert.Nil(t, dep)
	assert.Contains(t, err.Error(), "'v' not 'V'")
}

func TestParseDepStrMustHaveVPrefix(t *testing.T) {
	dep, err := ParseDepString("github.com/owner/repo-1.2.3")
	assert.NotNil(t, err)
	assert.Nil(t, dep)
	assert.Contains(t, err.Error(), "prefixed with 'v'")
}
