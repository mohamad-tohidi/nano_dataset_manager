package main

import (
	"encoding/json"
	"net/http"
)

type Server struct {
	store       *Store
	datasetsDir string
}

func NewServer(store *Store, datasetsDir string) *Server {
	return &Server{store: store, datasetsDir: datasetsDir}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /datasets/upload", s.handleUpload)
	mux.HandleFunc("POST /datasets/from-hf", s.handleFromHF)
	mux.HandleFunc("GET /datasets", s.handleList)
	mux.HandleFunc("GET /datasets/{id}", s.handleGet)
	mux.HandleFunc("DELETE /datasets/{id}", s.handleDelete)
	mux.HandleFunc("GET /health", s.handleHealth)
	return mux
}

func (s *Server) handleUpload(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 2<<30)

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "invalid form: "+err.Error())
		return
	}

	name := r.FormValue("name")
	if name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "file is required")
		return
	}
	defer file.Close()

	d := &Dataset{
		ID:         newID(),
		Name:       name,
		SourceType: SourceUpload,
		SourceRef:  header.Filename,
		Status:     StatusPending,
		CreatedAt:  now(),
		UpdatedAt:  now(),
	}

	if err := s.store.Insert(d); err != nil {
		writeError(w, http.StatusInternalServerError, "database error")
		return
	}

	if err := saveUpload(s.datasetsDir, d, file); err != nil {
		s.store.UpdateStatus(d.ID, StatusFailed, 0, "", err.Error())
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	s.store.UpdateStatus(d.ID, StatusReady, d.SizeBytes, d.LocalPath, "")
	d.Status = StatusReady

	writeJSON(w, http.StatusCreated, d)
}

func (s *Server) handleFromHF(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RepoID string `json:"repo_id"`
		Token  string `json:"token"`
		Name   string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}

	if req.RepoID == "" || req.Token == "" || req.Name == "" {
		writeError(w, http.StatusBadRequest, "repo_id, token, and name are required")
		return
	}

	d := &Dataset{
		ID:         newID(),
		Name:       req.Name,
		SourceType: SourceHuggingFace,
		SourceRef:  req.RepoID,
		Status:     StatusPending,
		CreatedAt:  now(),
		UpdatedAt:  now(),
	}

	if err := s.store.Insert(d); err != nil {
		writeError(w, http.StatusInternalServerError, "database error")
		return
	}

	if err := downloadFromHF(s.datasetsDir, d, req.Token); err != nil {
		s.store.UpdateStatus(d.ID, StatusFailed, 0, "", err.Error())
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	s.store.UpdateStatus(d.ID, StatusReady, d.SizeBytes, d.LocalPath, "")
	d.Status = StatusReady

	writeJSON(w, http.StatusCreated, d)
}

func (s *Server) handleList(w http.ResponseWriter, r *http.Request) {
	datasets, err := s.store.List()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "database error")
		return
	}
	writeJSON(w, http.StatusOK, datasets)
}

func (s *Server) handleGet(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	d, err := s.store.GetByID(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "database error")
		return
	}
	if d == nil {
		writeError(w, http.StatusNotFound, "dataset not found")
		return
	}
	writeJSON(w, http.StatusOK, d)
}

func (s *Server) handleDelete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	d, err := s.store.GetByID(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "database error")
		return
	}
	if d == nil {
		writeError(w, http.StatusNotFound, "dataset not found")
		return
	}

	removeDataset(d.LocalPath)
	s.store.Delete(id)

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
