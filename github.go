package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/google/go-github/v47/github"
	"golang.org/x/oauth2"
)

type TestStruct struct {
	Name string
}

func resolveDependency(dep *Dep) (string, error) {
	downloadToDir := filepath.Join(bzUserCacheDir, "deps", dep.Server, dep.Owner, dep.Repo, dep.Version)

	extractedDir := resolveExtractedDir(downloadToDir, dep)
	err := cache(extractedDir, func() error {
		downloadedFile, err := resolveDependencyDownloadAsset(downloadToDir, dep)
		if err != nil {
			return fmt.Errorf("resolveDependency error: %s", err)
		}

		if err := Uncompress(downloadedFile, extractedDir); err != nil {
			return fmt.Errorf("resolveDependency error: error unzipping '%s' to '%s': %s", downloadedFile, extractedDir, err)
		}

		return nil
	})
	return extractedDir, err
}

func resolveDependencyDownloadAsset(downloadToDir string, dep *Dep) (string, error) {
	// github client
	ctx := context.Background()
	githubAccessToken := bzUserConfig.GetServerToken(dep.Server)
	var tc *http.Client = nil
	if githubAccessToken != "" {
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: githubAccessToken},
		)
		tc = oauth2.NewClient(ctx, ts)
	}
	client := github.NewClient(tc)

	//
	release := &github.RepositoryRelease{}
	githubReleaseCacheFile := filepath.Join(downloadToDir, fmt.Sprintf("%s.githubReleaseJson", getCanonicalName(dep)))

	// query github
	err := cacheAny(githubReleaseCacheFile, release, func() (any, error) {
		// get release by tag
		release, _, err := client.Repositories.GetReleaseByTag(ctx, dep.Owner, dep.Repo, dep.Version)
		if err != nil {
			return nil, err
		}
		return release, nil
	})
	if err != nil {
		return "", err
	}

	// -------
	// Get List of all assets in that release
	var githubAssetNames []string
	for _, a := range release.Assets {
		githubAssetNames = append(githubAssetNames, a.GetName())
	}

	// Get the asset name that we should download in the priority order of possible asset names function
	var asset *github.ReleaseAsset
	expectedNames := possibleAssetNames(dep)
	for _, expected := range expectedNames {
		for _, a := range release.Assets {
			//fmt.Fprintf(os.Stderr, "Comparing %s == %s\n", expected.NameWithExt(), a.GetName())
			if expected.NameWithExt() == a.GetName() {
				asset = a
				break
			}
		}
	}

	if asset == nil {
		return "", fmt.Errorf(
			"Could not find asset %s in depedency %s",
			strings.Join(BzAssetArrHelper(expectedNames).CollectNames(), ","),
			dep.String(),
		)
	}

	fmt.Printf("Found asset %s\n", asset.GetName())

	downloadedFile := filepath.Join(downloadToDir, asset.GetName())
	err = cache(downloadedFile, func() error {
		readCloser, _, err := client.Repositories.DownloadReleaseAsset(ctx, dep.Owner, dep.Repo, asset.GetID(), http.DefaultClient)
		if err != nil {
			return err
		}
		defer readCloser.Close()

		// download file
		log.Printf("Creating dir if not exists %s", downloadToDir)
		if err := mkdirIfNotExists(downloadToDir); err != nil {
			return err
		}

		downloadFileTmp := fmt.Sprintf("%s.tmp", downloadedFile)
		w, err := os.Create(downloadFileTmp)
		if err != nil {
			return err
		}
		defer w.Close()

		log.Printf("Downloading file %s", downloadFileTmp)
		if _, err := io.Copy(w, readCloser); err != nil {
			return err
		}

		// rename tmp download file to downloadFile
		log.Printf("Renaming temp file to %s", downloadedFile)
		if err := os.Rename(downloadFileTmp, downloadedFile); err != nil {
			return err
		}

		return nil
	})
	return downloadedFile, nil
}

type BzAsset struct {
	Ext       string // zip
	Canonical string // project-name-linux-amd64-v1.2.3
}

func (a *BzAsset) NameWithExt() string {
	return fmt.Sprintf("%s.%s", a.Canonical, a.Ext)
}

type BzAssetArrHelper []BzAsset

func (o BzAssetArrHelper) CollectNames() []string {
	var names []string
	for _, i := range o {
		names = append(names, i.NameWithExt())
	}
	return names
}

func possibleAssetNames(dep *Dep) []BzAsset {
	osArch := fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH)

	extensions := []string{"zip", "tgz", "tar.gz"} // possible extensions

	var res []BzAsset
	for _, ext := range extensions {
		res = append(res,
			BzAsset{Canonical: fmt.Sprintf("%s-%s-%s", dep.Repo, osArch, dep.Version), Ext: ext}, // openjdk-linux-amd64-v1.2.3.zip
			BzAsset{Canonical: fmt.Sprintf("%s-%s", dep.Repo, dep.Version), Ext: ext},            // openjdk-v1.2.3.zip
			BzAsset{Canonical: fmt.Sprintf("%s", dep.Repo), Ext: ext},                            // openjdk.zip
		)
	}
	return res
}

func resolveExtractedDir(baseDir string, dep *Dep) string {
	//name := fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH)
	//dir := filepath.Join(baseDir, dep.String)
	//log.Printf("Resolved extrated dir to: %s", dir)
	dir := filepath.Join(baseDir, "extracted")
	return dir
}

func getCanonicalName(dep *Dep) string {
	osArch := fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH)
	return fmt.Sprintf("%s-%s-%s", dep.Repo, osArch, dep.Version)
}
