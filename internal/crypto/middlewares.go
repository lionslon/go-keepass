package crypto

import (
	"bytes"
	"fmt"
	"github.com/lionslon/go-keepass/internal/logger"
	"io"
	"net/http"
)

// Middleware работа по проверке подписи запроса и установка подписи в ответ.
func Middleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		//Если инициализирован ключ расшифровывания - расшифровываем входящие данные
		if deprypt != nil {
			buf, _ := io.ReadAll(r.Body)

			logger.Info("encrypt body: %s", string(buf))

			message, err := deprypt.Decrypt(buf)
			if err != nil {
				logger.Error(fmt.Sprintf("cannot decrypt request body: %s", err))
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			r.Body = io.NopCloser(bytes.NewBuffer(message))
		}

		//Вызов целевого handler
		h.ServeHTTP(w, r)
	})
}
