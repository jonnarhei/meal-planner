package main

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func getIntParamUrl(r *http.Request, paramName string) (int64, bool) {
	idParam := chi.URLParam(r, paramName)
	paramInt, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return 0, false
	}

	return paramInt, true
}
