package httpx

import (
	"encoding/json"
	"net/http"
)

type successResponse struct {
	OK   bool `json:"ok"`
	Data any  `json:"data"`
}

type errorBody struct {
	Code    Code           `json:"code"`
	Message string         `json:"message"`
	Detail  map[string]any `json:"detail,omitempty"`
}

type errorResponse struct {
	OK    bool      `json:"ok"`
	Error errorBody `json:"error"`
}

// RespondJSON writes a success envelope with the given HTTP status.
func RespondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(successResponse{OK: true, Data: data})
}

func RespondOK(w http.ResponseWriter, data any) {
	RespondJSON(w, http.StatusOK, data)
}

func RespondCreated(w http.ResponseWriter, data any) {
	RespondJSON(w, http.StatusCreated, data)
}

func RespondNoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// RespondErr writes the shared error envelope.
func RespondErr(w http.ResponseWriter, err error) {
	e, ok := AsError(err)
	if !ok {
		e = Internal("internal error", err)
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(StatusOf(e.Code))
	_ = json.NewEncoder(w).Encode(errorResponse{
		OK: false,
		Error: errorBody{
			Code:    e.Code,
			Message: e.Message,
			Detail:  e.Detail,
		},
	})
}

// DecodeJSON decodes a request body into dst with strict field checking.
func DecodeJSON(r *http.Request, dst any) error {
	if r.Body == nil {
		return Invalid("request body is required")
	}
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return Wrap(err, CodeInvalid, "invalid JSON body")
	}
	return nil
}
