package main

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func saveUpload(baseDir string, d *Dataset, src io.Reader) error {
	dir := filepath.Join(baseDir, d.ID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create dataset dir: %w", err)
	}

	rawPath := filepath.Join(dir, "raw.zip")
	f, err := os.Create(rawPath)
	if err != nil {
		os.RemoveAll(dir)
		return fmt.Errorf("create raw.zip: %w", err)
	}

	if _, err := io.Copy(f, src); err != nil {
		f.Close()
		os.RemoveAll(dir)
		return fmt.Errorf("write raw.zip: %w", err)
	}
	f.Close()

	zr, err := zip.OpenReader(rawPath)
	if err != nil {
		os.RemoveAll(dir)
		return fmt.Errorf("invalid zip archive: %w", err)
	}
	zr.Close()

	dataDir := filepath.Join(dir, "data")
	if err := extractZip(rawPath, dataDir); err != nil {
		os.RemoveAll(dir)
		return fmt.Errorf("extract zip: %w", err)
	}

	size, err := dirSize(dataDir)
	if err != nil {
		os.RemoveAll(dir)
		return fmt.Errorf("calculate size: %w", err)
	}

	d.LocalPath = dir
	d.SizeBytes = size
	return nil
}

func extractZip(zipPath, destDir string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		target := filepath.Join(destDir, f.Name)
		if !filepath.HasPrefix(target, filepath.Clean(destDir)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal path: %s", f.Name)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(target, 0755)
			continue
		}

		os.MkdirAll(filepath.Dir(target), 0755)

		out, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return fmt.Errorf("create %s: %w", f.Name, err)
		}

		rc, err := f.Open()
		if err != nil {
			out.Close()
			return fmt.Errorf("open %s in zip: %w", f.Name, err)
		}

		_, err = io.Copy(out, rc)
		out.Close()
		rc.Close()
		if err != nil {
			return fmt.Errorf("write %s: %w", f.Name, err)
		}
	}

	return nil
}

func removeDataset(dir string) error {
	return os.RemoveAll(dir)
}

func dirSize(dir string) (int64, error) {
	var size int64
	err := filepath.Walk(dir, func(_ string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !fi.IsDir() {
			size += fi.Size()
		}
		return nil
	})
	return size, err
}
