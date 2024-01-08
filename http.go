package wallpaper

import (
  "slices"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

type WallpaperHandler struct {
	Directory string
	srv       *gin.Engine
}

func (w *WallpaperHandler) Handler() http.Handler {
	return w.srv
}

func NewWallpaperHandler(fromDirectory string) *WallpaperHandler {
	srv := gin.New()
	srv.Use(gin.Logger(), gin.Recovery())

	w := &WallpaperHandler{
		Directory: fromDirectory,
		srv:       srv,
	}

	w.registerRoutes()
	return w
}

func (w *WallpaperHandler) registerRoutes() {
	w.srv.GET("/", getWallpaperHandler(w.Directory))
  w.srv.GET("/healthz", healthHandler)
	w.srv.Static("/w", w.Directory)
}

type errorResponseBody struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

func internalServerError(err string) (int, errorResponseBody) {
	return http.StatusInternalServerError, errorResponseBody{Error: "Internal Server Error", Message: err}
}


func healthHandler(ctx *gin.Context) {
  ctx.Header("Content-Type", "application/json")
  ctx.Writer.WriteString(`{"status":"ok"}`)
}

func getWallpaperHandler(rootPath string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		root, err := os.Open(rootPath)
		if err != nil {
			ctx.JSON(internalServerError(err.Error()))
			return
		}

		entries, err := root.ReadDir(0)
		if err != nil {
			ctx.JSON(internalServerError(err.Error()))
			return
		}

		amountOfEntries := len(entries)
		if amountOfEntries == 0 {
			ctx.AbortWithStatus(http.StatusNoContent)
			return
		}
  
    var picked os.DirEntry

    i := 0
    for picked == nil || inBannedWords(picked.Name()) {
		  rand := getPseudoRandom(i)
      i++
		  picked = entries[rand.Intn(len(entries))]
		  picked.Info()
    }

		ctx.Redirect(302, "/w/"+picked.Name())
	}
}
func getPseudoRandom(offset int) *rand.Rand {
	year, month, date := time.Now().Date()
	seed := (year * 100000) + (int(month) * 1000) + (date *10) + offset
	return rand.New(rand.NewSource(int64(seed)))
}

var bannedEntries = []string{
  "lost+found",
}

func inBannedWords(s string) bool {
  return slices.Contains(bannedEntries, s)
}
