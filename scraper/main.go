package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
)

var (
	username   = env("USERNAME", "someuser")
	password   = env("PASSWORD", "somepass")
	listen     = env("LISTEN", "0.0.0.0:443")
	baseURL    = env("BASE_URL", "http://127.0.0.1:3000")
	convertURL = env("CONVERT_URL", "http://127.0.0.1:8000")
	sslCert    = env("SSL_CRT", "ssl/service.crt")
	sslKey     = env("SSL_KEY", "ssl/service.key")
	useSSL     = env("USE_SSL", "true")
)

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func basicAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, p, ok := r.BasicAuth()
		if !ok || u != username || p != password {
			w.Header().Set("WWW-Authenticate", `Basic realm="scraper"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}

func browserlessContent(targetURL string) (string, error) {
	payload := fmt.Sprintf(`{"code":"module.exports=async function({page,context}){const r=await page.goto(context.url,{waitUntil:'networkidle0',timeout:55000});const d=await page.content();return{data:d,type:'html'}}","context":{"url":%q}}`, targetURL)
	resp, err := http.Post(baseURL+"/function", "application/json", strings.NewReader(payload))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	return string(b), err
}

func handleMarkdown(w http.ResponseWriter, r *http.Request) {
	targetURL := r.URL.Query().Get("url")
	if targetURL == "" {
		http.Error(w, "missing url parameter", http.StatusBadRequest)
		return
	}
	html, err := browserlessContent(targetURL)
	if err != nil {
		http.Error(w, fmt.Sprintf("browserless error: %v", err), http.StatusBadGateway)
		return
	}
	resp, err := http.Post(convertURL+"/convert", "text/html", strings.NewReader(html))
	if err != nil {
		http.Error(w, fmt.Sprintf("convert error: %v", err), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	io.Copy(w, resp.Body)
}

func handleHTML(w http.ResponseWriter, r *http.Request) {
	targetURL := r.URL.Query().Get("url")
	if targetURL == "" {
		http.Error(w, "missing url parameter", http.StatusBadRequest)
		return
	}
	html, err := browserlessContent(targetURL)
	if err != nil {
		http.Error(w, fmt.Sprintf("browserless error: %v", err), http.StatusBadGateway)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, html)
}

func handleLinks(w http.ResponseWriter, r *http.Request) {
	targetURL := r.URL.Query().Get("url")
	if targetURL == "" {
		http.Error(w, "missing url parameter", http.StatusBadRequest)
		return
	}
	payload := fmt.Sprintf(`{"code":"module.exports=async function({page,context}){await page.goto(context.url,{waitUntil:'networkidle0',timeout:55000});const links=await page.evaluate(()=>Array.from(document.querySelectorAll('a[href]')).map(a=>({title:a.textContent.trim(),link:a.href})));return{data:JSON.stringify(links),type:'application/json'}}","context":{"url":%q}}`, targetURL)
	resp, err := http.Post(baseURL+"/function", "application/json", strings.NewReader(payload))
	if err != nil {
		http.Error(w, fmt.Sprintf("browserless error: %v", err), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	w.Header().Set("Content-Type", "application/json")
	io.Copy(w, resp.Body)
}

func handleScreenshot(w http.ResponseWriter, r *http.Request) {
	targetURL := r.URL.Query().Get("url")
	if targetURL == "" {
		http.Error(w, "missing url parameter", http.StatusBadRequest)
		return
	}
	fullPage := r.URL.Query().Get("fullPage") == "true"
	payload := fmt.Sprintf(`{"url":%q,"options":{"fullPage":%t,"type":"png"},"gotoOptions":{"waitUntil":"networkidle0","timeout":55000}}`, targetURL, fullPage)
	resp, err := http.Post(baseURL+"/screenshot", "application/json", strings.NewReader(payload))
	if err != nil {
		http.Error(w, fmt.Sprintf("browserless error: %v", err), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	w.Header().Set("Content-Type", "image/png")
	io.Copy(w, resp.Body)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/markdown", basicAuth(handleMarkdown))
	mux.HandleFunc("/html", basicAuth(handleHTML))
	mux.HandleFunc("/links", basicAuth(handleLinks))
	mux.HandleFunc("/screenshot", basicAuth(handleScreenshot))
	mux.HandleFunc("/content", basicAuth(handleHTML))

	// Fallback: proxy everything else to browserless
	target, _ := url.Parse(baseURL)
	proxy := httputil.NewSingleHostReverseProxy(target)
	mux.HandleFunc("/", basicAuth(func(w http.ResponseWriter, r *http.Request) {
		proxy.ServeHTTP(w, r)
	}))

	log.Printf("scraper proxy listening on %s (ssl=%s)", listen, useSSL)

	if useSSL == "true" {
		srv := &http.Server{
			Addr:    listen,
			Handler: mux,
			TLSConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
			},
		}
		log.Fatal(srv.ListenAndServeTLS(sslCert, sslKey))
	} else {
		log.Fatal(http.ListenAndServe(listen, mux))
	}
}
