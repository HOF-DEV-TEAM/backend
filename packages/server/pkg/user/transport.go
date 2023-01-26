package user

import "database/sql"


type UserAndToken struct {
	User 	*User
	Token  	string
}

type AddressJSON string

type UserJSON struct {
	ID  			int 			`json:"id"`
	Username 		string 			`json:"usernmae"`
	Password        string 			`json:"password,omitempty"`
	Email 			string 			`json:"email"`
	FirstName 		string 			`json:"first_name"`
	LastName 		string 			`json:"last_name"`
	Address 		string 			`json:"address,omitempty"`
	Mobile 			string 			`json:"mobile,omitempty"`
	Gender 			string 			`json:"gender,omitempty"`
	IsVerified      IsVerifiedEnum  `json:"is_verified,omitempty"`
	NewJWTToken 	string			`json:"newToken,omitempty"`
}


type LoginRequestJSON struct {
	Email 		string `json:"email"`
	Password 	string `json:"password"`
}

func (u *UserJSON) ToUser() *User {
	result := &User{
		ID: u.ID,
		Email: u.Email,
		Password: u.Password,
		UserName: u.Username,
		FirstName: u.FirstName,
		LastName: u.LastName,
		Address: u.Address,
		Gender: u.Gender,
	}

	if u.Mobile != "" {
		result.Mobile = sql.NullString{Valid: true, String: u.Mobile}
	}
	return result
}

func NewJSONUser(u *User) *UserJSON {
	return &UserJSON{
		ID: u.ID,
		Email: u.Email,
		FirstName: u.FirstName,
		LastName: u.LastName,
		Address: u.Address,
		Mobile: u.Mobile.String,
		Gender: u.Gender,
		Username: u.UserName,
	}
}