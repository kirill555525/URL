package controllers_test

import (
	"context"
	"encoding/json"
	"github.com/kirill555525/URL/cmd/config"
	"github.com/kirill555525/URL/internal/controllers"
	"github.com/kirill555525/URL/internal/logger"
	"github.com/kirill555525/URL/internal/models"
	"github.com/kirill555525/URL/internal/store/memory"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBaseController(t *testing.T) {

	cfg := config.Init()
	err := logger.Initialize(cfg.FlagLogLevel)
	require.NoError(t, err)

	storage := memory.NewStore()
	defer storage.Close()

	controller := controllers.NewBaseController(storage)

	server := httptest.NewServer(controller.Route())
	defer server.Close()

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // Остановить редиректы
		},
	}

	urlToShorten := "https://www.google.ru/"

	testsCreateShortURL := []struct {
		name         string
		method       string
		body         string
		expectedCode int
		path         string
	}{
		{
			name:         "createShortURL NEW",
			method:       http.MethodPost,
			body:         urlToShorten,
			expectedCode: http.StatusCreated,
		},
		{
			name:         "createShortURL NO BODY",
			method:       http.MethodPost,
			body:         "",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "createShortURL ALREADY EXIST",
			method:       http.MethodPost,
			body:         urlToShorten,
			expectedCode: http.StatusConflict,
		},
		{
			name:         "createShortURL NO POST",
			method:       http.MethodGet,
			body:         "",
			expectedCode: http.StatusMethodNotAllowed,
		},
		{
			name:         "createShortURLJSON NEW",
			method:       http.MethodPost,
			body:         `{"url": "https://ya.ru/"}`,
			expectedCode: http.StatusCreated,
			path:         "/api/shorten",
		},
		{
			name:         "createShortURLJSON NO BODY",
			method:       http.MethodPost,
			body:         "",
			expectedCode: http.StatusBadRequest,
			path:         "/api/shorten",
		},
		{
			name:         "createShortURLJSON ALREADY EXIST",
			method:       http.MethodPost,
			body:         `{"url": "https://ya.ru/"}`,
			expectedCode: http.StatusConflict,
			path:         "/api/shorten",
		},
	}

	for _, test := range testsCreateShortURL {
		t.Run(test.name, func(t *testing.T) {
			req, err := http.NewRequest(test.method, server.URL+test.path, strings.NewReader(test.body))
			require.NoError(t, err)
			req.Header.Set("Accept-Encoding", "identity") // чтоб отключить gzip middleware
			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			require.Equal(t, test.expectedCode, resp.StatusCode)

			if resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusConflict {
				body, err := io.ReadAll(resp.Body)
				require.NoError(t, err)
				res := string(body)

				if test.path == "/api/shorten" {
					var obj models.Response
					err = json.Unmarshal(body, &obj)
					require.NoError(t, err)
					res = obj.ShortURL
				}

				l := len(res)
				ok := l > 8
				require.True(t, ok)

				req, err := http.NewRequest(http.MethodGet, server.URL+"/"+res[l-8:], strings.NewReader(res))
				require.NoError(t, err)
				resp, err := client.Do(req)
				require.NoError(t, err)
				defer resp.Body.Close()
				require.Equal(t, http.StatusTemporaryRedirect, resp.StatusCode)

				location := test.body

				if test.path == "/api/shorten" {
					var obj models.Request
					err = json.Unmarshal([]byte(test.body), &obj)
					require.NoError(t, err)
					location = obj.URL
				}

				require.Equal(t, location, resp.Header.Get("Location"))

			}
		})
	}

	testsGetOriginalURL := []struct {
		name         string
		method       string
		expectedCode int
	}{
		{
			name:         "getOriginalURL NOT EXIST",
			method:       http.MethodGet,
			expectedCode: http.StatusNotFound,
		},
		{
			name:         "getOriginalURL EXIST",
			method:       http.MethodGet,
			expectedCode: http.StatusTemporaryRedirect,
		},
		{
			name:         "getOriginalURL NO GET",
			method:       http.MethodPost,
			expectedCode: http.StatusMethodNotAllowed,
		},
	}

	for _, test := range testsGetOriginalURL {
		t.Run(test.name, func(t *testing.T) {
			path := "/12345678"
			if test.expectedCode == http.StatusTemporaryRedirect {
				id, err := storage.GetShortID(context.Background(), urlToShorten)
				require.NoError(t, err)
				path = "/" + id
			}
			req, err := http.NewRequest(test.method, server.URL+path, nil)
			require.NoError(t, err)
			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()
			require.Equal(t, test.expectedCode, resp.StatusCode)
			if test.expectedCode == http.StatusTemporaryRedirect {
				require.Equal(t, urlToShorten, resp.Header.Get("Location"))
			}
		})
	}

	testsCreateShortURLJSONList := []struct {
		name       string
		body       string
		statusCode int
	}{
		{
			name:       "CreateShortURLJSONList SUCCESS",
			body:       bodyCreateShortURLJSONListSUCCESS,
			statusCode: http.StatusCreated,
		},
		{
			name:       "CreateShortURLJSONList SUCCESS AND EXIST",
			body:       bodyCreateShortURLJSONListSUCCESSANDEXIST,
			statusCode: http.StatusCreated,
		},
		{
			name:       "CreateShortURLJSONList invalid JSON array end",
			body:       bodyCreateShortURLJSONListInvalidJSONArrayENDSTRING,
			statusCode: http.StatusBadRequest,
		},
		{
			name:       "CreateShortURLJSONList EMPTY LIST",
			body:       "[]",
			statusCode: http.StatusBadRequest,
		},
		{
			name:       "CreateShortURLJSONList MORE 1000",
			body:       bodyCreateShortURLJSONListMORE1000,
			statusCode: http.StatusCreated,
		},
	}

	for _, test := range testsCreateShortURLJSONList {
		t.Run(test.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodPost, server.URL+`/api/shorten/batch`, strings.NewReader(test.body))
			require.NoError(t, err)
			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()
			require.Equal(t, test.statusCode, resp.StatusCode)
			if resp.StatusCode == http.StatusCreated {

				var objects []models.RequestBatchURL
				err = json.Unmarshal([]byte(test.body), &objects)
				require.NoError(t, err)

				body, err := io.ReadAll(resp.Body)
				require.NoError(t, err)
				var res []models.ResponseBatchURL
				err = json.Unmarshal(body, &res)
				require.NoError(t, err)

				for i := range res {

					l := len(res[i].ShortURL)

					ok := l > 8
					require.True(t, ok)

					getReq, err := http.NewRequest(http.MethodGet, server.URL+"/"+res[i].ShortURL[l-8:], nil)
					require.NoError(t, err)
					resp, err := client.Do(getReq)
					require.NoError(t, err)
					defer resp.Body.Close()
					require.Equal(t, http.StatusTemporaryRedirect, resp.StatusCode)
					require.Equal(t, objects[i].OriginalURL, resp.Header.Get("Location"))
				}

			}
		})
	}

}
