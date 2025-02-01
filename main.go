package main

import (
	"embed"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"
)

//go:embed cash.avif
//go:embed cash-small.avif
//go:embed index.html
//go:embed robots.txt
var static embed.FS

var logger *slog.Logger

func main() {
	logger = slog.New(
		slog.NewTextHandler(
			os.Stdout,
			&slog.HandlerOptions{
				Level: slog.LevelInfo,
			},
		),
	)

	http.HandleFunc("/",
		maxBytesMiddle(
			pathCheckMiddle(
				allowedMethodsMiddle(
					logMiddle(
						commonHeadersMiddle(indexHandler))))))
	http.HandleFunc("/ofyou",
		maxBytesMiddle(
			pathCheckMiddle(
				allowedMethodsMiddle(
					logMiddle(
						commonHeadersMiddle(indexHandler))))))
	http.HandleFunc("/image.avif",
		maxBytesMiddle(
			pathCheckMiddle(
				allowedMethodsMiddle(
					logMiddle(
						commonHeadersMiddle(imageHandler))))))
	http.HandleFunc("/image-small.avif",
		maxBytesMiddle(
			pathCheckMiddle(
				allowedMethodsMiddle(
					logMiddle(
						commonHeadersMiddle(smallImageHandler))))))
	http.HandleFunc("/robots.txt",
		maxBytesMiddle(
			pathCheckMiddle(
				allowedMethodsMiddle(
					logMiddle(
						commonHeadersMiddle(robotsHandler))))))

	srv := &http.Server{
		Addr:              fmt.Sprintf("%s:%d", "0.0.0.0", 8080),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       15 * time.Second,
	}
	logger.Info("starting server", "address", srv.Addr)
	log.Fatal(srv.ListenAndServe())
}

func buildLogArgs(r *http.Request) []any {
	args := []any{
		"remote", r.RemoteAddr,
		"method", r.Method,
		"path", r.URL.Path,
		"user-agent", r.UserAgent(),
	}

	if xff := r.Header.Get("x-forwarded-for"); xff != "" {
		args = append(args, "x-forwarded-for", xff)
	}

	if realIP := r.Header.Get("x-real-ip"); realIP != "" {
		args = append(args, "x-real-ip", realIP)
	}

	return args
}

func logMiddle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info("handling new request", buildLogArgs(r)...)
		next.ServeHTTP(w, r)
	}
}

type handleErrorOptions struct {
	Error      error
	LogMsg     string
	PublicMsg  string
	StatusCode int
}

func handleError(w http.ResponseWriter, r *http.Request, opts handleErrorOptions) {
	args := buildLogArgs(r)
	if opts.Error != nil {
		args = append(args, "error", opts.Error.Error())
		logger.Error(opts.LogMsg, args...)
	} else {
		logger.Info(opts.LogMsg, args...)
	}
	http.Error(w, opts.PublicMsg, opts.StatusCode)
	return
}

func maxBytesMiddle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 1024)
		next.ServeHTTP(w, r)
	}
}

func pathCheckMiddle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "..") {
			handleError(w, r, handleErrorOptions{
				LogMsg:     "blocked - invalid path",
				PublicMsg:  "Invalid path",
				StatusCode: http.StatusBadRequest,
			})
			return
		}
		next.ServeHTTP(w, r)
	}
}

func allowedMethodsMiddle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			handleError(w, r, handleErrorOptions{
				LogMsg:     "blocked - method not allowed",
				PublicMsg:  "Method Not Allowed",
				StatusCode: http.StatusMethodNotAllowed,
			})
			return
		}
		next.ServeHTTP(w, r)
	}
}

func commonHeadersMiddle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; img-src 'self'; style-src 'self' 'unsafe-inline'")
		next.ServeHTTP(w, r)
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	content, err := static.ReadFile("index.html")
	if err != nil {
		handleError(w, r, handleErrorOptions{
			Error:      err,
			LogMsg:     "loading from static content",
			PublicMsg:  "Error loading page",
			StatusCode: http.StatusInternalServerError,
		})
		return
	}

	w.Header().Set("Content-Type", "text/html")
	_, err = w.Write(content)
	if err != nil {
		handleError(w, r, handleErrorOptions{
			Error:      err,
			LogMsg:     "writing to response writer",
			PublicMsg:  "Error loading page",
			StatusCode: http.StatusInternalServerError,
		})
		return
	}
}

func robotsHandler(w http.ResponseWriter, r *http.Request) {
	content, err := static.ReadFile("robots.txt")
	if err != nil {
		handleError(w, r, handleErrorOptions{
			Error:      err,
			LogMsg:     "loading from static content",
			PublicMsg:  "Error loading page",
			StatusCode: http.StatusInternalServerError,
		})
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	_, err = w.Write(content)
	if err != nil {
		handleError(w, r, handleErrorOptions{
			Error:      err,
			LogMsg:     "writing to response writer",
			PublicMsg:  "Error loading page",
			StatusCode: http.StatusInternalServerError,
		})
		return
	}
}

func imageHandler(w http.ResponseWriter, r *http.Request) {
	content, err := static.ReadFile("cash.avif")
	if err != nil {
		handleError(w, r, handleErrorOptions{
			Error:      err,
			LogMsg:     "loading from static content",
			PublicMsg:  "Error loading page",
			StatusCode: http.StatusInternalServerError,
		})
		return
	}

	w.Header().Set("Content-Type", "image/avif")
	w.Header().Set("Cache-Control", "public, max-age=31536000")
	_, err = w.Write(content)
	if err != nil {
		handleError(w, r, handleErrorOptions{
			Error:      err,
			LogMsg:     "writing to response writer",
			PublicMsg:  "Error loading page",
			StatusCode: http.StatusInternalServerError,
		})
		return
	}
}

func smallImageHandler(w http.ResponseWriter, r *http.Request) {
	content, err := static.ReadFile("cash-small.avif")
	if err != nil {
		handleError(w, r, handleErrorOptions{
			Error:      err,
			LogMsg:     "loading from static content",
			PublicMsg:  "Error loading page",
			StatusCode: http.StatusInternalServerError,
		})
		return
	}

	w.Header().Set("Content-Type", "image/avif")
	w.Header().Set("Cache-Control", "public, max-age=31536000")
	_, err = w.Write(content)
	if err != nil {
		handleError(w, r, handleErrorOptions{
			Error:      err,
			LogMsg:     "writing to response writer",
			PublicMsg:  "Error loading page",
			StatusCode: http.StatusInternalServerError,
		})
		return
	}
}
