package rest

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
	"strconv"
)

type V1Response struct {
	Data    any      `json:"data"`
	Success bool     `json:"success"`
	Errors  []string `json:"errors"`
}

type Pagination struct {
	Offset int64 `json:"offset"`
	Limit  int64 `json:"limit"`
}

const PaginationMaxLimit int64 = 100

func pagination(r *http.Request) Pagination {
	p := Pagination{
		Offset: 0,
		Limit:  100,
	}
	limit := r.URL.Query().Get("limit")
	if limit != "" {
		if value, err := strconv.ParseInt(limit, 10, 64); err == nil {
			p.Limit = value
		} else {
			p.Limit = PaginationMaxLimit
		}
	}
	offset := r.URL.Query().Get("offset")
	if offset != "" {
		if value, err := strconv.ParseInt(offset, 10, 64); err == nil {
			p.Offset = value
		} else {
			p.Offset = 0
		}
	}

	if p.Limit > PaginationMaxLimit || p.Limit < 1 {
		p.Limit = PaginationMaxLimit
	}

	return p
}

func webhookId(r *http.Request) (string, error) {
	projectId := chi.URLParam(r, WEBHOOK_ID_KEY)
	if len(projectId) != 36 {
		return projectId, ErrBadRequest.WithError(errors.New("invalid webhook id"))
	}
	return projectId, nil
}

func readJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	var maxBytes int64 = 1_000_000
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
	err := json.NewDecoder(r.Body).Decode(dst)
	if err != nil {
		syntaxErr := &json.SyntaxError{}
		unmarshalErr := &json.UnmarshalTypeError{}
		maxBytesErr := &http.MaxBytesError{}
		switch {
		case errors.As(err, &syntaxErr):
			return fmt.Errorf("invalid JSON at character %d", syntaxErr.Offset)
		case errors.Is(err, io.ErrUnexpectedEOF):
			return fmt.Errorf("malformed JSON")
		case errors.As(err, &unmarshalErr):
			if unmarshalErr.Field != "" {
				return fmt.Errorf("invalid JSON type at %s", unmarshalErr.Field)
			}
			return fmt.Errorf("invalid JSON type at %d", unmarshalErr.Offset)
		case errors.Is(err, io.EOF):
			return fmt.Errorf("malformed JSON")
		case errors.Is(err, &json.InvalidUnmarshalError{}):
			panic(err)
		case errors.As(err, &maxBytesErr):
			return fmt.Errorf("body must not be larger than %d bytes", maxBytes)
		default:
			return err
		}
	}
	return nil
}

func writeOK(w http.ResponseWriter, status int, data any) error {
	return writeJSON(w, status, data, true, nil)
}

func writeErr(w http.ResponseWriter, data any, err error) error {
	var apiError *APIError
	ok := errors.As(err, &apiError)
	if !ok {
		apiError = translateError(err)
	}
	return writeJSON(w, apiError.Status, data, false, []string{apiError.Error()})
}

func writeJSON(w http.ResponseWriter, status int, data any, success bool, errors []string) error {
	if errors == nil {
		errors = make([]string, 0)
	}
	resp := V1Response{
		Data:    data,
		Success: success,
		Errors:  errors,
	}
	j, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	w.WriteHeader(status)
	_, err = w.Write(j)
	return err
}
