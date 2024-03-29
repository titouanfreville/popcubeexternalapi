package models

import (
	"encoding/json"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	u "github.com/titouanfreville/popcubeexternalapi/utils"
)

const (
	organisationDisplayNameMaxRunes = 64
	organisationNameMaxLength       = 64
	organisationDescriptionMaxRunes = 1024
	organisationSubjectMaxRunes     = 250
)

var (
	// EmptyOrganisation empty org var
	EmptyOrganisation = Organisation{}
)

// Organisation object
//
// Describe organisation you are in. It is an unique object in the database.
//
// swagger:model
type Organisation struct {
	// id of the organisation
	//
	// min: 0
	IDOrganisation uint64 `gorm:"primary_key;column:idOrganisation;AUTO_INCREMENT" json:"id,omitempty"`
	// Stack into docker swarm
	//
	// required: true
	//min: 0
	DockerStack int `gorm:"column:dockerStack;not null;unique" json:"docker_stack,omitempty"`
	// required: true
	OrganisationName string `gorm:"column:organisationName;not null;unique" json:"name,omitempty"`
	// State if organisation is free to join or not. Default is private (false).
	Public      bool   `gorm:"column:public; not null" json:"public"`
	Description string `gorm:"column:description" json:"description,omitempty"`
	Avatar      string `gorm:"column:avatar" json:"avatar,omitempty"`
	// Domain name of the organisation
	Domain string `gorm:"column:domain" json:"domain,omitempty"`
}

// Bind method used in API
func (organisation *Organisation) Bind(r *http.Request) error {
	return nil
}

// ToJSON transfoorm an Organisation into JSON
func (organisation *Organisation) ToJSON() string {
	b, err := json.Marshal(organisation)
	if err != nil {
		return ""
	}
	return string(b)
}

// OganisationFromJSON Try to parse a json object as emoji
func OganisationFromJSON(data io.Reader) *Organisation {
	decoder := json.NewDecoder(data)
	var organisation Organisation
	err := decoder.Decode(&organisation)
	if err == nil {
		return &organisation
	}
	return nil
}

// IsValid is used to check validity of Organisation objects
func (organisation *Organisation) IsValid() *u.AppError {

	if len(organisation.OrganisationName) == 0 || utf8.RuneCountInString(organisation.OrganisationName) > organisationDisplayNameMaxRunes {
		return u.NewLocAppError("Organisation.IsValid", "model.organisation.is_valid.organisation_name.app_error", nil, "id="+strconv.FormatUint(organisation.IDOrganisation, 10))
	}

	if !IsValidOrganisationIdentifier(organisation.OrganisationName) {
		return u.NewLocAppError("Organisation.IsValid", "model.organisation.is_valid.not_alphanum_organisation_name.app_error", nil, "id="+strconv.FormatUint(organisation.IDOrganisation, 10))
	}

	if utf8.RuneCountInString(organisation.Description) > organisationDescriptionMaxRunes {
		return u.NewLocAppError("Organisation.IsValid", "model.organisation.is_valid.description.app_error", nil, "id="+strconv.FormatUint(organisation.IDOrganisation, 10))
	}

	return nil
}

// PreSave is used to add some default values to organisation before saving in DB (creation).
func (organisation *Organisation) PreSave() {
	organisation.OrganisationName = strings.ToLower(organisation.OrganisationName)

	if organisation.Avatar == "" {
		organisation.Avatar = "default_organisation_avatar.svg"
	}
}

//IsValidOrganisationIdentifier check if string provided is a correct organisation identifier
func IsValidOrganisationIdentifier(s string) bool {

	return IsValidAlphaNum(s, true)
}

var validAlphaNumUnderscore = regexp.MustCompile(`^[a-z0-9]+([a-z\-\_0-9]+|(__)?)[a-z0-9]+$`)
var validAlphaNum = regexp.MustCompile(`^[a-z0-9]+([a-z\-0-9]+|(__)?)[a-z0-9]+$`)

//IsValidAlphaNum Check that string is correct lower case alpha numeric chain
func IsValidAlphaNum(s string, allowUnderscores bool) bool {
	var match bool
	if allowUnderscores {
		match = validAlphaNumUnderscore.MatchString(s)
	} else {
		match = validAlphaNum.MatchString(s)
	}

	return match
}
