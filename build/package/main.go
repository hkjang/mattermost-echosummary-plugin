package main

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"flag"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type manifest struct {
	Server struct {
		Executables map[string]string `json:"executables"`
	} `json:"server"`
}

func main() {
	sourceDir := flag.String("source", "", "plugin directory to package")
	outputPath := flag.String("output", "", "path to the generated .tar.gz file")
	manifestPath := flag.String("manifest", "plugin.json", "path to plugin manifest")
	flag.Parse()

	if *sourceDir == "" || *outputPath == "" {
		panic("source and output are required")
	}

	executables, err := loadExecutables(*manifestPath)
	if err != nil {
		panic(err)
	}

	entries, err := collectEntries(*sourceDir)
	if err != nil {
		panic(err)
	}

	if err := os.MkdirAll(filepath.Dir(*outputPath), 0o755); err != nil {
		panic(err)
	}

	file, err := os.Create(*outputPath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	gzipWriter := gzip.NewWriter(file)
	defer gzipWriter.Close()

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	baseName := filepath.Base(filepath.Clean(*sourceDir))
	for _, entry := range entries {
		if err := addEntry(tarWriter, *sourceDir, baseName, entry, executables); err != nil {
			panic(err)
		}
	}
}

func loadExecutables(path string) (map[string]struct{}, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var parsed manifest
	if err := json.Unmarshal(data, &parsed); err != nil {
		return nil, err
	}

	executables := make(map[string]struct{}, len(parsed.Server.Executables))
	for _, executablePath := range parsed.Server.Executables {
		normalized := filepath.ToSlash(filepath.Clean(executablePath))
		executables[normalized] = struct{}{}
	}

	return executables, nil
}

func collectEntries(root string) ([]string, error) {
	entries := make([]string, 0, 32)
	if err := filepath.WalkDir(root, func(path string, _ fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		entries = append(entries, path)
		return nil
	}); err != nil {
		return nil, err
	}

	sort.Strings(entries)
	return entries, nil
}

func addEntry(tw *tar.Writer, root, baseName, path string, executables map[string]struct{}) error {
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}

	rel, err := filepath.Rel(root, path)
	if err != nil {
		return err
	}

	archiveName := filepath.ToSlash(baseName)
	if rel != "." {
		archiveName = filepath.ToSlash(filepath.Join(baseName, rel))
	}
	if info.IsDir() {
		archiveName += "/"
	}

	linkTarget := ""
	if info.Mode()&os.ModeSymlink != 0 {
		linkTarget, err = os.Readlink(path)
		if err != nil {
			return err
		}
	}

	header, err := tar.FileInfoHeader(info, linkTarget)
	if err != nil {
		return err
	}
	header.Name = archiveName
	header.ModTime = info.ModTime().UTC()
	header.AccessTime = header.ModTime
	header.ChangeTime = header.ModTime
	header.Uid = 0
	header.Gid = 0
	header.Uname = ""
	header.Gname = ""
	header.Mode = archiveMode(rel, info, executables)

	if err := tw.WriteHeader(header); err != nil {
		return err
	}

	if !info.Mode().IsRegular() {
		return nil
	}

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(tw, file)
	return err
}

func archiveMode(rel string, info os.FileInfo, executables map[string]struct{}) int64 {
	if info.IsDir() {
		return 0o755
	}

	normalized := filepath.ToSlash(filepath.Clean(rel))
	if _, ok := executables[normalized]; ok {
		return 0o755
	}

	if strings.HasSuffix(strings.ToLower(normalized), ".exe") {
		return 0o755
	}

	return 0o644
}
