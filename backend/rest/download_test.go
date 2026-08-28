package rest

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/syncloud/redirect/metrics"
	"go.uber.org/zap"
)

func downloadRequest(t *testing.T, target string) *httptest.ResponseRecorder {
	w := &Www{metrics: metrics.New(), logger: zap.NewNop()}
	r := mux.NewRouter()
	r.HandleFunc("/download/{board}", w.Download).Methods("GET")
	recorder := httptest.NewRecorder()
	r.ServeHTTP(recorder, httptest.NewRequest("GET", target, nil))
	return recorder
}

func TestDownloadRedirectsToTheImage(t *testing.T) {
	response := downloadRequest(t, "/download/raspberrypi-64?version=26.06.01")
	assert.Equal(t, http.StatusFound, response.Code)
	assert.Equal(t,
		"https://github.com/syncloud/image/releases/download/26.06.01/syncloud-raspberrypi-64-26.06.01.img.xz",
		response.Header().Get("Location"))
}

func TestDownloadServesTheVirtualBoxImage(t *testing.T) {
	response := downloadRequest(t, "/download/amd64?version=26.07.01&format=vdi")
	assert.Equal(t, http.StatusFound, response.Code)
	assert.Equal(t,
		"https://github.com/syncloud/image/releases/download/26.07.01/syncloud-amd64-26.07.01.vdi.xz",
		response.Header().Get("Location"))
}

func TestDownloadRejectsUnknownFormat(t *testing.T) {
	for _, format := range []string{"exe", "img.xz", "../img", "IMG"} {
		response := downloadRequest(t, "/download/amd64?version=26.07.01&format="+format)
		assert.Equal(t, http.StatusNotFound, response.Code, format)
	}
}

func TestDownloadRejectsUnknownVersionFormat(t *testing.T) {
	for _, version := range []string{"", "latest", "26.6.1", "../../etc", "26.06.01/x"} {
		response := downloadRequest(t, "/download/raspberrypi-64?version="+version)
		assert.Equal(t, http.StatusNotFound, response.Code, version)
	}
}

func TestDownloadRejectsBoardOutsideTheCharacterSet(t *testing.T) {
	for _, board := range []string{"Raspberry", "pi_64", "pi.64", "-pi", "pi-", "a%2f..%2fb"} {
		response := downloadRequest(t, "/download/"+board+"?version=26.06.01")
		assert.NotEqual(t, http.StatusFound, response.Code, board)
	}
}

func TestDownloadCannotBeRedirectedOffGithub(t *testing.T) {
	response := downloadRequest(t, "/download/raspberrypi-64?version=26.06.01&url=https://evil.example.com")
	assert.Equal(t, http.StatusFound, response.Code)
	assert.Contains(t, response.Header().Get("Location"), "https://github.com/syncloud/image/")
}
