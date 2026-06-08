package main

import (
	"crypto/rand"
	"fmt"
	"time"
)

type Dataset struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	SourceType string    `json:"source_type"`
	SourceRef  string    `json:"source_ref"`
	LocalPath  string    `json:"local_path"`
	SizeBytes  int64     `json:"size_bytes"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	ErrorMsg   string    `json:"error_msg,omitempty"`
}

const (
	StatusPending = "pending"
	StatusReady   = "ready"
	StatusFailed  = "failed"

	SourceUpload     = "upload"
	SourceHuggingFace = "huggingface"
)

func newID() string {
	uuid := make([]byte, 16)
	rand.Read(uuid)
	uuid[6] = (uuid[6] & 0x0f) | 0x40
	uuid[8] = (uuid[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:])
}

func now() time.Time {
	return time.Now().UTC()
}
