package util

import (
	"encoding/json"
	"io"
	"net/http"
)

func EnableCors(res http.ResponseWriter, origin string) {
	res.Header().Set("Access-Control-Allow-Origin", origin)
	res.Header().Set("Access-Control-Allow-Methods", "*")
	res.Header().Set("Access-Control-Allow-Headers", "*")
}

func ReadJSONBody(body io.ReadCloser, pointer any) error {
	decoded := json.NewDecoder(body)
	return decoded.Decode(&pointer)
}

func Test[T string](pointer T) {

}
