package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type hfSibling struct {
	Rfilename string `json:"rfilename"`
}

type hfRepoResponse struct {
	Siblings []hfSibling `json:"siblings"`
}

func downloadFromHF(baseDir string, d *Dataset, token string) error {
	dir := filepath.Join(baseDir, d.ID)
	dataDir := filepath.Join(dir, "data")

	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("create data dir: %w", err)
	}

	client := &http.Client{Timeout: 30 * time.Minute}

	siblings, repoType, err := listHFFiles(client, d.SourceRef, token)
	if err != nil {
		os.RemoveAll(dir)
		return err
	}

	var totalSize int64
	for _, s := range siblings {
		n, err := downloadHFFile(client, d.SourceRef, s.Rfilename, dataDir, token, repoType)
		if err != nil {
			os.RemoveAll(dir)
			return err
		}
		totalSize += n
	}

	d.LocalPath = dir
	d.SizeBytes = totalSize
	return nil
}

func listHFFiles(client *http.Client, repoID, token string) ([]hfSibling, string, error) {
	siblings, err := tryListRepo(client, "datasets", repoID, token)
	if err == nil {
		return siblings, "datasets", nil
	}

	siblings, err = tryListRepo(client, "models", repoID, token)
	if err == nil {
		return siblings, "models", nil
	}

	return nil, "", fmt.Errorf("repository %q not found as dataset or model", repoID)
}

func tryListRepo(client *http.Client, repoType, repoID, token string) ([]hfSibling, error) {
	url := fmt.Sprintf("https://huggingface.co/api/%s/%s", repoType, repoID)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("not found")
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HF API returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var apiResp hfRepoResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("decode HF API response: %w", err)
	}

	if len(apiResp.Siblings) == 0 {
		return nil, fmt.Errorf("no files found")
	}

	return apiResp.Siblings, nil
}

func downloadHFFile(client *http.Client, repoID, filePath, dataDir, token, repoType string) (int64, error) {
	url := fmt.Sprintf("https://huggingface.co/%s/resolve/main/%s", repoID, filePath)
	if repoType == "datasets" {
		url = fmt.Sprintf("https://huggingface.co/datasets/%s/resolve/main/%s", repoID, filePath)
	}

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("download %s: %w", filePath, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("download %s returned %d: %s", filePath, resp.StatusCode, strings.TrimSpace(string(body)))
	}

	target := filepath.Join(dataDir, filePath)
	if !strings.HasPrefix(filepath.Clean(target), filepath.Clean(dataDir)+string(os.PathSeparator)) {
		return 0, fmt.Errorf("illegal path: %s", filePath)
	}

	os.MkdirAll(filepath.Dir(target), 0755)

	out, err := os.Create(target)
	if err != nil {
		return 0, fmt.Errorf("create %s: %w", filePath, err)
	}
	defer out.Close()

	n, err := io.Copy(out, resp.Body)
	if err != nil {
		return 0, fmt.Errorf("write %s: %w", filePath, err)
	}

	return n, nil
}
