package apiserver

import (
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
)

type SignUpRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (req SignUpRequest) Validate() error {
	if req.Username == "" {
		return errors.New("username is required")
	}
	if req.Email == "" {
		return errors.New("email is required")
	}
	if req.Password == "" {
		return errors.New("password is required")
	}
	return nil
}

func (s *ApiServer) SignUpHandler(w http.ResponseWriter, r *http.Request) {
	req, err := Decode[SignUpRequest](r)
	if err != nil {
		slog.Info("failed to decode request", "err", err)
		w.WriteHeader(http.StatusBadRequest)
	}

	existingEmailUser, err := s.store.User.GetUserByEmail(r.Context(), req.Email)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if existingEmailUser != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	_, err = s.store.User.CreateUser(r.Context(), req.Username, req.Email, req.Password)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	if err := Encode[ApiResponse[struct{}]](ApiResponse[struct{}]{
		Message: "successfully signed up user",
	}, w, http.StatusCreated); err != nil {
		slog.Error("failed to encode response", "err", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
	}
}

type SignInRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type SignInResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func (req SignInRequest) Validate() error {
	if req.Email == "" {
		return errors.New("email is required")
	}
	if req.Password == "" {
		return errors.New("password is required")
	}
	return nil
}

func (s *ApiServer) SignInHandler(w http.ResponseWriter, r *http.Request) {
	req, err := Decode[SignInRequest](r)
	if err != nil {
		slog.Info("failed to decode request", "err", err)
		w.WriteHeader(http.StatusBadRequest)
	}

	user, err := s.store.User.GetUserByEmail(r.Context(), req.Email)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if err := user.CheckHashedPassword(req.Password); err != nil {
		slog.Error("failed to check hashed password", "err", err)
		w.WriteHeader(http.StatusUnauthorized)
	}

	tokenPair, err := s.jwtManager.GenerateTokenPair(user.Id)
	if err != nil {
		slog.Error("failed to generate token pair", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
	}

	_, err = s.store.Refresh.DeleteToken(r.Context(), user.Id)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		slog.Error("failed to delete refresh token", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
	}

	_, err = s.store.Refresh.CreateToken(r.Context(), user.Id, tokenPair.RefreshToken)
	if err != nil {
		slog.Error("failed to create refresh token", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
	}

	if err := Encode(ApiResponse[SignInResponse]{
		Data: &SignInResponse{
			AccessToken:  tokenPair.AccessToken.Raw,
			RefreshToken: tokenPair.RefreshToken.Raw,
		},
		Message: "bearer access",
	}, w, http.StatusOK); err != nil {
		slog.Error("failed to encode response", "err", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
	}
}
