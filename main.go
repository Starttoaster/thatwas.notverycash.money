package main

import (
	"embed"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

//go:embed cash.avif
//go:embed cash-small.avif
//go:embed index.html
//go:embed robots.txt
var static embed.FS

func main() {
	http.HandleFunc("/", maxBytesMiddle(pathCheckMiddle(allowedMethodsMiddle(commonHeadersMiddle(indexHandler)))))
	http.HandleFunc("/ofyou", maxBytesMiddle(pathCheckMiddle(allowedMethodsMiddle(commonHeadersMiddle(indexHandler)))))
	http.HandleFunc("/image.avif", maxBytesMiddle(pathCheckMiddle(allowedMethodsMiddle(commonHeadersMiddle(imageHandler)))))
	http.HandleFunc("/image-small.avif", maxBytesMiddle(pathCheckMiddle(allowedMethodsMiddle(commonHeadersMiddle(smallImageHandler)))))
	http.HandleFunc("/robots.txt", maxBytesMiddle(pathCheckMiddle(allowedMethodsMiddle(commonHeadersMiddle(robotsHandler)))))

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", 8080),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       15 * time.Second,
	}
	log.Printf("Server listening on %s...\n", srv.Addr)
	log.Fatal(srv.ListenAndServe())
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
			http.Error(w, "Invalid path", http.StatusBadRequest)
			return
		}
		next.ServeHTTP(w, r)
	}
}

func allowedMethodsMiddle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
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

func indexHandler(w http.ResponseWriter, _ *http.Request) {
	content, err := static.ReadFile("index.html")
	if err != nil {
		http.Error(w, "Error loading page", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	_, err = w.Write(content)
	if err != nil {
		http.Error(w, "Error loading page", http.StatusInternalServerError)
		return
	}
}

func robotsHandler(w http.ResponseWriter, _ *http.Request) {
	content, err := static.ReadFile("robots.txt")
	if err != nil {
		http.Error(w, "Error loading page", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	_, err = w.Write(content)
	if err != nil {
		http.Error(w, "Error loading page", http.StatusInternalServerError)
		return
	}
}

func imageHandler(w http.ResponseWriter, _ *http.Request) {
	content, err := static.ReadFile("cash.avif")
	if err != nil {
		http.Error(w, "Error loading image", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/avif")
	w.Header().Set("Cache-Control", "public, max-age=31536000")
	_, err = w.Write(content)
	if err != nil {
		http.Error(w, "Error loading page", http.StatusInternalServerError)
		return
	}
}

func smallImageHandler(w http.ResponseWriter, _ *http.Request) {
	content, err := static.ReadFile("cash-small.avif")
	if err != nil {
		http.Error(w, "Error loading image", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/avif")
	w.Header().Set("Cache-Control", "public, max-age=31536000")
	_, err = w.Write(content)
	if err != nil {
		http.Error(w, "Error loading page", http.StatusInternalServerError)
		return
	}
}
