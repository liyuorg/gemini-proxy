package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
)

func Listen(w http.ResponseWriter, r *http.Request) {
	token := os.Getenv("MOL_TOKEN")
	reqURL := os.Getenv("REQ_URL")
	uri := r.RequestURI
	header := r.Header
	tk := header["Mol"]
	if len(tk) == 0 || tk[0] != token {
		w.WriteHeader(http.StatusUnauthorized)
		RespJSON(w, map[string]string{
			"error": "Unauthorized",
		})
		return
	}
	if uri == "/" || uri == "/index.html" {
		RespJSON(w, map[string]string{
			"status": "running",
		})
		return
	}
	geminiProxy(w, r, reqURL)
}

func RespJSON(w http.ResponseWriter, m any) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(m)
}

func RespErr(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(map[string]string{
		"error": err.Error(),
	})
}

func geminiProxy(w http.ResponseWriter, r *http.Request, reqURL string) {
	url := r.URL
	buf, err := io.ReadAll(r.Body)
	if err != nil {
		RespErr(w, err)
		return
	}
	query := url.Query()
	url.RawQuery = query.Encode()
	url.Host = reqURL
	url.Scheme = "https"
	req, err := http.NewRequest(r.Method, url.String(), bytes.NewReader(buf))
	if err != nil {
		RespErr(w, err)
		return
	}
	for k, v := range r.Header {
		if k == "Mol" {
			continue
		}
		for _, vv := range v {
			req.Header.Add(k, vv)
		}
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		RespErr(w, err)
		return
	}
	defer resp.Body.Close()
	StreamData(resp.Body, w)
}

func StreamData(src io.Reader, dst io.Writer) error {
	buf := make([]byte, 32) // Reusable buffer
	_, err := io.CopyBuffer(dst, src, buf)
	return err
}
