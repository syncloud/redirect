package rest

import (
	"fmt"
	"net/http"
	"regexp"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

const imageReleaseUrl = "https://github.com/syncloud/image/releases/download/%s/syncloud-%s-%s.%s.xz"

var (
	boardPattern   = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{0,30}[a-z0-9])?$`)
	versionPattern = regexp.MustCompile(`^[0-9]{2}\.[0-9]{2}\.[0-9]{2}$`)
	imageFormats   = map[string]bool{"img": true, "vdi": true}
)

func (w *Www) Download(writer http.ResponseWriter, req *http.Request) {
	board := mux.Vars(req)["board"]
	version := req.URL.Query().Get("version")
	format := req.URL.Query().Get("format")
	if format == "" {
		format = "img"
	}
	if !boardPattern.MatchString(board) || !versionPattern.MatchString(version) || !imageFormats[format] {
		w.logger.Info("download rejected",
			zap.String("board", board), zap.String("version", version),
			zap.String("format", format))
		http.Error(writer, "unknown image", http.StatusNotFound)
		return
	}

	source := "direct"
	if req.URL.Query().Get("gclid") != "" {
		source = "ad"
	}
	w.metrics.Download(board, source)
	w.logger.Info("download", zap.String("board", board),
		zap.String("version", version), zap.String("format", format),
		zap.String("source", source))

	http.Redirect(writer, req,
		fmt.Sprintf(imageReleaseUrl, version, board, version, format), http.StatusFound)
}
