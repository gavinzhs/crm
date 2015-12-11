package main
import "net/http"

const (
    CONTENT_TYPE = "Content-Type"
)

func midTextDefault(w http.ResponseWriter) {
    if w.Header().Get(CONTENT_TYPE) == "" {
        w.Header().Set(CONTENT_TYPE, "text/plain; charset=UTF-8")
    }
}
