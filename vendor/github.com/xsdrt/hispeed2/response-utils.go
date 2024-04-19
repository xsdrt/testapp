package hispeed2

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"path"
	"path/filepath"
)

func (h *HiSpeed2) WriteJSON(w http.ResponseWriter, status int, data interface{}, headers ...http.Header) error {
	out, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	if len(headers) > 0 {
		for key, value := range headers[0] {
			w.Header()[key] = value
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, err = w.Write(out)
	if err != nil {
		return err
	}
	return nil
}

func (h *HiSpeed2) WriteXML(w http.ResponseWriter, status int, data interface{}, headers ...http.Header) error {
	out, err := xml.MarshalIndent(data, "", "	")
	if err != nil {
		return err
	}

	if len(headers) > 0 {
		for key, value := range headers[0] {
			w.Header()[key] = value
		}
	}

	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(status)
	_, err = w.Write(out)
	if err != nil {
		return err
	}
	return nil
}

func (h *HiSpeed2) DownloadFile(w http.ResponseWriter, r *http.Request, pathToFile, fileName string) error {
	fp := path.Join(pathToFile, fileName)
	fileToServe := filepath.Clean(fp)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; file=\"%s\"", fileName))
	http.ServeFile(w, r, fileToServe)
	return nil
}

func (h *HiSpeed2) Error404(w http.ResponseWriter, r *http.Request) {
	h.ErrorStatus(w, http.StatusNotFound)
}

func (h *HiSpeed2) Error500(w http.ResponseWriter, r *http.Request) {
	h.ErrorStatus(w, http.StatusInternalServerError)
}

func (h *HiSpeed2) ErrorUnauthorized(w http.ResponseWriter, r *http.Request) {
	h.ErrorStatus(w, http.StatusUnauthorized)
}

func (h *HiSpeed2) ErrorForbidden(w http.ResponseWriter, r *http.Request) {
	h.ErrorStatus(w, http.StatusForbidden)
}

func (h *HiSpeed2) ErrorStatus(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}
