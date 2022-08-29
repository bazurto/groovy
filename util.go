package main

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/vibrantbyte/go-antpath/antpath"
)

var (
	antMatcher = antpath.New()
)

func mkdir(d string) error {
	if err := os.MkdirAll(d, 0770); err != nil {
		return err
	}
	return nil
}

func fileExists(f string) bool {
	_, err := os.Stat(f)
	return !os.IsNotExist(err)
}

func mkdirIfNotExists(d string) error {
	if fileExists(d) {
		return nil
	}
	return mkdir(d)
}

func Uncompress(archiveFileName, dstDirName string) error {
	if strings.HasSuffix(archiveFileName, ".zip") {
		return Unzip(archiveFileName, dstDirName)
	} else if strings.HasSuffix(archiveFileName, ".tgz") {
		return Untgz(archiveFileName, dstDirName)
	} else if strings.HasSuffix(archiveFileName, ".tar.gz") {
		return Untgz(archiveFileName, dstDirName)
	}
	return fmt.Errorf("archive file extension not supported: %s", archiveFileName)
}

func Unzip(zipFileName, dstDirName string) error {
	reader, err := zip.OpenReader(zipFileName)
	if err != nil {
		return err
	}
	defer reader.Close()

	if err := os.MkdirAll(dstDirName, 0755); err != nil {
		return err
	}

	// valid prefix to avoid out of ZipSlip hack
	validPrefix := filepath.Clean(dstDirName) + string(os.PathSeparator)

	// Closure to address file descriptors issue with all the deferred .Close() methods
	extractAndWriteFile := func(zipFile *zip.File) error {
		rc, err := zipFile.Open()
		if err != nil {
			return err
		}
		defer rc.Close()

		path := filepath.Clean(filepath.Join(dstDirName, filepath.FromSlash(zipFile.Name)))

		// Check for ZipSlip (Directory traversal)
		if !strings.HasPrefix(path, validPrefix) {
			return fmt.Errorf("illegal file path: %s", path)
		}

		if zipFile.FileInfo().IsDir() {
			// Dir
			os.MkdirAll(path, zipFile.Mode())
		} else {
			// File
			os.MkdirAll(filepath.Dir(path), zipFile.Mode())
			f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, zipFile.Mode())
			if err != nil {
				return err
			}
			defer f.Close()

			_, err = io.Copy(f, rc)
			if err != nil {
				return err
			}
		}
		return nil
	}

	for _, zipFile := range reader.File {
		err := extractAndWriteFile(zipFile)
		if err != nil {
			return err
		}
	}

	return nil
}

// Zip zips up files
func Zip(srcDir string, writer io.Writer, include []string) error {
	var err error
	srcDir, err = filepath.Abs(srcDir) // clean path
	if err != nil {
		return fmt.Errorf("Zip() filepath.Abs: %w", err)
	}

	// fixed patterns by appendding /** to dir include names
	// - dirToIncude/ => dirToInclude/**
	// - dirToIncude => dirToInclude/**
	var fixedPatterns []string
	for _, pattern := range include {
		stat, err := os.Stat(pattern)
		if err == nil {
			if stat.IsDir() {
				if strings.HasSuffix(pattern, "/") { //pattern ends with /, then append **
					pattern = pattern + "**"
				} else { // does not end in /, append /**
					pattern = pattern + "/**"
				}
			}
		}
		fixedPatterns = append(fixedPatterns, pattern)
	}
	fmt.Printf("%v\n", fixedPatterns)

	tw := zip.NewWriter(writer)
	defer tw.Close()

	err = RecurseDir(srcDir, func(fullFilename string, file *os.File) error {
		filename := strings.Replace(fullFilename, srcDir, "", 1) // /path/to/srcDir/dir/file => /dir/file
		filename = strings.TrimLeft(filename, "/")               // dir/file
		filename = filepath.ToSlash(filename)                    // replace windows filenames to *nix filenames: dir\file => dir/file

		includeFile := false
		for _, pattern := range fixedPatterns {
			includeFile = antMatcher.Match(pattern, filename)
			//includeFile, _ = path.Match(pattern, filename)
			if includeFile {
				break
			}
		}
		if !includeFile {
			return nil
		}
		fmt.Printf("- %s\n", filename)

		// Get FileInfo about our file providing file size, mode, etc.
		info, err := file.Stat()
		if err != nil {
			return fmt.Errorf("file.Stat(%s): %w", fullFilename, err)
		}

		// Create a tar Header from the FileInfo data
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return fmt.Errorf("zip.FileInfoHeader(%s): %w", fullFilename, err)
		}
		header.Name = filename

		// Write file header to the tar archive
		w, err := tw.CreateHeader(header)
		if err != nil {
			return fmt.Errorf("zip.CreateHeader(%s): %w", filename, err)
		}

		// Copy contents if it is regular file
		if info.Mode().IsRegular() {
			_, err = io.Copy(w, file)
			if err != nil {
				return fmt.Errorf("io.Copy(%s, %s): %w", filename, fullFilename, err)
			}
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("Zip(): %w", err)
	}
	return nil
}

// RecurseDir loop throguh all files and directories recursively.  cb is called back
// with the name and the open file
func RecurseDir(absDir string, cb func(absName string, file *os.File) error) error {
	//
	diSlice, err := ioutil.ReadDir(absDir)
	if err != nil {
		return fmt.Errorf("RecurseDir(): %w", err)
	}
	for _, di := range diSlice {
		var err error
		if di.IsDir() {
			// func for defer to work
			err = func() error {
				absName := filepath.Join(absDir, di.Name())
				file, err := os.Open(absName)
				if err != nil {
					return fmt.Errorf("RecurseDir() os.Open: %w", err)
				}
				defer file.Close()
				if err := cb(absName, file); err != nil {
					return fmt.Errorf("RecurseDir(): %w", err)
				}
				if err := RecurseDir(absName, cb); err != nil {
					return err
				}
				return nil
			}()
		} else {
			err = func() error {
				absName := filepath.Join(absDir, di.Name())
				file, err := os.Open(absName)
				if err != nil {
					return fmt.Errorf("RecurseDir() os.Open: %w", err)
				}
				defer file.Close()
				if err := cb(absName, file); err != nil {
					return err
				}
				return nil
			}()
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func Untgz(fileName, dir string) error {
	gzipStream, err := os.Open(fileName)
	if err != nil {
		return fmt.Errorf("Untgz: Opening file (%s): %w", fileName, err)
	}

	uncompressedStream, err := gzip.NewReader(gzipStream)
	if err != nil {
		return fmt.Errorf("Untgz: NewReader failed: %w", err)
	}

	tarReader := tar.NewReader(uncompressedStream)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	for true {
		header, err := tarReader.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			return fmt.Errorf("Untgz: Next() failed: %w", err)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			realName, err := uncompressActualPath(dir, header.Name)
			if err != nil {
				return fmt.Errorf("untgz: filepath.abs() failed: %w", err)
			}

			if err := os.Mkdir(realName, 0755); err != nil {
				return fmt.Errorf("Untgz: Mkdir() failed: %w", err)
			}
		case tar.TypeReg:
			realName, err := uncompressActualPath(dir, header.Name)
			if err != nil {
				return fmt.Errorf("untgz: filepath.abs() failed: %w", err)
			}

			outFile, err := os.OpenFile(realName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return fmt.Errorf("Untgz: Create() failed: %w", err)
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				return fmt.Errorf("Untgz: Copy() failed: %w", err)
			}
			outFile.Close()

		default:
			return fmt.Errorf("Untgz: unknown type: %b in %s ", header.Typeflag, header.Name)
		}

	}
	return nil
}

func uncompressActualPath(dir, path string) (string, error) {
	var err error
	realName := filepath.Clean(filepath.Join(dir, filepath.FromSlash(path)))
	if err != nil {
		return "", fmt.Errorf("Uncompress: filepath.Abs() failed: %w", err)
	}
	if !strings.HasPrefix(realName, dir) {
		return "", fmt.Errorf("Uncompress: path(%s) not contained within path(%s)", realName, dir)
	}
	return realName, nil
}
