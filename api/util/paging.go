package util

import (
	"math"
	"net/http"
	"strconv"

	"github.com/efritz/response"
	"github.com/gorilla/mux"

	"github.com/efritz/ijci/api/db"
)

const (
	MaxPageSize     = 100
	DefaultPageSize = 5
)

func GetPageMeta(req *http.Request) (*db.PageMeta, response.Response) {
	page, ok := getPage(mux.Vars(req))
	if !ok {
		return nil, response.Empty(http.StatusBadRequest)
	}

	pageSize, ok := getPageSize(mux.Vars(req))
	if !ok {
		return nil, response.Empty(http.StatusBadRequest)
	}

	return &db.PageMeta{
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func getPage(vars map[string]string) (int, bool) {
	return extractPageValue(vars, "page", 1, math.MaxInt32)
}

func getPageSize(vars map[string]string) (int, bool) {
	return extractPageValue(vars, "per_page", DefaultPageSize, MaxPageSize)
}

func extractPageValue(vars map[string]string, key string, defaultValue, maxValue int) (int, bool) {
	raw := vars[key]
	if raw == "" {
		return defaultValue, true
	}

	val, err := strconv.Atoi(raw)
	if err != nil || val < 1 {
		return 0, false

	}

	if val > maxValue {
		return maxValue, true
	}

	return val, true
}
