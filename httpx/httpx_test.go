package httpx

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHTTPStatusMapping(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{"invalid", Invalid("bad"), http.StatusBadRequest},
		{"unauthorized", Unauthorized("no"), http.StatusUnauthorized},
		{"forbidden", Forbidden("no"), http.StatusForbidden},
		{"not_found", NotFound("gone"), http.StatusNotFound},
		{"conflict", Conflict("dup"), http.StatusConflict},
		{"internal", Internal("boom", nil), http.StatusInternalServerError},
		{"wrapped unknown", errors.New("x"), http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HTTPStatus(tt.err); got != tt.want {
				t.Fatalf("HTTPStatus = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestAsErrorUnwrap(t *testing.T) {
	base := errors.New("db down")
	err := Internal("query failed", base)
	wrapped := Wrap(err, CodeInternal, "service failed")

	got, ok := AsError(wrapped)
	if !ok {
		t.Fatal("expected *Error")
	}
	if got.Code != CodeInternal {
		t.Fatalf("code = %s", got.Code)
	}
	if !errors.Is(wrapped, base) {
		t.Fatal("expected wrap chain to reach base")
	}
}

func TestRespondOKEnvelope(t *testing.T) {
	rec := httptest.NewRecorder()
	RespondOK(rec, map[string]string{"id": "1"})

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["ok"] != true {
		t.Fatalf("ok = %v", body["ok"])
	}
}

func TestRespondErrEnvelope(t *testing.T) {
	rec := httptest.NewRecorder()
	err := WithDetail(NotFound("task not found"), "id", "9")
	RespondErr(rec, err)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d", rec.Code)
	}
	var body struct {
		OK    bool `json:"ok"`
		Error struct {
			Code    string         `json:"code"`
			Message string         `json:"message"`
			Detail  map[string]any `json:"detail"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.OK || body.Error.Code != "not_found" || body.Error.Detail["id"] != "9" {
		t.Fatalf("unexpected body: %+v", body)
	}
}

func TestDecodeJSONStrict(t *testing.T) {
	type in struct {
		Title string `json:"title"`
	}
	req := httptest.NewRequest(http.MethodPost, "/api/tasks", strings.NewReader(`{"title":"x","nope":1}`))
	var dst in
	err := DecodeJSON(req, &dst)
	if err == nil {
		t.Fatal("expected unknown field error")
	}
	if e, ok := AsError(err); !ok || e.Code != CodeInvalid {
		t.Fatalf("want invalid error, got %v", err)
	}
}

func TestChainOrder(t *testing.T) {
	var order []string
	mw := func(name string) Middleware {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, name)
				next.ServeHTTP(w, r)
			})
		}
	}
	h := Chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		order = append(order, "handler")
	}), mw("a"), mw("b"))
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
	want := []string{"a", "b", "handler"}
	if strings.Join(order, ",") != strings.Join(want, ",") {
		t.Fatalf("order = %v, want %v", order, want)
	}
}
