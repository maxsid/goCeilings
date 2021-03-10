package api

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"time"

	"github.com/gorilla/mux"
	"github.com/maxsid/goCeilings/drawing/raster"
	"github.com/maxsid/goCeilings/server/common"
	"github.com/maxsid/goCeilings/value"
	"github.com/urfave/negroni"
)

const ctxKeyUserStorage = ctxKey(iota)

const defaultAddress = "127.0.0.1:8081"

const (
	pathVarUserID      = pathVarKey("user_id")
	pathVarDrawingID   = pathVarKey("drawing_id")
	pathVarPointNumber = pathVarKey("point_num")
)

const (
	urlParamInfo      = urlParamKey("info")
	urlParamPage      = urlParamKey("p")
	urlParamPageLimit = urlParamKey("lim")
	urlParamPrecision = urlParamKey("p")
	urlParamMeasure   = urlParamKey("m")
)

// Run runs the REST API server.
func Run(addr string, st common.Storage) error {
	if addr == "" {
		addr = defaultAddress
	}
	router := mux.NewRouter()
	addHandlersToRouter(router, st)
	addMiddlewaresToRouter(router, st)

	n := getNegroniHandler(router)

	log.Printf("API listening on %s...", addr)
	return http.ListenAndServe(addr, n)
}

// addMiddlewaresToRouter adds all middlewares into router.
func addMiddlewaresToRouter(router *mux.Router, st common.Storage) {
	router.Use(getAuthorizationMiddleware(st))
	router.Use(mux.CORSMethodMiddleware(router))
}

// addHandlersToRouter adds all REST API handlers into router.
func addHandlersToRouter(router *mux.Router, st common.Storage) {
	router.HandleFunc("/login", loginHandler(st)).Methods(http.MethodPost)

	path := "/users"
	router.HandleFunc(path, usersListHandler).Methods(http.MethodGet)
	router.HandleFunc(path, userCreatingHandler).Methods(http.MethodPost)

	path = fmt.Sprintf("/users/{%s:[0-9]+}", pathVarUserID)
	router.HandleFunc(path, userGettingHandler).Methods(http.MethodGet)
	router.HandleFunc(path, userUpdatingHandler).Methods(http.MethodPut)
	router.HandleFunc(path, userDeletingHandler).Methods(http.MethodDelete)

	path = fmt.Sprintf("/users/{%s:[0-9]+}/permissions", pathVarUserID)
	router.HandleFunc(path, permissionsOfUserGettingHandler).Methods(http.MethodGet)
	router.HandleFunc(path, permissionCreatingHandler).Methods(http.MethodPost)

	path = fmt.Sprintf("/users/{%s:[0-9]+}/permissions/drawings/{%s:[0-9]+}", pathVarUserID, pathVarDrawingID)
	router.HandleFunc(path, permissionGetterAndDeletingHandler).Methods(http.MethodGet)
	router.HandleFunc(path, permissionGetterAndDeletingHandler).Methods(http.MethodDelete)

	path = "/drawings"
	router.HandleFunc(path, drawingsListGettingHandler).Methods(http.MethodGet)
	router.HandleFunc(path, drawingCreatingHandler).Methods(http.MethodPost)

	path = fmt.Sprintf("/drawings/{%s:[0-9]+}", pathVarDrawingID)
	router.HandleFunc(path, drawingGettingHandler).Methods(http.MethodGet)
	router.HandleFunc(path, drawingDeletingHandler).Methods(http.MethodDelete)

	path = fmt.Sprintf("/drawings/{%s:[0-9]+}/image", pathVarDrawingID)
	router.HandleFunc(path, drawingImageHandler).Methods(http.MethodGet)

	path = fmt.Sprintf("/drawings/{%s:[0-9]+}/permissions", pathVarDrawingID)
	router.HandleFunc(path, permissionsOfDrawingGettingHandler).Methods(http.MethodGet)
	router.HandleFunc(path, permissionCreatingHandler).Methods(http.MethodPost)

	path = fmt.Sprintf("/drawings/{%s:[0-9]+}/permissions/users/{%s:[0-9]+}", pathVarDrawingID, pathVarUserID)
	router.HandleFunc(path, permissionGetterAndDeletingHandler).Methods(http.MethodGet)
	router.HandleFunc(path, permissionGetterAndDeletingHandler).Methods(http.MethodDelete)

	path = fmt.Sprintf("/drawings/{%s:[0-9]+}/points", pathVarDrawingID)
	router.HandleFunc(path, drawingPointsListGettingHandler).Methods(http.MethodGet)
	router.HandleFunc(path, drawingPointsAddingHandler).Methods(http.MethodPost)

	path = fmt.Sprintf("/drawings/{%s:[0-9]+}/points/{%s:[0-9]+}", pathVarDrawingID, pathVarPointNumber)
	router.HandleFunc(path, drawingPointGettingHandler).Methods(http.MethodGet)
	router.HandleFunc(path, drawingPointUpdatingHandler).Methods(http.MethodPut)
	router.HandleFunc(path, drawingPointDeletingHandler).Methods(http.MethodDelete)
}

// drawingCreatingHandler handles creating one drawing by drawingPostPutRequestData body.
// Handles: POST /drawings
func drawingCreatingHandler(w http.ResponseWriter, req *http.Request) {
	var storage common.UserStorage
	if storage = getUserStorageOrWriteError(w, req); storage == nil {
		return
	}
	user := storage.GetCurrentUser()

	var requestData drawingPostPutRequestData
	if err := unmarshalReaderContent(req.Body, &requestData); writeError(w, err) {
		return
	}
	if requestData.Name == "" {
		http.Error(w, "Bad Request: Wrong JSON body", http.StatusBadRequest)
		return
	}
	drawing := common.Drawing{DrawingBasic: requestData.DrawingBasic, GGDrawing: *raster.NewEmptyGGDrawing()}
	drawing.Measures = requestData.Measures.ToFigureMeasures(drawing.Measures)

	if err := drawing.AddPoints(getPointsFromRequestPoint(requestData.Points...)...); writeError(w, err) {
		return
	}

	if err := storage.CreateDrawings(user.ID, &drawing); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Print(err)
		return
	}

	w.Header().Add("Location", fmt.Sprintf("/drawings/%d", drawing.ID))
	http.Error(w, "", http.StatusCreated)
}

// drawingDeletingHandler handles deleting one drawing by its ID.
// Handles: DELETE /drawings/{id}
func drawingDeletingHandler(w http.ResponseWriter, req *http.Request) {
	var storage common.UserStorage
	if storage = getUserStorageOrWriteError(w, req); storage == nil {
		return
	}

	drawingID := uint(0)
	if err := parsePathValue(mux.Vars(req), pathVarDrawingID, &drawingID); writeError(w, err) {
		return
	}

	if err := storage.RemoveDrawing(drawingID); writeError(w, err) {
		return
	}
}

// drawingGettingHandler handles getting one drawing by ID and presents it as drawingGetResponseData type.
// Handles: GET /drawings/{id}
func drawingGettingHandler(w http.ResponseWriter, req *http.Request) {
	drawing, _ := getDrawingByRequestOrWriteError(w, req)
	if drawing == nil {
		return
	}

	respData := drawingGetResponseData{
		DrawingBasic: drawing.DrawingBasic,
		Points:       drawing.GetPoints(),
		drawingCalculatedData: drawingCalculatedData{
			Area:        drawing.Area(),
			Perimeter:   drawing.Perimeter(),
			PointsCount: drawing.Len(),
			Width:       drawing.Width(),
			Height:      drawing.Height(),
		},
		Measures: drawing.Measures.ToFigureMeasuresNames(),
	}

	marshalAndWrite(w, &respData)
}

// drawingImageHandler handle getting an image of the drawing by its ID.
// Handles: GET /drawings/{id}/image
func drawingImageHandler(w http.ResponseWriter, req *http.Request) {
	drawing, _ := getDrawingByRequestOrWriteError(w, req)
	if drawing == nil {
		return
	}

	var err error
	drawDescription := false
	if err := parseURLParamValue(req.URL.Query(), urlParamInfo, &drawDescription); err != nil && !errors.Is(err, ErrNotFound) && writeError(w, err) {
		return
	}
	drawer := drawing.GetDrawer()
	imageBytes, err := drawer.Draw(drawDescription)
	if writeError(w, err) {
		return
	}

	w.Header().Set("Content-Type", drawer.DrawingMIME())
	_, _ = w.Write(imageBytes)
}

// drawingsListGettingHandler handles getting a list of drawings the current user
// and presents it as drawingsListResponseData.
// Handles: GET /drawings
func drawingsListGettingHandler(w http.ResponseWriter, req *http.Request) {
	var storage common.UserStorage
	if storage = getUserStorageOrWriteError(w, req); storage == nil {
		return
	}
	user := storage.GetCurrentUser()

	amount, err := storage.DrawingsAmount(user.ID)
	if writeError(w, err) {
		return
	}
	stat, err := readListStatData(req.URL.Query(), amount)
	if writeError(w, err) {
		return
	}

	respData := drawingsListResponseData{Drawings: make([]*common.DrawingBasic, 0), listStatData: *stat}
	if amount > 0 {
		respData.Drawings, err = storage.GetDrawingsList(user.ID, respData.Page, respData.PageLimit)
		if writeError(w, err) {
			return
		}
	}

	marshalAndWrite(w, &respData)
}

// drawingPointsAddingHandler handles adding new points into the drawing by its ID and pointsCalculatingWithMeasures body.
// Handles: POST /drawings/{id}/points
func drawingPointsAddingHandler(w http.ResponseWriter, req *http.Request) {
	drawing, _ := getDrawingByRequestOrWriteError(w, req)
	if drawing == nil {
		return
	}

	var reqData pointsCalculatingWithMeasures
	if err := unmarshalReaderContent(req.Body, &reqData); writeError(w, err) {
		return
	}

	dmCopy := drawing.Measures
	drawing.Measures = reqData.Measures.ToFigureMeasures(drawing.Measures)

	if err := drawing.AddPoints(getPointsFromRequestPoint(reqData.Points...)...); writeError(w, err) {
		return
	}

	respData := drawingPointsGettingResponseData{
		DrawingBasic: drawing.DrawingBasic,
		Points:       drawing.GetPointsWithParams(drawing.Measures.Length, 2),
		Measure:      reqData.Measures.Length,
	}

	drawing.Measures = dmCopy

	var storage common.UserStorage
	if storage = getUserStorageOrWriteError(w, req); storage == nil {
		return
	}

	if err := storage.UpdateDrawing(drawing); writeError(w, err) {
		return
	}

	marshalAndWrite(w, &respData)
}

// drawingPointDeletingHandler handles deleting one point from the drawing by drawing ID and a number of the point.
// The first point of the drawing has a number one.
// Handles: DELETE /drawings/{id}/points/{number}
func drawingPointDeletingHandler(w http.ResponseWriter, req *http.Request) {
	drawing, _ := getDrawingByRequestOrWriteError(w, req)
	if drawing == nil {
		return
	}
	pointIndex, ok := getPointIndexByRequestOrWriteError(w, req, drawing)
	if !ok {
		return
	}

	drawing.Points = append(drawing.Points[:pointIndex], drawing.Points[pointIndex+1:]...)

	var storage common.UserStorage
	if storage = getUserStorageOrWriteError(w, req); storage == nil {
		return
	}

	if err := storage.UpdateDrawing(drawing); writeError(w, err) {
		return
	}
}

// drawingPointGettingHandler handles getting one point of a drawing by drawing ID and a number of the point.
// The first point of the drawing has a number one.
// Handles: GET /drawings/{id}/points/{number}
func drawingPointGettingHandler(w http.ResponseWriter, req *http.Request) {
	drawing, _ := getDrawingByRequestOrWriteError(w, req)
	if drawing == nil {
		return
	}
	pointIndex, ok := getPointIndexByRequestOrWriteError(w, req, drawing)
	if !ok {
		return
	}

	precision, measure := 2, drawing.Measures.Length
	if err := readLengthMeasureAndPrecision(req.URL.Query(), &measure, &precision); writeError(w, err) {
		return
	}
	point := drawing.Points[pointIndex]
	marshalAndWrite(w, pointWithMeasure{
		X:       value.ConvertFromOneRound(measure, point.X, precision),
		Y:       value.ConvertFromOneRound(measure, point.Y, precision),
		Measure: value.NameOfLengthMeasure(measure),
	})
}

// drawingPointsListGettingHandler handles getting points of the drawing by its ID.
// Handles: GET /drawings/{id}/points
func drawingPointsListGettingHandler(w http.ResponseWriter, req *http.Request) {
	drawing, _ := getDrawingByRequestOrWriteError(w, req)
	if drawing == nil {
		return
	}

	precision, measure := 2, drawing.Measures.Length
	if err := readLengthMeasureAndPrecision(req.URL.Query(), &measure, &precision); writeError(w, err) {
		return
	}
	respData := drawingPointsGettingResponseData{
		DrawingBasic: drawing.DrawingBasic,
		Points:       drawing.GetPointsWithParams(measure, precision),
		Measure:      value.NameOfLengthMeasure(measure),
	}

	marshalAndWrite(w, &respData)
}

// drawingPointUpdatingHandler updates a point of the drawing by drawing ID, a number of the point and
// pointCalculatingWithMeasures body.
// Handles: PUT /drawings/{id}/points/{number}
func drawingPointUpdatingHandler(w http.ResponseWriter, req *http.Request) {
	drawing, _ := getDrawingByRequestOrWriteError(w, req)
	if drawing == nil {
		return
	}
	pointIndex, ok := getPointIndexByRequestOrWriteError(w, req, drawing)
	if !ok {
		return
	}

	var pointWithMeasure pointCalculatingWithMeasures
	if err := unmarshalReaderContent(req.Body, &pointWithMeasure); writeError(w, err) {
		return
	}

	drawingMeasures := drawing.Measures
	drawing.Measures = pointWithMeasure.Measures.ToFigureMeasures(drawing.Measures)

	point := getPointsFromRequestPoint(&pointWithMeasure.Point)[0]
	if err := drawing.SetPoint(pointIndex, point); writeError(w, err) {
		return
	}

	drawing.Measures = drawingMeasures

	var storage common.UserStorage
	if storage = getUserStorageOrWriteError(w, req); storage == nil {
		return
	}

	if err := storage.UpdateDrawing(drawing); writeError(w, err) {
		return
	}
}

// getAuthorizationMiddleware returns middleware authorization handler.
// Handles: Middleware
func getAuthorizationMiddleware(st common.Storage) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			var withoutLoginURLReg = regexp.MustCompile(`/login`)

			if !withoutLoginURLReg.MatchString(req.URL.Path) {
				auth := ""
				if auth = req.Header.Get("Authorization"); auth == "" || auth[:6] != "Bearer" {
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}
				userClaims, err := readUserJWTToken(auth[7:], SigningSecret)
				if err != nil {
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}
				userSt, err := st.GetUserStorage(&userClaims.UserBasic)
				if writeError(w, err) {
					return
				}
				ctx := context.WithValue(req.Context(), ctxKeyUserStorage, userSt)
				req = req.WithContext(ctx)
			}
			next.ServeHTTP(w, req)
		})
	}
}

func getNegroniHandler(router http.Handler) http.Handler {
	rec := negroni.NewRecovery()
	rec.PrintStack = false

	n := negroni.New()
	n.Use(rec)
	n.Use(negroni.NewLogger())
	n.UseHandler(router)
	return n
}

// loginHandler returns authentication handler.
// Handles: POST /login
func loginHandler(getter common.UserGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		var userAuth common.UserConfident
		if err := unmarshalReaderContent(req.Body, &userAuth); writeError(w, err) {
			return
		}
		if userAuth.Login == "" || userAuth.Password == "" {
			http.Error(w, "Bad Request: Login or Password weren't specified", http.StatusBadRequest)
			return
		}
		user, err := getter.GetUser(userAuth.Login, userAuth.Password)
		if writeError(w, err) {
			return
		}

		token, err := createUserJWTToken(user.UserBasic, SigningSecret, time.Hour*24)
		if writeError(w, err) {
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintf(w, `{"token":"%s"}`, token)
	}
}

// permissionCreatingHandler handles adding a drawing permission by user ID or drawing ID, and
// drawingPermissionCreating body.
// Handles: POST /drawings/{id}/permissions, POST /users/{id}/permissions
func permissionCreatingHandler(w http.ResponseWriter, req *http.Request) {
	storage := getUserStorageOrWriteError(w, req)
	body := new(drawingPermissionCreating)
	if err := unmarshalReaderContent(req.Body, body); writeError(w, err) {
		return
	}
	if body.UserID == 0 {
		if err := parsePathValue(mux.Vars(req), pathVarUserID, &body.UserID); writeError(w, err) {
			return
		}
	}
	if body.DrawingID == 0 {
		if err := parsePathValue(mux.Vars(req), pathVarDrawingID, &body.DrawingID); writeError(w, err) {
			return
		}
	}
	perm := common.DrawingPermission{
		User:    &common.UserBasic{ID: body.UserID},
		Drawing: &common.DrawingBasic{ID: body.DrawingID},
		Get:     body.Get,
		Change:  body.Change,
		Delete:  body.Delete,
		Share:   body.Share,
	}
	if err := storage.CreateDrawingPermission(&perm); writeError(w, err) {
		return
	}
	w.Header().Set("Location", fmt.Sprintf("/drawings/%d/permissions/users/%d", body.DrawingID, body.UserID))
	w.WriteHeader(http.StatusCreated)
}

// permissionGetterAndDeletingHandler handles getting or deleting one drawing permissions by user ID and drawing ID.
// Handles: GET|DELETE /users/{id}/permissions/drawings/{id}, GET|DELETE /drawings/{id}/permissions/users/{id}
func permissionGetterAndDeletingHandler(w http.ResponseWriter, req *http.Request) {
	storage := getUserStorageOrWriteError(w, req)
	var drawingID, userID uint
	if err := parsePathValue(mux.Vars(req), pathVarDrawingID, &drawingID); writeError(w, err) {
		return
	}
	if err := parsePathValue(mux.Vars(req), pathVarUserID, &userID); writeError(w, err) {
		return
	}
	// GET Method
	if req.Method == http.MethodGet {
		perm, err := storage.GetDrawingPermission(userID, drawingID)
		if writeError(w, err) {
			return
		}
		marshalAndWrite(w, perm)
		return
	}
	// DELETE Method
	err := storage.RemoveDrawingPermission(userID, drawingID)
	_ = writeError(w, err)
}

// permissionsOfUserGettingHandler handles getting drawing permissions of a user by ID.
// Handles: GET /users/{id}/permissions
func permissionsOfUserGettingHandler(w http.ResponseWriter, req *http.Request) {
	storage := getUserStorageOrWriteError(w, req)
	userID := uint(0)
	if err := parsePathValue(mux.Vars(req), pathVarUserID, &userID); writeError(w, err) {
		return
	}
	permissions, err := storage.GetDrawingsPermissionsOfUser(userID)
	if writeError(w, err) {
		return
	}
	marshalAndWrite(w, &permissions)
}

// permissionsOfDrawingGettingHandler handles getting drawing permissions of a drawing by its ID.
// Handles: GET /drawings/{id}/permissions
func permissionsOfDrawingGettingHandler(w http.ResponseWriter, req *http.Request) {
	storage := getUserStorageOrWriteError(w, req)
	drawingID := uint(0)
	if err := parsePathValue(mux.Vars(req), pathVarDrawingID, &drawingID); writeError(w, err) {
		return
	}
	permissions, err := storage.GetDrawingsPermissionsOfDrawing(drawingID)
	if writeError(w, err) {
		return
	}
	marshalAndWrite(w, &permissions)
}

// userCreatingHandler handles the user creating by UserConfident body.
// Handles: POST /uses
func userCreatingHandler(w http.ResponseWriter, req *http.Request) {
	var storage common.UserStorage
	if storage = getUserStorageOrWriteError(w, req); storage == nil {
		return
	}

	var user common.UserConfident
	if err := unmarshalReaderContent(req.Body, &user); writeError(w, err) {
		return
	}
	if user.Login == "" || user.Password == "" {
		_ = writeError(w, ErrBadLoginOrPassword)
		return
	}
	if user.Role == 0 {
		user.Role = common.RoleUser
	}
	if err := storage.CreateUsers(&user); writeError(w, err) {
		return
	}
	w.Header().Add("Location", fmt.Sprintf("/users/%d", user.ID))
	w.WriteHeader(http.StatusCreated)
}

// userDeletingHandler handles removing of one user by ID.
// Handles: DELETE /users/{id}
func userDeletingHandler(w http.ResponseWriter, req *http.Request) {
	var storage common.UserStorage
	if storage = getUserStorageOrWriteError(w, req); storage == nil {
		return
	}

	userID := uint(0)
	if err := parsePathValue(mux.Vars(req), pathVarUserID, &userID); writeError(w, err) {
		return
	}

	if err := storage.RemoveUser(userID); writeError(w, err) {
		return
	}
}

// userGettingHandler handles getting of one user by ID.
// Handles: GET /user/{id}
func userGettingHandler(w http.ResponseWriter, req *http.Request) {
	var storage common.UserStorage
	if storage = getUserStorageOrWriteError(w, req); storage == nil {
		return
	}

	userID := uint(0)
	if err := parsePathValue(mux.Vars(req), pathVarUserID, &userID); writeError(w, err) {
		return
	}

	user, err := storage.GetUserByID(userID)
	if writeError(w, err) {
		return
	}

	marshalAndWrite(w, &user.UserBasic)
}

// usersListHandler handles request of getting users list.
// Handles: GET /users
func usersListHandler(w http.ResponseWriter, req *http.Request) {
	var storage common.UserStorage
	if storage = getUserStorageOrWriteError(w, req); storage == nil {
		return
	}

	amount, err := storage.UsersAmount()
	if writeError(w, err) {
		return
	}
	stat, err := readListStatData(req.URL.Query(), amount)
	if writeError(w, err) {
		return
	}

	respData := usersListResponseData{listStatData: *stat}
	respData.Users, err = storage.GetUsersList(respData.Page, respData.PageLimit)
	if writeError(w, err) {
		return
	}

	marshalAndWrite(w, &respData)
}

// userUpdatingHandler handles updating of one user by ID and UserConfident body.
// Handles: PUT /users/{id}
func userUpdatingHandler(w http.ResponseWriter, req *http.Request) {
	var storage common.UserStorage
	if storage = getUserStorageOrWriteError(w, req); storage == nil {
		return
	}

	userID := uint(0)
	if err := parsePathValue(mux.Vars(req), pathVarUserID, &userID); writeError(w, err) {
		return
	}

	var user common.UserConfident
	if err := unmarshalReaderContent(req.Body, &user); writeError(w, err) {
		return
	}

	user.ID = userID
	err := storage.UpdateUser(&user)
	_ = writeError(w, err)
}
