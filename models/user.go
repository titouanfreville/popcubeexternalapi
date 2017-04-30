package models

import (
	"encoding/json"
	"io"
	"net/http"
	"net/mail"
	"regexp"
	"strings"
	"unicode/utf8"

	u "github.com/titouanfreville/popcubeexternalapi/utils"
)

const (
	userNotifyAll           = "all"
	userNotifyMention       = "mention"
	userNotifyNone          = "none"
	userAuthServiceEmail    = "email"
	userAuthServiceUsername = "username"
)

var (
	userChannel = []string{"general", "random"}
	// Protected user name cause they are taken by system or used for special mentions.
	restrictedUsernames = []string{
		"all",
		"channel",
		"popcubebot",
		"here",
	}
	// EmptyUser em user
	EmptyUser = User{}
	// Definition of character user can possess in there names.
	validUsernameChars = regexp.MustCompile(`^[a-z0-9\.\-_]+$`)
)

// User object.
//
// An user is an account who have an access to a specific organisation. Each user is unique inside a given organisation, but users are not shared between
// organisations. Required apply only for creation of the object.
//
// swagger:model
type User struct {
	// id of the user
	//
	// min: 0
	IDUser uint64 `gorm:"primary_key;column:idUser;AUTO_INCREMENT" json:"id,omitempty"`
	// User name
	//
	// required: true
	// max length: 64
	Username string `gorm:"column:userName; not null; unique;" json:"username,omitempty"`
	// User email
	//
	// required: true
	// max lenght: 128
	Email string `gorm:"column:email; not null; unique;" json:"email,omitempty"`
	// State if email was verified
	//
	// required: true
	EmailVerified bool `gorm:"column:emailVerified; not null;" json:"email_verified,omitempty"`
	// User is deleted from organisation but still in database
	//
	// required: true
	Deleted bool `gorm:"column:deleted; not null;" json:"deleted,omitempty"`
	// Avatar used by user
	Avatar string `gorm:"column:avatar;" json:"avatar, omitempty"`
	// User nickname
	NickName string `gorm:"column:nickName; unique" json:"nickname, omitempty"`
	// First name
	FirstName string `gorm:"column:firstName;" json:"first_name, omitempty"`
	// User Lastname
	LastName     string         `gorm:"column:lastName;" json:"last_name, omitempty"`
	Organisation []Organisation `gorm:"ForeignKey:IDOrganisation;AssociationForeignKey:Refer" db:"-" json:"-"`
	// Org key of user in the organisation
	//
	// required: true
	IDOrganisation uint64 `gorm:"column:idOrganisation; not null;" json:"id_organisation,omitempty"`
}

// Bind method used in API to manage request.
func (user *User) Bind(r *http.Request) error {
	return nil
}

// IsValid valwebIDates the user and returns an error if it isn't configured
// correctly.
func (user *User) IsValid(isUpdate bool) *u.AppError {
	if !isUpdate {

		if len(user.Email) == 0 {
			return u.NewLocAppError("user.IsValid", "model.user.is_valid.Email.app_error", nil, "")
		}
	}

	if !IsValidUsername(user.Username) {
		return u.NewLocAppError("user.IsValid", "model.user.is_valid.Username.app_error", nil, "")
	}

	if len(user.Email) > 128 || !IsValidEmail(user.Email) {
		return u.NewLocAppError("user.IsValid", "model.user.is_valid.Email.app_error", nil, "")
	}

	if utf8.RuneCountInString(user.NickName) > 64 {
		return u.NewLocAppError("user.IsValid", "model.user.is_valid.NickName.app_error", nil, "")
	}

	if utf8.RuneCountInString(user.FirstName) > 64 {
		return u.NewLocAppError("user.IsValid", "model.user.is_valid.first_name.app_error", nil, "")
	}

	if utf8.RuneCountInString(user.LastName) > 64 {
		return u.NewLocAppError("user.IsValid", "model.user.is_valid.last_name.app_error", nil, "")
	}

	return nil
}

// PreSave have to be run before saving user in DB. It will fill necessary information (webID, username, etc. ) and hash password
func (user *User) PreSave() {

	user.Username = strings.ToLower(user.Username)
	user.Email = strings.ToLower(user.Email)
}

// ToJSON convert a user to a json string
func (user *User) ToJSON() string {
	b, err := json.Marshal(user)
	if err != nil {
		return ""
	}
	return string(b)
}

// UserFromJSON will decode the input and return a user
func UserFromJSON(data io.Reader) *User {
	decoder := json.NewDecoder(data)
	var user User
	err := decoder.Decode(&user)
	if err == nil {
		return &user
	}
	return nil
}

// IsValidUsername will check if provided userName is correct
func IsValidUsername(user string) bool {
	if len(user) == 0 || len(user) > 64 {
		return false
	}

	if !validUsernameChars.MatchString(user) {
		return false
	}

	for _, restrictedUsername := range restrictedUsernames {
		if user == restrictedUsername {
			return false
		}
	}

	return true
}

// IsValidEmail check email validity
func IsValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil && u.IsLower(email)
}
