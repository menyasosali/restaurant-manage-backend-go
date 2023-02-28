package middleware

import (
	"fmt"
	"github.com/menyasosali/restaurant-manage-backend-go/helpers"
	"net/http"
)

func Authentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientToken := r.Header.Get("token")
		if clientToken == "" {
			msg := fmt.Sprintf("No Authorization header provided")
			http.Error(w, msg, http.StatusInternalServerError)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		claims, err := helpers.ValidateToken(clientToken)
		if err != "" {
			http.Error(w, err, http.StatusInternalServerError)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		r.Header.Set("email", claims.Email)
		r.Header.Set("first_name", claims.FirstName)
		r.Header.Set("second_name", claims.SecondName)
		r.Header.Set("uid", claims.Uid)

		next.ServeHTTP(w, r)
	})
}
