package model

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"github.com/chocosin/otus-hl/social/templates"
	uuid "github.com/satori/go.uuid"
	"regexp"
	"strconv"
	"strings"
)

type GenderType = string

const (
	Male   GenderType = "male"
	Female GenderType = "female"
	Other  GenderType = "other"
)

type User struct {
	ID           uuid.UUID
	Username     string
	PasswordHash string
	FirstName    string
	LastName     string
	Age          int
	Interests    []string
	City         string
	Gender       GenderType
}

func (u *User) JoinInterests() string {
	return strings.Join(u.Interests, ", ")
}
func (u *User) SetInterests(joined string) {
	u.Interests = strings.FieldsFunc(joined, func(r rune) bool {
		return r == ','
	})
	for idx := range u.Interests {
		u.Interests[idx] = strings.TrimSpace(u.Interests[idx])
	}
}

var usernameRegexp = regexp.MustCompile("^[a-zA-Z]\\w+$")

func NewUserFromSignup(response *templates.SignupInfo) (*User, error) {
	username := strings.TrimSpace(response.Username)
	if len(username) < 3 {
		return nil, errors.New("username contains less than 3 chars")
	}
	if !usernameRegexp.MatchString(username) {
		return nil, errors.New("invalid username")
	}
	password := response.Password
	if len(password) < 3 {
		return nil, errors.New("password contains less than 3 chars")
	}
	passHash := HashPassword(password)

	lastName := strings.TrimSpace(response.LastName)
	if len(lastName) < 2 {
		return nil, errors.New("last name is less than 2 chars")
	}
	firstName := strings.TrimSpace(response.FirstName)
	if len(firstName) < 2 {
		return nil, errors.New("first name is less than 2 chars")
	}
	age, err := strconv.Atoi(response.Age)
	if err != nil {
		return nil, errors.New("couldn't parse age " + response.Age)
	}
	gender, err := getGender(response.Gender)
	if err != nil {
		return nil, err
	}
	city := strings.TrimSpace(response.City)
	if len(city) < 2 {
		return nil, errors.New("city is less than 2 chars")
	}
	interests := strings.FieldsFunc(response.Interests, func(r rune) bool {
		return r == ','
	})

	user := User{
		ID:           uuid.NewV1(),
		Username:     username,
		PasswordHash: passHash,
		FirstName:    firstName,
		LastName:     lastName,
		Age:          age,
		Gender:       gender,
		Interests:    interests,
		City:         city,
	}

	return &user, nil
}

func HashPassword(pass string) string {
	hash := md5.New()
	hash.Write([]byte(pass))
	return hex.EncodeToString(hash.Sum(nil))
}

func getGender(str string) (GenderType, error) {
	if str == Male || str == Female || str == Other {
		return str, nil
	}
	return "", errors.New("unknown gender " + str)
}

func (u *User) ToUserInfo(me bool) *templates.UserInfo {
	return &templates.UserInfo{
		Username:  u.Username,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Age:       u.Age,
		Interests: u.Interests,
		Gender:    u.Gender,
		City:      u.City,
		IsMe:      me,
	}
}
