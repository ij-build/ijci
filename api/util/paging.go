package util

import (
	"math"
	"net/http"
	"net/url"
	"strconv"

	"github.com/efritz/response"
	"github.com/go-nacelle/pgutil"
)

const (
	MaxPageSize     = 100
	DefaultPageSize = 5
)

func GetPageMeta(req *http.Request) (*pgutil.PageMeta, response.Response) {
	page, ok := getPage(req.URL.Query())
	if !ok {
		return nil, response.Empty(http.StatusBadRequest)
	}

	pageSize, ok := getPageSize(req.URL.Query())
	if !ok {
		return nil, response.Empty(http.StatusBadRequest)
	}

	return &pgutil.PageMeta{
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func getPage(values url.Values) (int, bool) {
	return extractPageValue(values, "page", 1, math.MaxInt32)
}

func getPageSize(values url.Values) (int, bool) {
	return extractPageValue(values, "per_page", DefaultPageSize, MaxPageSize)
}

func extractPageValue(values url.Values, key string, defaultValue, maxValue int) (int, bool) {
	raw := values.Get(key)
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
