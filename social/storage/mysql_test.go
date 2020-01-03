package storage

import (
	"github.com/chocosin/otus-hl/social/model"
	"github.com/satori/go.uuid"
	"reflect"
	"testing"
)

var testStorage *MysqlStorage

func init() {
	testConfig := &MysqlConfig{
		host:     "localhost",
		username: "root",
		password: "pass",
		dbName:   "test",
	}

	CreateDatabase(testConfig, true)
	Migrate(testConfig)

	var err error
	testStorage, err = NewMysqlStorage(testConfig)
	if err != nil {
		panic(err)
	}
}

func randomUser() *model.User {
	id := uuid.NewV1()
	idStr := id.String()
	passHash := model.HashPassword("password-" + idStr)
	return &model.User{
		ID:           id,
		Username:     "username-" + idStr,
		PasswordHash: passHash,
		FirstName:    "firstname-" + idStr,
		LastName:     "lastname-" + idStr,
		Age:          33,
		Interests:    []string{"cars", "cards", "news"},
		Gender:       "male",
		City:         "city" + idStr,
	}
}

func TestAddAndGet(t *testing.T) {
	u := randomUser()
	err := testStorage.InsertUser(u)
	if err != nil {
		t.Fatalf("error inserting user: %v", err)
	}
	dbUser, err := testStorage.FindUserByUsername(u.Username)
	if err != nil {
		t.Fatalf("error finding by username: %v", err)
	}
	if !reflect.DeepEqual(u, dbUser) {
		t.Fatalf("wrong user returned, \nexpected:\t%+v\nactual:\t\t%+v\n", u, dbUser)
	}
}

func TestReturnsNilWhenNotExists(t *testing.T) {
	u, err := testStorage.FindUserByUsername(uuid.NewV4().String())
	if err != nil {
		t.Fatalf("error finding by username %v", err)
	}
	if u != nil {
		t.Fatalf("expected to return nil, actual: %+v", u)
	}
}

func TestGetUserByTokenAndThenDelete(t *testing.T) {
	u := randomUser()
	err := testStorage.InsertUser(u)
	if err != nil {
		t.Fatalf("error inserting user: %v", err)
	}

	token := uuid.NewV1()
	usr, err := testStorage.GetUserByToken(token)
	if err != nil {
		t.Fatalf("error getting user by token: %v", err)
	}
	if usr != nil {
		t.Fatalf("shouldn't have found user")
	}

	err = testStorage.InsertToken(token, u.ID)
	if err != nil {
		t.Fatalf("error inserting token: %v", err)
	}

	dbUser, err := testStorage.GetUserByToken(token)
	if err != nil {
		t.Fatalf("error getting user by token: %v", err)
	}
	if !reflect.DeepEqual(dbUser, u) {
		t.Fatalf("wrong user returned, \nexpected:\t%+v\nactual:\t\t%+v\n", u, dbUser)
	}

	err = testStorage.DeleteToken(token)
	if err != nil {
		t.Fatalf("error deleting token: %v", err)
	}
	usr, err = testStorage.GetUserByToken(token)
	if err != nil {
		t.Fatalf("error getting user by token: %v", err)
	}
	if usr != nil {
		t.Fatalf("shouldn't have found user by deleted token")
	}
}

func TestLastUsernames(t *testing.T) {
	for idx := 0; idx < 5; idx++ {
		u := randomUser()
		err := testStorage.InsertUser(u)
		if err != nil {
			t.Fatalf("error inserting user: %v", err)
		}
	}
	expected := make([]string, 10)
	for idx := 0; idx < 10; idx++ {
		u := randomUser()
		err := testStorage.InsertUser(u)
		if err != nil {
			t.Fatalf("error inserting user: %v", err)
		}
		expected[10-idx-1] = u.Username
	}

	last, err := testStorage.LastUsernames()
	if err != nil {
		t.Fatalf("failed getting last usernames")
	}
	if !reflect.DeepEqual(last, expected) {
		t.Fatalf("wrong usernames returned, \nexpected:\t%+v\nactual:\t\t%+v\n", expected, last)
	}
}
