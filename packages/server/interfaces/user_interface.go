package interfaces

import (
	"encoding/json"
	"fmt"
	"net/http"

	"bitbucket.org/hofng/hofApp/domain/entity"
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
func (httpHandler *HTTPHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var user entity.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		encodeResult(w, err)
		return
	}
	httpHandler.repo.CreateUser(user)
	if err != nil {
		encodeResult(w, err)
		return
	}
	encodeResult(w, user)
}


func CreateGetUserHandler() http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request) {
		fmt.Println("Gets called here")
		var user entity.User
		err := json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			encodeResult(w, err)
			return
		}
		// httpHandler.repo.CreateUser(user)
		if err != nil {
			encodeResult(w, err)
			return
		}
		encodeResult(w, user)
	}
}