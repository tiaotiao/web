package web

import (
	"errors"
	// "fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestResponser(t *testing.T) {
	type T struct {
		A string `json:"A"`
		B string `json:"B"`
	}

	testCases := []struct {
		Result interface{}
		Body   string
		Code   int
	}{
		{nil, "", http.StatusOK},
		{"test1", "test1", http.StatusOK},
		{[]byte("test2"), "test2", http.StatusOK},
		{errors.New("test3"), `{"error":"server error","message":"test3"}`, http.StatusInternalServerError},
		{T{"a", "b"}, `{"A":"a","B":"b"}`, http.StatusOK},
	}

	for i, tt := range testCases {
		responser := NewDefaultResponser()
		w := httptest.NewRecorder()

		c := new(Context)
		c.ResponseWriter = w

		code, err := responser.Response(c, tt.Result)

		if tt.Body != w.Body.String() {
			t.Errorf("case %v: body = %v; want %v", i, w.Body, tt.Body)
		}
		if err != nil {
			t.Errorf("case %v: err expected nil actual %v", i, err)
		}
		if code != tt.Code {
			t.Errorf("case %v: code expected %v actual %v", i, tt.Code, code)
			t.Log(tt.Body)
		}
	}
}
