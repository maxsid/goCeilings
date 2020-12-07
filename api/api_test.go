package api

import (
	"bytes"
	"errors"
	"github.com/gorilla/mux"
	"github.com/maxsid/goCeilings/drawing/raster"
	"github.com/maxsid/goCeilings/figure"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"
)

// Storage testing realization
var errTestDB = errors.New("test error")
var tokens = make(map[uint]string)

type ErrorSimulation struct {
	Error              error
	RequestsUntilError int
}

type MockStorageT struct {
	ErrorSimulate          ErrorSimulation
	autoincrementUserID    uint
	autoincrementDrawingID uint
	users                  []*User
	drawings               []*Drawing
	relations              [][2]uint // 0 - userID, 1 - drawingID
}

func newMockStorage() *MockStorageT {
	storage := MockStorageT{autoincrementUserID: 4, autoincrementDrawingID: 10}
	storage.users = []*User{
		{UserOpen{1, "maxim", AdminPermission}, "12345"},
		{UserOpen{2, "oleg", UserPermission}, "123456"},
		{UserOpen{3, "elena", UserPermission}, "1234567"},
	}

	for _, u := range storage.users {
		t, err := createUserJWTToken(u.UserOpen, SigningSecret, time.Hour)
		if err != nil {
			panic(err)
		}
		tokens[u.ID] = t
	}

	drawing1 := raster.NewEmptyGGDrawing()
	_ = drawing1.Polygon.AddPoints([]*figure.Point{{X: 0, Y: 0}, {X: 0, Y: 1.25}, {X: 0.27, Y: 1.25}, {X: 0.2701, Y: 1.71},
		{X: 2.2201, Y: 1.6998}, {X: 2.25, Y: 0}}...)
	drawing2 := raster.NewEmptyGGDrawing()
	_ = drawing2.Polygon.AddPoints([]*figure.Point{{X: 0, Y: 0}, {X: 0, Y: 1.55}, {X: 0.725, Y: 1.55}, {X: 0.725, Y: 1.675},
		{X: 0.125, Y: 1.6751}, {X: 0.1253, Y: 5.9751}, {X: 3.4252, Y: 5.9999}, {X: 3.45, Y: 0}}...)
	storage.drawings = []*Drawing{
		{DrawingOpen{ID: 1, Name: "Drawing 1"}, *drawing1},
		{DrawingOpen{ID: 2, Name: "Drawing 2"}, *drawing2},
		{DrawingOpen{ID: 3, Name: "Drawing 3"}, *drawing1},
		{DrawingOpen{ID: 4, Name: "Drawing 4"}, *raster.NewGGDrawing()},
		{DrawingOpen{ID: 5, Name: "Drawing 5"}, *drawing1},
		{DrawingOpen{ID: 6, Name: "Drawing 6"}, *raster.NewGGDrawing()},
		{DrawingOpen{ID: 7, Name: "Drawing 7"}, *drawing2},
		{DrawingOpen{ID: 8, Name: "Drawing 8"}, *raster.NewGGDrawing()},
		{DrawingOpen{ID: 9, Name: "Drawing 9"}, *drawing2},
	}
	storage.relations = [][2]uint{
		{1, 2}, {1, 6}, {1, 8}, {1, 9},
		{2, 1},
		{3, 3}, {3, 4}, {3, 5}, {3, 7}}
	return &storage
}

func (td *MockStorageT) simulateError() (err error) {
	if td.ErrorSimulate.Error == nil {
		return nil
	}
	if td.ErrorSimulate.RequestsUntilError > 0 {
		td.ErrorSimulate.RequestsUntilError--
		return nil
	}
	err, td.ErrorSimulate.Error = td.ErrorSimulate.Error, nil
	return
}

func (td *MockStorageT) findUserByLogin(login string) *User {
	for _, u := range td.users {
		if u.Login == login {
			return u
		}
	}
	return nil
}

func (td *MockStorageT) getRelationsOfUser(userID uint) [][2]uint {
	out := make([][2]uint, 0)
	for _, da := range td.relations {
		if da[0] == userID {
			out = append(out, da)
		}
	}
	return out
}

func (td *MockStorageT) CreateUsers(users ...*User) error {
	if err := td.simulateError(); err != nil {
		return err
	}
	for _, u := range users {
		if td.findUserByLogin(u.Login) != nil {
			return ErrUserAlreadyExist
		}
		u.ID = td.autoincrementUserID
		td.autoincrementDrawingID++
	}
	td.users = append(td.users, users...)
	return nil
}

func (td *MockStorageT) GetUser(login, pass string) (*User, error) {
	if err := td.simulateError(); err != nil {
		return nil, err
	}
	for _, u := range td.users {
		if u.Login == login && u.Password == pass {
			return u, nil
		}
	}
	return nil, ErrUserNotFound
}

func (td *MockStorageT) GetUserByID(id uint) (*User, error) {
	if err := td.simulateError(); err != nil {
		return nil, err
	}
	for _, u := range td.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, ErrUserNotFound
}

func (td *MockStorageT) GetUsersList(page, pageLimit uint) ([]*UserOpen, error) {
	if err := td.simulateError(); err != nil {
		return nil, err
	}
	out := make([]*UserOpen, 0)
	amount, err := td.UsersAmount()
	if err != nil {
		return nil, err
	}
	for i, ui := uint(0), pageLimit*(page-1); ui < amount && i < pageLimit; ui, i = ui+1, i+1 {
		out = append(out, &td.users[ui].UserOpen)
	}
	return out, nil
}

func (td *MockStorageT) UsersAmount() (uint, error) {
	if err := td.simulateError(); err != nil {
		return 0, err
	}
	return uint(len(td.users)), nil
}

func (td *MockStorageT) RemoveUser(id uint) error {
	if err := td.simulateError(); err != nil {
		return err
	}
	pos := -1
	for i, u := range td.users {
		if u.ID == id {
			pos = i
			break
		}
	}
	if pos == -1 {
		return ErrUserNotFound
	}
	td.users = append(td.users[:pos], td.users[pos+1:]...)
	return nil
}

func (td *MockStorageT) UpdateUser(user *User) error {
	if err := td.simulateError(); err != nil {
		return err
	}
	for i := 0; i < len(td.users); i++ {
		dbUser := td.users[i]
		if dbUser.ID == user.ID {
			td.users[i].Password = user.Password
			td.users[i].Login = user.Login
			return nil
		}
	}
	return ErrUserNotFound
}

func (td *MockStorageT) GetDrawing(id uint) (*Drawing, error) {
	if err := td.simulateError(); err != nil {
		return nil, err
	}
	for _, d := range td.drawings {
		if d.ID == id {
			return d, nil
		}
	}
	return nil, ErrDrawingNotFound
}

func (td *MockStorageT) GetDrawingOfUser(userID, drawingID uint) (*Drawing, error) {
	if err := td.simulateError(); err != nil {
		return nil, err
	}
	for _, a := range td.relations {
		if a[0] == userID && a[1] == drawingID {
			return td.GetDrawing(drawingID)
		}
	}
	return nil, ErrDrawingNotFound
}

func (td *MockStorageT) GetDrawingsList(userID, page, pageLimit uint) ([]*DrawingOpen, error) {
	if err := td.simulateError(); err != nil {
		return nil, err
	}
	out := make([]*DrawingOpen, 0)

	offset := pageLimit * (page - 1)
	end := offset + pageLimit
	das := td.getRelationsOfUser(userID)
	if l := uint(len(das)); end > l {
		end = l
	}
	for _, da := range das[offset:end] {
		drawing, err := td.GetDrawing(da[1])
		if err != nil {
			return nil, err
		}
		out = append(out, &drawing.DrawingOpen)
	}
	return out, nil
}

func (td *MockStorageT) DrawingsAmount(userID uint) (uint, error) {
	if err := td.simulateError(); err != nil {
		return 0, err
	}
	l := len(td.getRelationsOfUser(userID))
	return uint(l), nil
}

func (td *MockStorageT) CreateDrawings(userID uint, drawings ...*Drawing) error {
	if err := td.simulateError(); err != nil {
		return err
	}
	for _, d := range drawings {
		d.ID = td.autoincrementDrawingID
		td.relations = append(td.relations, [2]uint{userID, td.autoincrementDrawingID})
		td.autoincrementDrawingID++
	}
	td.drawings = append(td.drawings, drawings...)
	return nil
}

func (td *MockStorageT) UpdateDrawing(drawing *Drawing) error {
	if err := td.simulateError(); err != nil {
		return err
	}
	for i := 0; i < len(td.drawings); i++ {
		if td.drawings[i].ID == drawing.ID {
			td.drawings[i] = drawing
			return nil
		}
	}
	return ErrDrawingNotFound
}

func (td *MockStorageT) UpdateDrawingOfUser(userID uint, drawing *Drawing) error {
	if err := td.simulateError(); err != nil {
		return err
	}
	_, err := td.GetDrawingOfUser(userID, drawing.ID)
	if err != nil {
		return err
	}
	return td.UpdateDrawing(drawing)
}

func (td *MockStorageT) RemoveDrawing(id uint) error {
	if err := td.simulateError(); err != nil {
		return err
	}
	for i := 0; i < len(td.drawings); i++ {
		if td.drawings[i].ID == id {
			td.drawings = append(td.drawings[:i], td.drawings[i+1:]...)
			return nil
		}
	}
	return ErrDrawingNotFound
}

func (td *MockStorageT) RemoveDrawingOfUser(userID, drawingID uint) error {
	if err := td.simulateError(); err != nil {
		return err
	}
	_, err := td.GetDrawingOfUser(userID, drawingID)
	if err != nil {
		return err
	}
	return td.RemoveDrawing(drawingID)
}

// Test cases
type TestCase struct {
	name                      string
	url                       string
	method                    string
	requestBody               string
	tokenUserID               uint
	inPanic                   bool
	wantStatus                int
	wantResponseBodyByPattern string
	wantResponseBodyEquality  string
	wantResponseHeaders       map[string]string
	testingHandler            http.Handler
	doWithRequest             func(r *http.Request)
	simulateDBError           ErrorSimulation
}

func checkTestCase(t *testing.T, tt TestCase, data *MockStorageT) {
	data.ErrorSimulate = tt.simulateDBError
	recorder := httptest.NewRecorder()

	var router *mux.Router
	if tt.testingHandler == nil {
		router = mux.NewRouter()
		addMiddlewaresToRouter(router, data)
		addHandlersToRouter(router, data)
	}

	req, err := http.NewRequest(tt.method, tt.url, bytes.NewBufferString(tt.requestBody))
	if tt.doWithRequest != nil {
		tt.doWithRequest(req)
	}
	if err != nil {
		t.Error(err)
	}
	if tt.tokenUserID != 0 {
		req.Header.Add("Authorization", "Bearer "+tokens[tt.tokenUserID])
	}
	// checks in a defer function
	defer func(recorder *httptest.ResponseRecorder, t *testing.T) {
		if err := recover(); err != nil && tt.inPanic {
			t.Log("Caught panic!")
		} else if err != nil && !tt.inPanic {
			t.Error(err)
		}
		if recorder.Code != tt.wantStatus {
			t.Errorf("Got status code = %v, want %v", recorder.Code, tt.wantStatus)
		}
		if tt.wantResponseBodyEquality != "" {
			if respBody := recorder.Body.String(); tt.wantResponseBodyEquality != respBody {
				t.Errorf("Got body = %v, want = %v", respBody, tt.wantResponseBodyEquality)
			}
		}

		if tt.wantResponseBodyByPattern != "" {
			respBody := recorder.Body.String()
			if ok, err := regexp.MatchString(tt.wantResponseBodyByPattern, respBody); err != nil {
				t.Error(err)
			} else if !ok {
				t.Errorf("Got body = %v, but pattern = %v", respBody, tt.wantResponseBodyByPattern)
			}
		}
		if tt.wantResponseHeaders != nil {
			for h, v := range tt.wantResponseHeaders {
				if respV := recorder.Header().Get(h); respV != v {
					t.Errorf("Got header %s = %s, want %s", h, respV, v)
				}
			}
		}
	}(recorder, t)
	if router != nil {
		router.ServeHTTP(recorder, req)
	} else if tt.testingHandler != nil {
		tt.testingHandler.ServeHTTP(recorder, req)
	} else {
		panic("Wrong TestCase: Router == nil and testingHandler == nil")
	}

}

// ===========
// Middlewares
// ===========

func Test_authorizationMiddleware(t *testing.T) {
	var err error
	storage := newMockStorage()
	tokens[300] = tokens[storage.users[0].ID] + "a"                                             // bad
	tokens[301], err = createUserJWTToken(storage.users[0].UserOpen, SigningSecret, -time.Hour) // expired
	if err != nil {
		t.Error(err)
		return
	}

	tests := []TestCase{
		{
			name:        "OK",
			url:         "/users",
			method:      http.MethodGet,
			tokenUserID: 1,
			wantStatus:  http.StatusOK,
		},
		{
			name:        "Forbidden",
			url:         "/users",
			method:      http.MethodGet,
			tokenUserID: 3,
			wantStatus:  http.StatusForbidden,
		},
		{
			name:        "Bad token",
			url:         "/users",
			method:      http.MethodGet,
			tokenUserID: 300,
			wantStatus:  http.StatusUnauthorized,
		},
		{
			name:        "Expired token",
			url:         "/users",
			method:      http.MethodGet,
			tokenUserID: 301,
			wantStatus:  http.StatusUnauthorized,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkTestCase(t, tt, storage)
		})
	}
}

func Test_gettingDrawingMiddlewareErrors(t *testing.T) {
	storage := newMockStorage()
	mf := gettingDrawingMiddleware(storage)
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	tests := []TestCase{
		{
			name:           "Skip",
			url:            "/",
			method:         http.MethodGet,
			testingHandler: mf(nextHandler),
			wantStatus:     http.StatusOK,
		},
		{
			name:           "Bad userID",
			url:            "/drawings/1",
			method:         http.MethodGet,
			testingHandler: mf(nextHandler),
			inPanic:        true,
			wantStatus:     http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkTestCase(t, tt, storage)
		})
	}
}

// ======
// /login
// ======

func Test_loginHandler(t *testing.T) {
	tests := []TestCase{
		{
			name:                      "OK",
			url:                       "/login",
			method:                    http.MethodPost,
			requestBody:               `{"login": "maxim", "password": "12345"}`,
			wantResponseBodyByPattern: `{"token":"[a-zA-Z0-9-_]+\.[a-zA-Z0-9-_]+\.[a-zA-Z0-9-_]+"}`,
			wantStatus:                http.StatusOK,
		},
		{
			name:        "Not found",
			url:         "/login",
			method:      http.MethodPost,
			requestBody: `{"login": "maxim", "password": "12345345"}`,
			wantStatus:  http.StatusNotFound,
		},
		{
			name:        "Bad request 1",
			url:         "/login",
			method:      http.MethodPost,
			requestBody: `{"login": "maxim", "apassword": "12345"}`,
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:        "Bad request 2",
			url:         "/login",
			method:      http.MethodPost,
			requestBody: `{"logrin": "maxim", "password": "12345"}`,
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:       "Bad request 3",
			url:        "/login",
			method:     http.MethodPost,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:        "Bad request 4",
			url:         "/login",
			method:      http.MethodPost,
			requestBody: `{"login": "", "password": "12345345"}`,
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:        "Bad request 5",
			url:         "/login",
			method:      http.MethodPost,
			requestBody: `{"login": "maxim", "password": ""}`,
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:            "DB Error",
			url:             "/login",
			method:          http.MethodPost,
			requestBody:     `{"login": "maxim", "password": "12345"}`,
			simulateDBError: ErrorSimulation{Error: errTestDB},
			inPanic:         true,
			wantStatus:      http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := newMockStorage()
			checkTestCase(t, tt, data)
		})
	}
}

// ======
// /users
// ======

func Test_getUsersListHandler(t *testing.T) {
	tests := []TestCase{
		{
			name:        "OK",
			url:         "/users",
			method:      http.MethodGet,
			tokenUserID: 1,
			wantStatus:  http.StatusOK,
			wantResponseBodyEquality: `{"users":[{"id":1,"login":"maxim","permission":1},` +
				`{"id":2,"login":"oleg","permission":2},{"id":3,"login":"elena","permission":2}],` +
				`"amount":3,"page":1,"page_limit":30,"pages":1}`,
		},
		{
			name:        "OK with params p=1&lim=2",
			url:         "/users?p=1&lim=2",
			method:      http.MethodGet,
			tokenUserID: 1,
			wantStatus:  http.StatusOK,
			wantResponseBodyEquality: `{"users":[{"id":1,"login":"maxim","permission":1},` +
				`{"id":2,"login":"oleg","permission":2}],"amount":3,"page":1,"page_limit":2,"pages":2}`,
		},
		{
			name:        "OK with params p=2&lim=2",
			url:         "/users?p=2&lim=2",
			method:      http.MethodGet,
			tokenUserID: 1,
			wantStatus:  http.StatusOK,
			wantResponseBodyEquality: `{"users":[{"id":3,"login":"elena","permission":2}],` +
				`"amount":3,"page":2,"page_limit":2,"pages":2}`,
		},
		{
			name:        "OK with params p=2 (page>pages -> page=pages)",
			url:         "/users?p=2",
			method:      http.MethodGet,
			tokenUserID: 1,
			wantStatus:  http.StatusOK,
			wantResponseBodyEquality: `{"users":[{"id":1,"login":"maxim","permission":1},` +
				`{"id":2,"login":"oleg","permission":2},{"id":3,"login":"elena","permission":2}],` +
				`"amount":3,"page":1,"page_limit":30,"pages":1}`,
		},
		{
			name:        "OK with params lim=2",
			url:         "/users?lim=1",
			method:      http.MethodGet,
			tokenUserID: 1,
			wantStatus:  http.StatusOK,
			wantResponseBodyEquality: `{"users":[{"id":1,"login":"maxim","permission":1}],` +
				`"amount":3,"page":1,"page_limit":1,"pages":3}`,
		},
		{
			name:       "Unauthorized",
			url:        "/users",
			method:     http.MethodGet,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:            "DB Amount error",
			url:             "/users?p=2&limit=2",
			method:          http.MethodGet,
			tokenUserID:     1,
			wantStatus:      http.StatusInternalServerError,
			inPanic:         true,
			simulateDBError: ErrorSimulation{Error: errTestDB},
		},
		{
			name:            "DB GetList error",
			url:             "/users?p=2&limit=2",
			method:          http.MethodGet,
			tokenUserID:     1,
			wantStatus:      http.StatusInternalServerError,
			inPanic:         true,
			simulateDBError: ErrorSimulation{Error: errTestDB, RequestsUntilError: 1},
		},
		{
			name:        "Params Error",
			url:         "/users?p=d&limit=c",
			method:      http.MethodGet,
			tokenUserID: 1,
			wantStatus:  http.StatusBadRequest,
			inPanic:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := newMockStorage()
			checkTestCase(t, tt, data)
		})
	}
}

func Test_getUserCreatingHandler(t *testing.T) {
	tests := []TestCase{
		{
			name:                "OK",
			url:                 "/users",
			method:              http.MethodPost,
			requestBody:         `{"login": "zhenya", "password": "321456"}`,
			tokenUserID:         1,
			wantStatus:          http.StatusCreated,
			wantResponseHeaders: map[string]string{"Location": "/users/4"},
		},
		{
			name:        "Bad Request 1",
			url:         "/users",
			method:      http.MethodPost,
			requestBody: `{"loin": "zhenya", "pasword": "321456"}`,
			tokenUserID: 1,
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:        "Bad Request 2",
			url:         "/users",
			method:      http.MethodPost,
			requestBody: `{"login": "", "password": "321456"}`,
			tokenUserID: 1,
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:        "Bad Request 3",
			url:         "/users",
			method:      http.MethodPost,
			requestBody: `{"login": "asket", "password": ""}`,
			tokenUserID: 1,
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:        "Bad Request 4",
			url:         "/users",
			method:      http.MethodPost,
			tokenUserID: 1,
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:        "Unauthorized",
			url:         "/users",
			requestBody: `{"login": "zhenya", "password": "321456"}`,
			method:      http.MethodPost,
			wantStatus:  http.StatusUnauthorized,
		},
		{
			name:            "DB Error",
			url:             "/users",
			method:          http.MethodPost,
			requestBody:     `{"login": "zhenya", "password": "321456"}`,
			tokenUserID:     1,
			wantStatus:      http.StatusInternalServerError,
			simulateDBError: ErrorSimulation{Error: errTestDB},
			inPanic:         true,
		},
		{
			name:        "User with this login already exists",
			url:         "/users",
			method:      http.MethodPost,
			requestBody: `{"login": "elena", "password": "321456"}`,
			tokenUserID: 1,
			wantStatus:  http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := newMockStorage()
			checkTestCase(t, tt, data)
		})
	}
}

// ===========
// /users/{id}
// ===========

func Test_getUserGettingHandler(t *testing.T) {
	tests := []TestCase{
		{
			name:                     "OK 1",
			url:                      "/users/1",
			method:                   http.MethodGet,
			tokenUserID:              1,
			wantStatus:               http.StatusOK,
			wantResponseBodyEquality: `{"id":1,"login":"maxim","permission":1}`,
		},
		{
			name:                     "OK 2",
			url:                      "/users/2",
			method:                   http.MethodGet,
			tokenUserID:              1,
			wantStatus:               http.StatusOK,
			wantResponseBodyEquality: `{"id":2,"login":"oleg","permission":2}`,
		},
		{
			name:        "Not Found",
			url:         "/users/25",
			method:      http.MethodGet,
			tokenUserID: 1,
			wantStatus:  http.StatusNotFound,
		},
		{
			name:       "Unauthorized",
			url:        "/users/1",
			method:     http.MethodGet,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:            "DB Error",
			url:             "/users/1",
			method:          http.MethodGet,
			tokenUserID:     1,
			wantStatus:      http.StatusInternalServerError,
			simulateDBError: ErrorSimulation{Error: errTestDB},
			inPanic:         true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := newMockStorage()
			checkTestCase(t, tt, data)
		})
	}
}

func Test_getUserRemovingHandler(t *testing.T) {
	type RemoveTestCase struct {
		TestCase
		RemovingUserID uint
	}
	tests := []RemoveTestCase{
		{TestCase: TestCase{
			name:        "OK 1",
			url:         "/users/1",
			method:      http.MethodDelete,
			tokenUserID: 1,
			wantStatus:  http.StatusOK,
		},
			RemovingUserID: 1,
		},
		{TestCase: TestCase{
			name:        "OK 2",
			url:         "/users/2",
			method:      http.MethodDelete,
			tokenUserID: 1,
			wantStatus:  http.StatusOK,
		},
			RemovingUserID: 2,
		},
		{TestCase: TestCase{
			name:        "Not Found",
			url:         "/users/25",
			method:      http.MethodDelete,
			tokenUserID: 1,
			wantStatus:  http.StatusNotFound,
		}},
		{TestCase: TestCase{
			name:       "Unauthorized",
			url:        "/users/1",
			method:     http.MethodDelete,
			wantStatus: http.StatusUnauthorized,
		}},
		{TestCase: TestCase{
			name:            "DB Error",
			url:             "/users/1",
			method:          http.MethodDelete,
			tokenUserID:     1,
			wantStatus:      http.StatusInternalServerError,
			simulateDBError: ErrorSimulation{Error: errTestDB},
			inPanic:         true,
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := newMockStorage()
			checkTestCase(t, tt.TestCase, data)
			if tt.RemovingUserID != 0 {
				_, err := data.GetUserByID(tt.RemovingUserID)
				if !errors.Is(err, ErrUserNotFound) {
					t.Errorf("After removing data.GetUserByID() got %v, want ErrUserNotFound", err)
				}
			}
		})
	}
}

func Test_getUserUpdatingHandler(t *testing.T) {
	type UpdatingTestCase struct {
		TestCase
		userID          uint
		login, password string
		permission      Permission
	}
	tests := []UpdatingTestCase{
		{TestCase: TestCase{
			name:        "OK 1",
			url:         "/users/1",
			method:      http.MethodPut,
			requestBody: `{"login":"petr", "password":"12345","is_admin":true}`,
			tokenUserID: 1,
			wantStatus:  http.StatusOK,
		},
			userID: 1, login: "petr",
		},
		{TestCase: TestCase{
			name:        "OK 2",
			url:         "/users/2",
			method:      http.MethodPut,
			requestBody: `{"login":"oleg", "password":"32145","is_admin":false}`,
			tokenUserID: 1,
			wantStatus:  http.StatusOK,
		},
			userID: 2, password: "32145",
		},
		{TestCase: TestCase{
			name:        "Not Found",
			url:         "/users/25",
			method:      http.MethodPut,
			requestBody: `{"login":"oleg", "password":"32145","is_admin":false}`,
			tokenUserID: 1,
			wantStatus:  http.StatusNotFound,
		}},
		{TestCase: TestCase{
			name:        "Bad Request",
			url:         "/users/2",
			method:      http.MethodPut,
			tokenUserID: 1,
			wantStatus:  http.StatusBadRequest,
		}},
		{TestCase: TestCase{
			name:        "Unauthorized",
			url:         "/users/2",
			method:      http.MethodPut,
			requestBody: `{"login":"oleg", "password":"32145","is_admin":false}`,
			wantStatus:  http.StatusUnauthorized,
		}},
		{TestCase: TestCase{
			name:            "DB Error",
			url:             "/users/1",
			method:          http.MethodPut,
			requestBody:     `{"login":"oleg", "password":"32145","is_admin":false}`,
			tokenUserID:     1,
			wantStatus:      http.StatusInternalServerError,
			simulateDBError: ErrorSimulation{Error: errTestDB},
			inPanic:         true,
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := newMockStorage()
			checkTestCase(t, tt.TestCase, data)
			if tt.userID != 0 {
				u, err := data.GetUserByID(tt.userID)
				if err != nil {
					t.Error(err)
				}
				if tt.login != "" && tt.login != u.Login {
					t.Errorf("getUserUpdatingHandler() not changed login of %d user. Got %s, want %s", tt.userID, u.Login, tt.login)
				}
				if tt.password != "" && tt.password != u.Password {
					t.Errorf("getUserUpdatingHandler() not changed password of %d user. Got %s, want %s", tt.userID, u.Password, tt.password)
				}
				if tt.permission != 0 && tt.permission != u.Permission {
					t.Errorf("getUserUpdatingHandler() not changed permission of %d user. Got %v, want %v",
						tt.userID, u.Permission, tt.permission)
				}
			}
		})
	}
}

// =========
// /drawings
// =========

func Test_getDrawingsListGettingHandler(t *testing.T) {
	tests := []TestCase{
		{
			name:        "OK 1",
			url:         "/drawings",
			method:      http.MethodGet,
			wantStatus:  http.StatusOK,
			tokenUserID: 1,
			wantResponseBodyEquality: `{"drawings":[{"id":2,"name":"Drawing 2"},{"id":6,"name":"Drawing 6"},` +
				`{"id":8,"name":"Drawing 8"},{"id":9,"name":"Drawing 9"}],"amount":4,"page":1,"page_limit":30,"pages":1}`,
		},
		{
			name:        "OK 2",
			url:         "/drawings",
			method:      http.MethodGet,
			wantStatus:  http.StatusOK,
			tokenUserID: 2,
			wantResponseBodyEquality: `{"drawings":[{"id":1,"name":"Drawing 1"}],` +
				`"amount":1,"page":1,"page_limit":30,"pages":1}`,
		},
		{
			name:        "OK 3 with params",
			url:         "/drawings?p=2&lim=2",
			method:      http.MethodGet,
			wantStatus:  http.StatusOK,
			tokenUserID: 3,
			wantResponseBodyEquality: `{"drawings":[{"id":5,"name":"Drawing 5"},{"id":7,"name":"Drawing 7"}],` +
				`"amount":4,"page":2,"page_limit":2,"pages":2}`,
		},
		{
			name:        "OK 1 with only p=2 (page>pages -> page=pages)",
			url:         "/drawings?p=2",
			method:      http.MethodGet,
			wantStatus:  http.StatusOK,
			tokenUserID: 1,
			wantResponseBodyEquality: `{"drawings":[{"id":2,"name":"Drawing 2"},{"id":6,"name":"Drawing 6"},` +
				`{"id":8,"name":"Drawing 8"},{"id":9,"name":"Drawing 9"}],"amount":4,"page":1,"page_limit":30,"pages":1}`,
		},
		{
			name:        "OK 1 with only lim=3",
			url:         "/drawings?lim=3",
			method:      http.MethodGet,
			wantStatus:  http.StatusOK,
			tokenUserID: 1,
			wantResponseBodyEquality: `{"drawings":[{"id":2,"name":"Drawing 2"},{"id":6,"name":"Drawing 6"},` +
				`{"id":8,"name":"Drawing 8"}],"amount":4,"page":1,"page_limit":3,"pages":2}`,
		},
		{
			name:       "Unauthorized",
			url:        "/drawings",
			method:     http.MethodGet,
			wantStatus: http.StatusUnauthorized,
			inPanic:    true,
		},
		{
			name:            "DB Amount Error",
			url:             "/drawings",
			method:          http.MethodGet,
			wantStatus:      http.StatusInternalServerError,
			tokenUserID:     1,
			simulateDBError: ErrorSimulation{Error: errTestDB},
			inPanic:         true,
		},
		{
			name:            "DB GetList Error",
			url:             "/drawings",
			method:          http.MethodGet,
			wantStatus:      http.StatusInternalServerError,
			tokenUserID:     1,
			simulateDBError: ErrorSimulation{Error: errTestDB, RequestsUntilError: 1},
			inPanic:         true,
		},
		{
			name:        "Bad params",
			url:         "/drawings?p=d&lim=a",
			method:      http.MethodGet,
			wantStatus:  http.StatusBadRequest,
			tokenUserID: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := newMockStorage()
			checkTestCase(t, tt, data)
		})
	}
}

func Test_getDrawingCreatingHandler(t *testing.T) {
	tests := []TestCase{
		{
			name:                "OK",
			url:                 "/drawings",
			method:              http.MethodPost,
			requestBody:         `{"name":"New Drawing"}`,
			wantStatus:          http.StatusCreated,
			tokenUserID:         1,
			wantResponseHeaders: map[string]string{"Location": "/drawings/10"},
		},
		{
			name:        "Bad request 1",
			url:         "/drawings",
			method:      http.MethodPost,
			requestBody: `{}`,
			wantStatus:  http.StatusBadRequest,
			tokenUserID: 1,
		},
		{
			name:        "Bad request 2",
			url:         "/drawings",
			method:      http.MethodPost,
			requestBody: `{"game":"New Drawing"}`,
			wantStatus:  http.StatusBadRequest,
			tokenUserID: 1,
		},
		{
			name:        "Bad request 3",
			url:         "/drawings",
			method:      http.MethodPost,
			wantStatus:  http.StatusBadRequest,
			tokenUserID: 1,
		},
		{
			name:       "Unauthorized",
			url:        "/drawings",
			method:     http.MethodPost,
			wantStatus: http.StatusUnauthorized,
			inPanic:    true,
		},
		{
			name:            "DB Error",
			url:             "/drawings",
			method:          http.MethodPost,
			requestBody:     `{"name":"New Drawing"}`,
			tokenUserID:     1,
			wantStatus:      http.StatusInternalServerError,
			simulateDBError: ErrorSimulation{Error: errTestDB},
			inPanic:         true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := newMockStorage()
			checkTestCase(t, tt, data)
		})
	}
}

// ==============
// /drawings/{id}
// =============

func Test_drawingGettingHandler(t *testing.T) {
	tests := []TestCase{
		{
			name:        "OK 1",
			url:         "/drawings/2",
			method:      http.MethodGet,
			wantStatus:  http.StatusOK,
			tokenUserID: 1,
			wantResponseBodyEquality: `{"id":2,"name":"Drawing 2","area":19.95,"perimeter":20.05,"points_count":8,` +
				`"width":345,"height":599.99,` +
				`"points":[{"x":0,"y":0},{"x":0,"y":155},{"x":72.5,"y":155},{"x":72.5,"y":167.5},` +
				`{"x":12.5,"y":167.51},{"x":12.53,"y":597.51},{"x":342.52,"y":599.99},{"x":345,"y":0}],` +
				`"measures":{"length":"cm","area":"m2","perimeter":"m","angle":"deg"}}`,
		},
		{
			name:        "OK 2",
			url:         "/drawings/1",
			method:      http.MethodGet,
			wantStatus:  http.StatusOK,
			tokenUserID: 2,
			wantResponseBodyEquality: `{"id":1,"name":"Drawing 1","area":3.69,"perimeter":7.88,"points_count":6,` +
				`"width":225,"height":171,` +
				`"points":[{"x":0,"y":0},{"x":0,"y":125},{"x":27,"y":125},{"x":27.01,"y":171},{"x":222.01,"y":169.98},` +
				`{"x":225,"y":0}],"measures":{"length":"cm","area":"m2","perimeter":"m","angle":"deg"}}`,
		},
		{
			name:        "Not found",
			url:         "/drawings/432",
			method:      http.MethodGet,
			wantStatus:  http.StatusNotFound,
			tokenUserID: 3,
		},
		{
			name:        "Don't have access",
			url:         "/drawings/1",
			method:      http.MethodGet,
			wantStatus:  http.StatusNotFound,
			tokenUserID: 1,
		},
		{
			name:       "Unauthorized",
			url:        "/drawings/1",
			method:     http.MethodGet,
			wantStatus: http.StatusUnauthorized,
			inPanic:    true,
		},
		{
			name:            "DB Error",
			url:             "/drawings/1",
			method:          http.MethodGet,
			tokenUserID:     1,
			wantStatus:      http.StatusInternalServerError,
			simulateDBError: ErrorSimulation{Error: errTestDB},
			inPanic:         true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := newMockStorage()
			checkTestCase(t, tt, data)
		})
	}
}

// ====================
// /drawings/{id}/image
// ====================

func Test_drawingImageHandler(t *testing.T) {
	type DrawingTestCase struct {
		TestCase
		DrawingID uint
	}
	tests := []DrawingTestCase{
		{TestCase: TestCase{
			name:                "OK",
			url:                 "/drawings/2/image",
			method:              http.MethodGet,
			wantStatus:          http.StatusOK,
			wantResponseHeaders: map[string]string{"Content-Type": "image/png"},
			tokenUserID:         1,
		},
			DrawingID: 2,
		},
		{TestCase: TestCase{
			name:        "User doesn't have access",
			url:         "/drawings/1/image",
			method:      http.MethodGet,
			wantStatus:  http.StatusNotFound,
			tokenUserID: 1,
		}},
		{TestCase: TestCase{
			name:        "Not found",
			url:         "/drawings/432/image",
			method:      http.MethodGet,
			wantStatus:  http.StatusNotFound,
			tokenUserID: 1,
		}},
		{TestCase: TestCase{
			name:       "Unauthorized",
			url:        "/drawings/1/image",
			method:     http.MethodGet,
			wantStatus: http.StatusUnauthorized,
			inPanic:    true,
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := newMockStorage()
			checkTestCase(t, tt.TestCase, data)
		})
	}
}

func Test_getDrawingDeletingHandler(t *testing.T) {
	type DrawingTestCase struct {
		TestCase
		DrawingID uint
	}
	tests := []DrawingTestCase{
		{TestCase: TestCase{
			name:        "OK",
			url:         "/drawings/6",
			method:      http.MethodDelete,
			wantStatus:  http.StatusOK,
			tokenUserID: 1,
		},
			DrawingID: 6,
		},
		{TestCase: TestCase{
			name:        "User doesn't have access",
			url:         "/drawings/1",
			method:      http.MethodDelete,
			wantStatus:  http.StatusNotFound,
			tokenUserID: 1,
		}},
		{TestCase: TestCase{
			name:        "Not found",
			url:         "/drawings/432",
			method:      http.MethodDelete,
			wantStatus:  http.StatusNotFound,
			tokenUserID: 1,
		}},
		{TestCase: TestCase{
			name:       "Unauthorized",
			url:        "/drawings/1",
			method:     http.MethodDelete,
			wantStatus: http.StatusUnauthorized,
			inPanic:    true,
		}},
		{TestCase: TestCase{
			name:            "DB Error",
			url:             "/drawings/2",
			method:          http.MethodDelete,
			tokenUserID:     1,
			wantStatus:      http.StatusInternalServerError,
			simulateDBError: ErrorSimulation{Error: errTestDB},
			inPanic:         true,
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := newMockStorage()
			checkTestCase(t, tt.TestCase, data)
			if tt.DrawingID != 0 {
				if _, err := data.GetDrawing(tt.DrawingID); !errors.Is(err, ErrDrawingNotFound) {
					t.Errorf("Drawing getting by ID have to return ErrDrawingNotFound error, but got %v", err)
				}
			}
		})
	}
}

// =====================
// /drawings/{id}/points
// =====================

func Test_drawingPointsGettingHandler(t *testing.T) {
	tests := []TestCase{
		{
			name:        "OK 1",
			url:         "/drawings/1/points",
			method:      http.MethodGet,
			wantStatus:  http.StatusOK,
			tokenUserID: 2,
			wantResponseBodyEquality: `{"id":1,"name":"Drawing 1","points":[{"x":0,"y":0},{"x":0,"y":125},{"x":27,"y":125},{"x":27.01,"y":171},` +
				`{"x":222.01,"y":169.98},{"x":225,"y":0}],"measure":"cm"}`,
		},
		{
			name:        "OK 2",
			url:         "/drawings/2/points",
			method:      http.MethodGet,
			wantStatus:  http.StatusOK,
			tokenUserID: 1,
			wantResponseBodyEquality: `{"id":2,"name":"Drawing 2","points":[{"x":0,"y":0},{"x":0,"y":155},{"x":72.5,"y":155},` +
				`{"x":72.5,"y":167.5},{"x":12.5,"y":167.51},{"x":12.53,"y":597.51},` +
				`{"x":342.52,"y":599.99},{"x":345,"y":0}],"measure":"cm"}`,
		},
		{
			name:        "OK with params m=m&p=4",
			url:         "/drawings/2/points?m=m&p=4",
			method:      http.MethodGet,
			wantStatus:  http.StatusOK,
			tokenUserID: 1,
			wantResponseBodyEquality: `{"id":2,"name":"Drawing 2","points":[{"x":0,"y":0},{"x":0,"y":1.55},{"x":0.725,"y":1.55},` +
				`{"x":0.725,"y":1.675},{"x":0.125,"y":1.6751},{"x":0.1253,"y":5.9751},{"x":3.4252,"y":5.9999},` +
				`{"x":3.45,"y":0}],"measure":"m"}`,
		},
		{
			name:        "Bad request param m=de",
			url:         "/drawings/2/points?m=de",
			method:      http.MethodGet,
			wantStatus:  http.StatusBadRequest,
			tokenUserID: 1,
		},
		{
			name:        "Bad request param m=deg",
			url:         "/drawings/2/points?m=deg",
			method:      http.MethodGet,
			wantStatus:  http.StatusBadRequest,
			tokenUserID: 1,
		},
		{
			name:        "Bad request param p=dsa",
			url:         "/drawings/2/points?p=dsa",
			method:      http.MethodGet,
			wantStatus:  http.StatusBadRequest,
			tokenUserID: 1,
		},
		{
			name:        "Bad request param p=-2",
			url:         "/drawings/2/points?p=-2",
			method:      http.MethodGet,
			wantStatus:  http.StatusBadRequest,
			tokenUserID: 1,
		},
		{
			name:        "User doesn't have access",
			url:         "/drawings/1/points",
			method:      http.MethodGet,
			wantStatus:  http.StatusNotFound,
			tokenUserID: 1,
		},
		{
			name:        "Not found",
			url:         "/drawings/432/points",
			method:      http.MethodGet,
			wantStatus:  http.StatusNotFound,
			tokenUserID: 1,
		},
		{
			name:       "Unauthorized",
			url:        "/drawings/1/points",
			method:     http.MethodGet,
			wantStatus: http.StatusUnauthorized,
			inPanic:    true,
		},
		{
			name:            "DB Error",
			url:             "/drawings/2/points",
			method:          http.MethodGet,
			tokenUserID:     1,
			wantStatus:      http.StatusInternalServerError,
			simulateDBError: ErrorSimulation{Error: errTestDB},
			inPanic:         true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := newMockStorage()
			checkTestCase(t, tt, data)
		})
	}
}

func Test_getDrawingPointsAddingHandler(t *testing.T) {
	tests := []TestCase{
		{
			name:   "OK only coords",
			url:    "/drawings/6/points",
			method: http.MethodPost,
			requestBody: `{"points":[{"x":0,"y":125},{"x":27,"y":125},{"x":27.01,"y":171},` +
				`{"x":222.01,"y":169.98},{"x":225,"y":0}],"measures":{"length":"cm"}}`,
			wantStatus:  http.StatusOK,
			tokenUserID: 1,
			wantResponseBodyEquality: `{"id":6,"name":"Drawing 6","points":[{"x":0,"y":0},{"x":0,"y":125},{"x":27,"y":125},{"x":27.01,"y":171},` +
				`{"x":222.01,"y":169.98},{"x":225,"y":0}],"measure":"cm"}`,
		},
		{
			name:   "OK mixed",
			url:    "/drawings/6/points",
			method: http.MethodPost,
			requestBody: `{"points":[{"distance":125,"direction":90},{"distance":27,"angle":90},{"x":27.01,"y":171},` +
				`{"x":222.01,"y":169.98},{"x":225,"y":0}],"measures":{"length":"cm","angle":"deg"}}`,
			wantStatus:  http.StatusOK,
			tokenUserID: 1,
			wantResponseBodyEquality: `{"id":6,"name":"Drawing 6","points":[{"x":0,"y":0},{"x":0,"y":125},{"x":27,"y":125},{"x":27.01,"y":171},` +
				`{"x":222.01,"y":169.98},{"x":225,"y":0}],"measure":"cm"}`,
		},
		{
			name:   "User doesn't have access",
			url:    "/drawings/1/points",
			method: http.MethodPost,
			requestBody: `{"points":[{"distance":125,"angle":90},{"distance":27,"angle":90},{"x":27.01,"y":171},` +
				`{"x":222.01,"y":169.98},{"x":225,"y":0}],"measures":{"length":"cm","angle":"deg"}}`,
			wantStatus:  http.StatusNotFound,
			tokenUserID: 1,
		},
		{
			name:   "Not found",
			url:    "/drawings/432/points",
			method: http.MethodPost,
			requestBody: `{"points":[{"distance":125,"angle":90},{"distance":27,"angle":90},{"x":27.01,"y":171},` +
				`{"x":222.01,"y":169.98},{"x":225,"y":0}],"measures":{"length":"cm","angle":"deg"}}`,
			wantStatus:  http.StatusNotFound,
			tokenUserID: 1,
		},
		{
			name:   "Unauthorized",
			url:    "/drawings/1/points",
			method: http.MethodPost,
			requestBody: `{"points":[{"distance":125,"angle":90},{"distance":27,"angle":90},{"x":27.01,"y":171},` +
				`{"x":222.01,"y":169.98},{"x":225,"y":0}],"measures":{"length":"cm","angle":"deg"}}`,
			wantStatus: http.StatusUnauthorized,
			inPanic:    true,
		},
		{
			name:   "DB Error",
			url:    "/drawings/2/points",
			method: http.MethodPost,
			requestBody: `{"points":[{"distance":125,"angle":90},{"distance":27,"angle":90},{"x":27.01,"y":171},` +
				`{"x":222.01,"y":169.98},{"x":225,"y":0}],"measures":{"length":"cm","angle":"deg"}}`,
			tokenUserID:     1,
			wantStatus:      http.StatusInternalServerError,
			simulateDBError: ErrorSimulation{Error: errTestDB, RequestsUntilError: 1},
			inPanic:         true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := newMockStorage()
			checkTestCase(t, tt, data)
		})
	}
}

// =========================
// /drawings/{id}/points/{n}
// =========================

func Test_drawingPointGettingHandler(t *testing.T) {
	tests := []TestCase{
		{
			name:                     "OK",
			url:                      "/drawings/1/points/1",
			method:                   http.MethodGet,
			wantStatus:               http.StatusOK,
			tokenUserID:              2,
			wantResponseBodyEquality: `{"x":0,"y":0,"measure":"cm"}`,
		},
		{
			name:                     "OK with params",
			url:                      "/drawings/1/points/2?m=km&p=4",
			method:                   http.MethodGet,
			wantStatus:               http.StatusOK,
			tokenUserID:              2,
			wantResponseBodyEquality: `{"x":0,"y":0.0013,"measure":"km"}`,
		},
		{
			name:        "Not found point number",
			url:         "/drawings/1/points/132",
			method:      http.MethodGet,
			wantStatus:  http.StatusNotFound,
			tokenUserID: 2,
		},
		{
			name:        "Bad m param",
			url:         "/drawings/1/points/2?m=ll",
			method:      http.MethodGet,
			wantStatus:  http.StatusBadRequest,
			tokenUserID: 2,
		},
		{
			name:        "Bad p param",
			url:         "/drawings/1/points/2?m=km&p=p",
			method:      http.MethodGet,
			wantStatus:  http.StatusBadRequest,
			tokenUserID: 2,
		},
	}
	storage := newMockStorage()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkTestCase(t, tt, storage)
		})
	}
}

func Test_getDrawingPointDeletingHandler(t *testing.T) {
	type DrawingTestCase struct {
		TestCase
		DrawingID uint
	}
	tests := []DrawingTestCase{
		{TestCase: TestCase{
			name:        "OK",
			url:         "/drawings/1/points/1",
			method:      http.MethodDelete,
			tokenUserID: 2,
			wantStatus:  http.StatusOK,
		}, DrawingID: 1},
		{TestCase: TestCase{
			name:        "Too big point number",
			url:         "/drawings/1/points/412",
			method:      http.MethodDelete,
			tokenUserID: 2,
			wantStatus:  http.StatusNotFound,
		}},
		{TestCase: TestCase{
			name:            "DB Error",
			url:             "/drawings/1/points/1",
			method:          http.MethodDelete,
			wantStatus:      http.StatusInternalServerError,
			tokenUserID:     2,
			simulateDBError: ErrorSimulation{Error: errTestDB, RequestsUntilError: 1},
			inPanic:         true,
		}},
	}
	storage := newMockStorage()
	original := newMockStorage()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkTestCase(t, tt.TestCase, storage)
			if tt.wantStatus == http.StatusOK {
				d, err := storage.GetDrawing(tt.DrawingID)
				if err != nil {
					t.Errorf("Fail of drawing deleting chacking: %v", err)
					return
				}
				od, err := original.GetDrawing(tt.DrawingID)
				if err != nil {
					t.Errorf("Fail of drawing deleting chacking: %v", err)
					return
				}
				if want, got := od.Len()-1, d.Len(); want != got {
					t.Errorf("Point hasn't deleted! Len got %d, want = %d", got, want)
				}
			}
		})
	}
}

func Test_getDrawingPointUpdatingHandler(t *testing.T) {
	type UpdatePointTestCase struct {
		TestCase
		DrawingID          uint
		PointsCoordsResult map[int][2]float64
	}
	tests := []UpdatePointTestCase{
		{TestCase: TestCase{
			name:        "OK Empty coordinates",
			url:         "/drawings/1/points/2",
			method:      http.MethodPut,
			wantStatus:  http.StatusOK,
			requestBody: `{}`,
			tokenUserID: 2,
		},
			DrawingID: 1,
			PointsCoordsResult: map[int][2]float64{
				0: {0, 0},
				1: {0, 0},
				2: {0.27, 1.25},
			}},
		{TestCase: TestCase{
			name:        "OK Coordinates",
			url:         "/drawings/1/points/2",
			method:      http.MethodPut,
			wantStatus:  http.StatusOK,
			requestBody: `{"point":{"x":1.32,"y":3.1},"measures":{"length":"m"}}`,
			tokenUserID: 2,
		},
			DrawingID: 1,
			PointsCoordsResult: map[int][2]float64{
				0: {0, 0},
				1: {1.32, 3.1},
				2: {0.27, 1.25},
			}},
		{TestCase: TestCase{
			name:        "OK Direction",
			url:         "/drawings/1/points/2",
			method:      http.MethodPut,
			wantStatus:  http.StatusOK,
			requestBody: `{"point":{"distance":132,"direction":90},"measures":{"length":"cm","angle":"deg"}}`,
			tokenUserID: 2,
		},
			DrawingID: 1,
			PointsCoordsResult: map[int][2]float64{
				0: {0, 0},
				1: {0, 1.32},
				2: {0.27, 1.25},
			}},
		{TestCase: TestCase{
			name:        "OK Angle",
			url:         "/drawings/1/points/3",
			method:      http.MethodPut,
			wantStatus:  http.StatusOK,
			requestBody: `{"point":{"distance":3,"angle":270},"measures":{"length":"dm","angle":"deg"}}`,
			tokenUserID: 2,
		},
			DrawingID: 1,
			PointsCoordsResult: map[int][2]float64{
				0: {0, 0},
				1: {0, 1.25},
				2: {-0.3, 1.25},
			}},
		{TestCase: TestCase{
			name:        "Not found point number",
			url:         "/drawings/1/points/42",
			method:      http.MethodPut,
			wantStatus:  http.StatusNotFound,
			requestBody: `{"point":{"distance":3,"angle":270},"measures":{"length":"dm","angle":"deg"}}`,
			tokenUserID: 2,
		}},
		{TestCase: TestCase{
			name:        "Not found drawing ID",
			url:         "/drawings/2/points/1",
			method:      http.MethodPut,
			wantStatus:  http.StatusNotFound,
			requestBody: `{"point":{"distance":3,"angle":270},"measures":{"length":"dm","angle":"deg"}}`,
			tokenUserID: 2,
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := newMockStorage()
			checkTestCase(t, tt.TestCase, storage)
			if tt.DrawingID != 0 {
				d, _ := storage.GetDrawing(tt.DrawingID)
				d.RoundAllPoints(2)
				for i, coords := range tt.PointsCoordsResult {
					p := d.Points[i]
					if p.X != coords[0] || p.Y != coords[1] {
						t.Errorf("Got wrong point coordinates. Got %v, want %v", d.Points[i], coords)
					}
				}
			}
		})
	}
}
