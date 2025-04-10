package main

import (
	"github.com/kirill555525/URL/cmd/config"
	"github.com/kirill555525/URL/internal/logger"
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

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // Остановить редиректы
		},
	}

	testsPost := []struct {
		name         string
		body         string
		expectedCode int
	}{
		{
			name: "SUCCESS #1", body: `https://metanit.com/go/tutorial/9.5.php`, expectedCode: http.StatusCreated,
		},
		{
			name: "SUCCESS #2", body: "", expectedCode: http.StatusBadRequest,
		},
		{
			name: "SUCCESS #3", body: `https://metanit.com/go/tutorial/9.5.php`, expectedCode: http.StatusCreated,
		},
	}

	for _, tt := range testsPost {
		t.Run(tt.name, func(t *testing.T) {
			request, err := http.NewRequest(http.MethodPost, server.URL+"/", strings.NewReader(tt.body))
			require.NoError(t, err)

			res, err := client.Do(request)
			require.NoError(t, err)

			defer res.Body.Close()

			require.Equal(t, tt.expectedCode, res.StatusCode, "Ошибка в POST запросе")

			if tt.expectedCode == http.StatusCreated {
				body, err := io.ReadAll(res.Body)

				require.NoError(t, err, "Ошибка в теле ответа POST запроса")

				id := string(body[len(body)-8:])

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
