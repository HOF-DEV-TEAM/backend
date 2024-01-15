package globalParameters

import (
	"bitbucket.org/hofng/hofApp/infrastructure/library/http_helper"
	"bitbucket.org/hofng/hofApp/infrastructure/library/security"
	"encoding/json"
	"net/http"
)

func UpdateGlobalVariablesHandler(svc Service) http.HandlerFunc {
	return http_helper.NewHTTPHandler(updateGlobalVariablesHandler, svc)
}

func updateGlobalVariablesHandler(w http.ResponseWriter, r *http.Request, svc interface{}) {
	var globalVariables UpdateGlobalParameters
	err := json.NewDecoder(r.Body).Decode(&globalVariables)
	if err != nil {
		http_helper.EncodeJSONError(r.Context(), err, w)
		return
	}

	result, err := svc.(Service).UpdateGlobalVariables(r.Context(), &globalVariables)
	if err != nil {
		http_helper.EncodeJSONError(r.Context(), err, w)
		return
	}

	http_helper.EncodeResult(w, result, http.StatusOK)
}

func GetGlobalVariablesHandler(svc Service) http.HandlerFunc {
	return http_helper.NewHTTPHandler(getGlobalVariablesHandler, svc)
}

func getGlobalVariablesHandler(w http.ResponseWriter, r *http.Request, svc interface{}) {
	_, ok := r.Context().Value(security.JWTClaimsContextKey).(*security.JWTClaim[any])
	if !ok {
		return
	}

	result, err := svc.(Service).GetGlobalVariables(r.Context())
	if err != nil {
		http_helper.EncodeJSONError(r.Context(), err, w)
		return
	}
	http_helper.EncodeResult(w, result, http.StatusOK)

}
