package errorHandler

import "fmt"

const (
	DatabaseError              = 1001
	DatabaseNotFoundError      = 1002
	InvalidAuthenticationError = 1003
	InvalidProductError        = 1004
	InternalError              = 1005
	CustomerAlreadyExist       = 1006
	InvalidRequest             = 1007
	InvalidTransactionPin      = 1008
	CustomerBlockedError       = 1009
)

var (
	errorTypes = map[int]string{
		DatabaseError:              "DatabaseError",
		DatabaseNotFoundError:      "DatabaseNotFoundError",
		InvalidAuthenticationError: "InvalidAuthenticationError",
		InvalidProductError:        "InvalidProductError",
		InternalError:              "InternalError",
		CustomerAlreadyExist:       "CustomerAlreadyExist",
		InvalidRequest:             "InvalidRequest",
		InvalidTransactionPin:      "InvalidTransactionPassword",
		CustomerBlockedError:       "CustomerBlockedError",
	}

	errorMessages = map[int]string{
		DatabaseError:              "failed to handle request at this time due to technical issues. Please retry",
		DatabaseNotFoundError:      "model not found",
		InvalidAuthenticationError: "invalid authentication token provided",
		InvalidProductError:        "invalid product id",
		InternalError:              "unable to process the request at this time",
		CustomerAlreadyExist:       "customer already exists",
		InvalidRequest:             "invalid request parameters",
		InvalidTransactionPin:      "Invalid transaction password",
		CustomerBlockedError:       "customer blocked error",
	}
)

func Type(code int) string {
	if value, ok := errorTypes[code]; ok {
		return value
	}
	return "UnKnownError"
}

func Message(code int) string {
	if value, ok := errorMessages[code]; ok {
		return value
	}
	return "unknown"
}

func Format(code int, err error) error {
	return NewHerror(code, Type(code), Message(code), fmt.Sprintf("%s: %v", Message(code), err))
}
