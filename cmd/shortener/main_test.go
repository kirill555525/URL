package main

import (
	"encoding/json"
	"github.com/kirill555525/URL/cmd/config"
	"github.com/kirill555525/URL/internal/logger"
	"github.com/kirill555525/URL/internal/models"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestShortenURLHandler(t *testing.T) {

	cfg := config.Init()

	err := logger.Initialize(cfg.FlagLogLevel)
	require.NoError(t, err)

	server := httptest.NewServer(URLRouter())
	defer server.Close()
	t.Log(server.URL)

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // Остановить редиректы
		},
	}

	testsPost := []struct {
		name         string
		body         string
		expectedCode int
		url          string
	}{
		{
			name: "SUCCESS #1", body: `https://metanit.com/go/tutorial/9.5.php`, expectedCode: http.StatusCreated, url: server.URL + "/",
		},
		{
			name: "SUCCESS #2", body: "", expectedCode: http.StatusBadRequest, url: server.URL + "/",
		},
		{
			name: "SUCCESS #3", body: `https://metanit.com/go/tutorial/9.5.php`, expectedCode: http.StatusCreated, url: server.URL + "/",
		},
		{
			name: "SUCCESS #4", body: "", expectedCode: http.StatusInternalServerError, url: server.URL + "/api/shorten",
		},
		{
			name: "SUCCESS #5", body: `{"url": "https://www.google.nl/"}`, expectedCode: http.StatusCreated, url: server.URL + "/api/shorten",
		},
		{
			name: "SUCCESS #6", body: `{"url": "https://www.google.nl/"}`, expectedCode: http.StatusCreated, url: server.URL + "/api/shorten",
		},
	}

	for _, tt := range testsPost {
		t.Run(tt.name, func(t *testing.T) {
			request, err := http.NewRequest(http.MethodPost, tt.url, strings.NewReader(tt.body))
			require.NoError(t, err)

			res, err := client.Do(request)
			require.NoError(t, err)

			defer res.Body.Close()

			require.Equal(t, tt.expectedCode, res.StatusCode, "Ошибка в POST запросе")

			if tt.expectedCode == http.StatusCreated {

				var id string

				if res.Header.Get("Content-Type") == "application/json" {

					var resp models.Request
					require.NoError(t, json.NewDecoder(strings.NewReader(tt.body)).Decode(&resp))
					tt.body = resp.URL

					var body models.Response
					decoder := json.NewDecoder(res.Body)
					err = decoder.Decode(&body)
					require.NoError(t, err, "Ошибка в теле ответа POST запроса")
					id = body.ShortURL[len(body.ShortURL)-8:]

				} else {
					body, err := io.ReadAll(res.Body)

					require.NoError(t, err, "Ошибка в теле ответа POST запроса")

					id = string(body[len(body)-8:])
				}

				require.Equal(t, urlMap[tt.body], id)
				require.Equal(t, idMap[id], tt.body)

				request, err = http.NewRequest(http.MethodGet, server.URL+"/"+id, nil)
				require.NoError(t, err)

				res, err := client.Do(request)

				require.NoError(t, err)
				defer res.Body.Close()

				require.Equal(t, http.StatusTemporaryRedirect, res.StatusCode, "Ошибка в GET запросе")
			}

		})
	}

	testsFail := []struct {
		name         string
		url          string
		expectedCode int
		method       string
	}{
		{
			name: "FAIL GET #1", url: "/", expectedCode: http.StatusMethodNotAllowed, method: http.MethodGet,
		},
		{
			name: "FAIL GET #2", url: "/123", expectedCode: http.StatusNotFound, method: http.MethodGet,
		},
		{
			name: "FAIL GET #3", url: "/acbq236b", expectedCode: http.StatusNotFound, method: http.MethodGet,
		},
		{
			name: "FAIL DELETE #1", url: "/", expectedCode: http.StatusMethodNotAllowed, method: http.MethodDelete,
		},
	}

	for _, tt := range testsFail {
		t.Run(tt.name, func(t *testing.T) {
			request, err := http.NewRequest(tt.method, server.URL+tt.url, nil)
			require.NoError(t, err)

			res, err := client.Do(request)
			require.NoError(t, err)
			defer res.Body.Close()

			require.Equal(t, tt.expectedCode, res.StatusCode)

		})
	}

}
