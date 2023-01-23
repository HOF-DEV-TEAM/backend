package interfaces

import (
	"encoding/json"
	"net/http"

	"bitbucket.org/hofng/hofApp/pkg/user"
)

// CreateUser godoc
// @Summary Create a new user
// @Description Create a new user with the input paylod
// @Tags users
// @Accept  json
// @Produce  json
// @Param user body User true "Create user"
// @Success 200 {object} User
// @Router /user [post]


func CreateGetUserHandler(svc user.Service) http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request) {		
		var user user.User
		err := json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			encodeResult(w, err)
			return
		}

		svc.CreateUser(r.Context(), &user)
		if err != nil {
			encodeResult(w, err)
			return
		}
		encodeResult(w, user)
	}
}