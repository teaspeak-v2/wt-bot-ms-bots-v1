package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/teaspeak-v2/wt-bot-ms-bots-v1/internal/apperror"
)

var validate = validator.New()

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func readJSON(r *http.Request, dst any) error {
	dec := json.NewDecoder(io.LimitReader(r.Body, 2<<20))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return apperror.InvalidRequest("invalid json body", err)
	}
	return validate.Struct(dst)
}

func validationErr(err error) error {
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		return apperror.InvalidRequest(ve.Error(), err)
	}
	return err
}

func parseBoolPtr(s string) (*bool, error) {
	if s == "" {
		return nil, nil
	}
	b, err := strconv.ParseBool(s)
	if err != nil {
		return nil, err
	}
	return &b, nil
}
