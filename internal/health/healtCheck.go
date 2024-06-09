package health

import "net/http"

func Handler(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusOK)
	_, err := writer.Write([]byte("UP"))
	if err != nil {
		return
	}
}
