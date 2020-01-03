package templates

import (
	"html/template"
	"net/url"
	"path"
)

type SignupInfo struct {
	Err       string
	Username  string
	Password  string
	FirstName string
	LastName  string
	Age       string
	Gender    string
	Interests string
	City      string
}

func NewSignupInfo(m url.Values) *SignupInfo {
	return &SignupInfo{
		Username:  m.Get("Username"),
		Password:  m.Get("Password"),
		FirstName: m.Get("FirstName"),
		LastName:  m.Get("LastName"),
		Age:       m.Get("Age"),
		Gender:    m.Get("Gender"),
		Interests: m.Get("Interests"),
		City:      m.Get("City"),
	}
}

type Hint struct {
	HintText string
	IsError  bool
}

type LoginInfo struct {
	Username string
	Password string
	Hint
}

func NewLoginInfo(m url.Values) *LoginInfo {
	return &LoginInfo{
		Username: m.Get("Username"),
		Password: m.Get("Password"),
	}
}

func (li *LoginInfo) ToResponse(hint Hint) {
	li.Password = ""
	li.Hint = hint
}

type UserInfo struct {
	Username  string
	FirstName string
	LastName  string
	Age       int
	Interests []string
	Gender    string
	City      string

	IsMe bool
}

type Templates struct {
	dir           string
	Signup        *template.Template
	Login         *template.Template
	User          *template.Template
	Index         *template.Template
	LastUsernames *template.Template
}

func NewTemplates(dir string) (*Templates, error) {
	templates := Templates{
		dir: dir,
	}
	var err error
	templates.Signup, err = template.ParseFiles(path.Join(dir, "signup.html"))
	if err != nil {
		return nil, err
	}
	templates.Login, err = template.ParseFiles(path.Join(dir, "login.html"))
	if err != nil {
		return nil, err
	}
	templates.User, err = template.ParseFiles(path.Join(dir, "user.html"))
	if err != nil {
		return nil, err
	}
	templates.Index, err = template.ParseFiles(path.Join(dir, "index.html"))
	if err != nil {
		return nil, err
	}
	templates.LastUsernames, err = template.ParseFiles(path.Join(dir, "lastUsernames.html"))
	if err != nil {
		return nil, err
	}
	return &templates, nil
}
