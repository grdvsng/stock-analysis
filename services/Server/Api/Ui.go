package Api

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func ui(root string) http.Handler {
	files := http.FileServer(http.Dir(root))
	index := filepath.Join(root, "index.html")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := filepath.Join(root, filepath.Clean("/"+r.URL.Path))

		if info, e := os.Stat(path); e == nil && !info.IsDir() && strings.HasPrefix(path, root) {
			files.ServeHTTP(w, r)

			return
		}

		http.ServeFile(w, r, index)
	})
}
