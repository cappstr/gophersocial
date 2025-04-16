package apiserver

import (
	"errors"
	"log/slog"
	"net/http"
)

type PostRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

func (req PostRequest) Validate() error {
	if req.Title == "" {
		return errors.New("title is required")
	}
	if req.Content == "" {
		return errors.New("content is required")
	}
	return nil
}

func (s *ApiServer) CreatePostHandler(w http.ResponseWriter, r *http.Request) {
	req, err := Decode[PostRequest](r)
	if err != nil {
		slog.Error("failed to decode request", "err", err)
		w.WriteHeader(http.StatusBadRequest)
	}
	_, err = s.store.Posts.CreatePost(r.Context(), req.Title, req.Content)
	if err != nil {
		slog.Error("failed to create post", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
	}

	if err := Encode[ApiResponse[struct{}]](ApiResponse[struct{}]{
		Message: "successfully created post",
	}, w, http.StatusCreated); err != nil {
		slog.Error("failed to encode response", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}
