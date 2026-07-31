package response

import (
	"encoding/json"
	"net/http"
)

// JSON writes v to w as a JSON response. When pretty is true the output
// is indented for human readability.
func JSON(w http.ResponseWriter, status int, v any, pretty bool) {
	var (
		b   []byte
		err error
	)
	if pretty {
		b, err = json.MarshalIndent(v, "", "  ")
	} else {
		b, err = json.Marshal(v)
	}
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"failed to encode response"}`))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(b)
}

// Error writes a JSON error response with the given status and message.
func Error(w http.ResponseWriter, status int, message string) {
	JSON(w, status, map[string]string{"error": message}, false)
}
