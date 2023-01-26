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
		var u user.UserJSON
		err := json.NewDecoder(r.Body).Decode(&u)
		
		if err != nil {
			encodeResult(w, err)
			return
		}

		result, err := svc.CreateUser(r.Context(), u.ToUser())

		if err != nil {
			EncodeJSONError(r.Context(), err, w)
			return
		}

		var userJSON *user.UserJSON

		userJSON = user.NewJSONUser(result.User)
		userJSON.NewJWTToken = result.Token
		encodeResult(w, userJSON)
	}
}

func CreatePostLoginHandler(svc user.Service) http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request) {		
		var req user.LoginRequestJSON
		err := json.NewDecoder(r.Body).Decode(&req)
		
		if err != nil {
			encodeResult(w, err)
			return
		}

		result, err := svc.Login(r.Context(), req.Email, req.Password)

		if err != nil {
			EncodeJSONError(r.Context(), err, w)
			return
		}

		var userJSON *user.UserJSON

		userJSON = user.NewJSONUser(result.User)
		userJSON.NewJWTToken = result.Token
		encodeResult(w, userJSON)
	}
}