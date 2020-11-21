package api

import (
	"context"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/maxsid/goCeilings/drawing/raster"
	"github.com/maxsid/goCeilings/value"
	"image/png"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

type middlewareKey int

const (
	userIDMiddlewareKey = middlewareKey(iota)
	drawingMiddlewareKey
)

const defaultAddress = "127.0.0.1:8081"

func Run(addr string, st Storage) error {
	if addr == "" {
		addr = defaultAddress
	}
	router := mux.NewRouter()
	addMiddlewaresToRouter(router, st)
	addHandlersToRouter(router, st)

	log.Printf("API listening on %s...", addr)
	if err := http.ListenAndServe(addr, router); err != nil {
		return err
	}
	return nil
}

func addMiddlewaresToRouter(router *mux.Router, st Storage) {
	router.Use(authorizationMiddleware)
	router.Use(gettingDrawingMiddleware(st))
}

func addHandlersToRouter(router *mux.Router, st Storage) {
	router.HandleFunc("/login", loginHandler(st)).Methods(http.MethodPost)

	router.HandleFunc("/users", usersListHandler(st)).Methods(http.MethodGet)
	router.HandleFunc("/users", userCreatingHandler(st)).Methods(http.MethodPost)

	router.HandleFunc("/users/{id:[0-9]+}", userGettingHandler(st)).Methods(http.MethodGet)
	router.HandleFunc("/users/{id:[0-9]+}", userUpdatingHandler(st)).Methods(http.MethodPut)
	router.HandleFunc("/users/{id:[0-9]+}", userRemovingHandler(st)).Methods(http.MethodDelete)

	router.HandleFunc("/drawings", drawingsListGettingHandler(st)).Methods(http.MethodGet)
	router.HandleFunc("/drawings", drawingCreatingHandler(st)).Methods(http.MethodPost)

	router.HandleFunc("/drawings/{id:[0-9]+}", drawingGettingHandler).Methods(http.MethodGet)
	router.HandleFunc("/drawings/{id:[0-9]+}", drawingDeletingHandler(st)).Methods(http.MethodDelete)

	router.HandleFunc("/drawings/{id:[0-9]+}/image", drawingImageHandler).Methods(http.MethodGet)

	router.HandleFunc("/drawings/{id:[0-9]+}/points", drawingPointsHandler).Methods(http.MethodGet)
	router.HandleFunc("/drawings/{id:[0-9]+}/points", drawingPointsAddingHandler(st)).Methods(http.MethodPost)

	router.HandleFunc("/drawings/{id:[0-9]+}/points/{point_num:[0-9]+}", drawingPointGettingHandler).Methods(http.MethodGet)
	router.HandleFunc("/drawings/{id:[0-9]+}/points/{point_num:[0-9]+}", drawingPointUpdatingHandler(st)).Methods(http.MethodPut)
	router.HandleFunc("/drawings/{id:[0-9]+}/points/{point_num:[0-9]+}", drawingPointDeletingHandler(st)).Methods(http.MethodDelete)
}

// ===========
// Middlewares
// ===========

func authorizationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		var (
			onlyAdminURLReg    = regexp.MustCompile(`/users[a-zA-Z0-9/]*`)
			withoutLoginURLReg = regexp.MustCompile(`/login`)
		)

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
			ctx := context.WithValue(req.Context(), userIDMiddlewareKey, userClaims.ID)
			req = req.WithContext(ctx)
			if ok := onlyAdminURLReg.MatchString(req.URL.Path); ok && userClaims.Permission != AdminPermission {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
		}
		next.ServeHTTP(w, req)
	})
}

func gettingDrawingMiddleware(getter DrawingGetter) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			var (
				drawingPathReg = regexp.MustCompile(`/drawings/[0-9]+.*`)
			)
			if drawingPathReg.MatchString(req.URL.Path) {
				userID, ok := req.Context().Value(userIDMiddlewareKey).(uint)
				if !ok {
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					log.Print(ErrCouldNotReadUserIDFromCtx)
					return
				}
				drawingID, err := strconv.ParseUint(mux.Vars(req)["id"], 10, 64)
				if err != nil {
					http.Error(w, "Bad Request", http.StatusBadRequest)
					log.Print(err)
					return
				}

				drawing, err := getter.GetDrawingOfUser(userID, uint(drawingID))
				if errors.Is(err, ErrDrawingNotFound) {
					http.Error(w, "Not Found", http.StatusNotFound)
					return
				} else if err != nil {
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					log.Print(err)
					return
				}
				ctx := context.WithValue(req.Context(), drawingMiddlewareKey, drawing)
				req = req.WithContext(ctx)
			}
			next.ServeHTTP(w, req)
		})
	}
}

// ======
// /login
// ======

func loginHandler(getter UserGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		var userAuth User
		if !unmarshalRequestBody(w, req, &userAuth) {
			return
		}
		if userAuth.Login == "" || userAuth.Password == "" {
			http.Error(w, "Bad Request: Login or Password weren't specified", http.StatusBadRequest)
			return
		}
		user, err := getter.GetUser(userAuth.Login, userAuth.Password)
		if errors.Is(err, ErrUserNotFound) {
			http.Error(w, "User not found!", http.StatusNotFound)
			return
		} else if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			log.Print(err)
			return
		}

		token, err := createUserJWTToken(user.UserOpen, SigningSecret, time.Hour*24)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			log.Print(err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintf(w, `{"token":"%s"}`, token)
	}
}

// ======
// /users
// ======

func usersListHandler(getter UsersListGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		amount, err := getter.UsersAmount()
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			log.Print(err)
			return
		}
		stat, err := readListStatData(req.URL.Query(), amount)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			log.Print(err)
			return
		}

		respData := UsersListResponseData{ListStatData: *stat}
		respData.Users, err = getter.GetUsersList(respData.Page, respData.PageLimit)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			log.Print(err)
			return
		}

		marshalAndWrite(w, &respData)
	}
}

func userCreatingHandler(creator UserCreator) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		var user User
		if !unmarshalRequestBody(w, req, &user) {
			return
		}
		if user.Login == "" || user.Password == "" {
			http.Error(w, "Bad Request: Login or Password weren't specified", http.StatusBadRequest)
			return
		}
		if user.Permission == 0 {
			user.Permission = UserPermission
		}
		if err := creator.CreateUsers(&user); err != nil {
			if errors.Is(err, ErrUserAlreadyExist) {
				http.Error(w, "User already exists", http.StatusBadRequest)
				return
			} else {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				log.Print(err)
			}
			return
		}
		w.Header().Add("Location", fmt.Sprintf("/users/%d", user.ID))
		w.WriteHeader(http.StatusCreated)
	}
}

// ===========
// /users/{id}
// ===========

func userGettingHandler(getter UserGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		userID, err := strconv.ParseUint(mux.Vars(req)["id"], 10, 64)
		if err != nil {
			http.Error(w, "Bad Request Error: Incorrect user ID", http.StatusBadRequest)
			return
		}

		user, err := getter.GetUserByID(uint(userID))
		if errors.Is(err, ErrUserNotFound) {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		} else if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			log.Print(err)
			return
		}

		marshalAndWrite(w, &user.UserOpen)
	}
}

func userRemovingHandler(remover UserRemover) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		userID, err := strconv.ParseUint(mux.Vars(req)["id"], 10, 64)
		if err != nil {
			http.Error(w, "Bad Request Error: Incorrect user ID", http.StatusBadRequest)
			return
		}

		err = remover.RemoveUser(uint(userID))
		if errors.Is(err, ErrUserNotFound) {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		} else if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			log.Print(err)
			return
		}
	}
}

func userUpdatingHandler(updater UserUpdater) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		userID, err := strconv.ParseUint(mux.Vars(req)["id"], 10, 64)
		if err != nil {
			http.Error(w, "Bad Request Error: Incorrect user ID", http.StatusBadRequest)
			return
		}

		var user User
		if !unmarshalRequestBody(w, req, &user) {
			return
		}

		user.ID = uint(userID)
		err = updater.UpdateUser(&user)
		if errors.Is(err, ErrUserNotFound) {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		} else if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			log.Print(err)
			return
		}
	}
}

// =========
// /drawings
// =========

func drawingsListGettingHandler(getter DrawingsListGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		userID, ok := req.Context().Value(userIDMiddlewareKey).(uint)
		if !ok {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			log.Print(ErrCouldNotReadUserIDFromCtx)
			return
		}

		amount, err := getter.DrawingsAmount(userID)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			log.Print(err)
			return
		}
		stat, err := readListStatData(req.URL.Query(), amount)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			log.Print(err)
			return
		}

		respData := DrawingsListResponseData{Drawings: make([]*DrawingOpen, 0), ListStatData: *stat}
		if amount > 0 {
			respData.Drawings, err = getter.GetDrawingsList(userID, respData.Page, respData.PageLimit)
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				log.Print(err)
				return
			}
		}

		marshalAndWrite(w, &respData)
	}
}

func drawingCreatingHandler(creator DrawingCreator) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		userID, ok := req.Context().Value(userIDMiddlewareKey).(uint)
		if !ok {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			log.Print(ErrCouldNotReadUserIDFromCtx)
			return
		}

		var requestData DrawingPostPutRequestData
		if !unmarshalRequestBody(w, req, &requestData) {
			return
		}
		if requestData.Name == "" {
			http.Error(w, "Bad Request: Wrong JSON body", http.StatusBadRequest)
			return
		}
		drawing := Drawing{DrawingOpen: requestData.DrawingOpen, GGDrawing: *raster.NewEmptyGGDrawing()}
		drawing.Measures = requestData.Measures.ToFigureMeasures(drawing.Measures)

		if err := addRequestPointsToDrawing(&drawing, requestData.Points...); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := creator.CreateDrawings(userID, &drawing); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			log.Print(err)
			return
		}

		w.Header().Add("Location", fmt.Sprintf("/drawings/%d", drawing.ID))
		http.Error(w, "", http.StatusCreated)
	}
}

// ==============
// /drawings/{id}
// =============

func drawingGettingHandler(w http.ResponseWriter, req *http.Request) {
	drawing, ok := req.Context().Value(drawingMiddlewareKey).(*Drawing)
	if !ok {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Print(ErrCouldNotReadDrawingFromCtx)
		return
	}

	respData := DrawingGetResponseData{
		DrawingOpen: drawing.DrawingOpen,
		Points:      drawing.GetPoints(),
		DrawingCalculatingData: DrawingCalculatingData{
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

func drawingDeletingHandler(remover DrawingRemover) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		userID, ok := req.Context().Value(userIDMiddlewareKey).(uint)
		if !ok {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			log.Print(ErrCouldNotReadUserIDFromCtx)
			return
		}
		drawingID, err := strconv.ParseUint(mux.Vars(req)["id"], 10, 64)
		if err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		if err := remover.RemoveDrawingOfUser(userID, uint(drawingID)); errors.Is(err, ErrDrawingNotFound) {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		} else if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			log.Print(err)
			return
		}
	}
}

// ====================
// /drawings/{id}/image
// ====================

func drawingImageHandler(w http.ResponseWriter, req *http.Request) {
	drawing, ok := req.Context().Value(drawingMiddlewareKey).(*Drawing)
	if !ok {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Print(ErrCouldNotReadDrawingFromCtx)
		return
	}

	var err error
	drawDescription := false
	if info, ok := req.URL.Query()["info"]; ok && len(info) != 0 {
		drawDescription, err = strconv.ParseBool(info[0])
		if err != nil {
			http.Error(w, "Bad Request: Incorrect 'info' parameter", http.StatusBadRequest)
			return
		}
	}

	image, err := drawing.Draw(drawDescription)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Print(err)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	err = png.Encode(w, image)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Print(err)
		return
	}
}

// =====================
// /drawings/{id}/points
// =====================

func drawingPointsHandler(w http.ResponseWriter, req *http.Request) {
	drawing, ok := req.Context().Value(drawingMiddlewareKey).(*Drawing)
	if !ok {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Print(ErrCouldNotReadDrawingFromCtx)
		return
	}

	measure, precision, ok := readMeasureAndPrecisionFromURL(w, req, drawing.Measures.Length, 2)
	if !ok {
		return
	}
	respData := DrawingPointsGettingResponseData{
		DrawingOpen: drawing.DrawingOpen,
		Points:      drawing.GetPointsWithParams(measure, precision),
		Measure:     value.NameOfLengthMeasure(measure),
	}

	marshalAndWrite(w, &respData)
}

func drawingPointsAddingHandler(updater DrawingUpdater) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		drawing, ok := req.Context().Value(drawingMiddlewareKey).(*Drawing)
		if !ok {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			log.Print(ErrCouldNotReadDrawingFromCtx)
			return
		}

		var reqData PointsCalculatingWithMeasures
		if !unmarshalRequestBody(w, req, &reqData) {
			return
		}

		dmCopy := drawing.Measures
		drawing.Measures = reqData.Measures.ToFigureMeasures(drawing.Measures)

		if err := addRequestPointsToDrawing(drawing, reqData.Points...); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		respData := DrawingPointsGettingResponseData{
			DrawingOpen: drawing.DrawingOpen,
			Points:      drawing.GetPointsWithParams(drawing.Measures.Length, 2),
			Measure:     reqData.Measures.Length,
		}

		drawing.Measures = dmCopy
		err := updater.UpdateDrawing(drawing)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			log.Print(err)
			return
		}

		marshalAndWrite(w, &respData)
	}
}

// =========================
// /drawings/{id}/points/{n}
// =========================

func drawingPointGettingHandler(w http.ResponseWriter, req *http.Request) {
	drawing, ok := req.Context().Value(drawingMiddlewareKey).(*Drawing)
	if !ok {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Print(ErrCouldNotReadDrawingFromCtx)
		return
	}
	pointNumber, err := strconv.ParseUint(mux.Vars(req)["point_num"], 10, 64)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	if uint(drawing.Len()) <= uint(pointNumber) {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	measure, precision, ok := readMeasureAndPrecisionFromURL(w, req, drawing.Measures.Length, 2)
	if !ok {
		return
	}
	point := drawing.Points[pointNumber]
	marshalAndWrite(w, PointWithMeasure{
		X:       value.ConvertFromOneRound(measure, point.X, precision),
		Y:       value.ConvertFromOneRound(measure, point.Y, precision),
		Measure: value.NameOfLengthMeasure(measure),
	})
}

func drawingPointDeletingHandler(updater DrawingUpdater) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		drawing, ok := req.Context().Value(drawingMiddlewareKey).(*Drawing)
		if !ok {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			log.Print(ErrCouldNotReadDrawingFromCtx)
			return
		}
		pointNumber, err := strconv.ParseUint(mux.Vars(req)["point_num"], 10, 64)
		if err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		drawing.Points = append(drawing.Points[:pointNumber], drawing.Points[pointNumber+1:]...)

		if err := updater.UpdateDrawing(drawing); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			log.Print(err)
			return
		}
	}
}

func drawingPointUpdatingHandler(updater DrawingUpdater) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		drawing, ok := req.Context().Value(drawingMiddlewareKey).(*Drawing)
		if !ok {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			log.Print(ErrCouldNotReadDrawingFromCtx)
			return
		}
		pointNumber, err := strconv.ParseUint(mux.Vars(req)["point_num"], 10, 64)
		if err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		if uint(drawing.Len()) <= uint(pointNumber) {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}

		var pointWithMeasure PointCalculatingWithMeasures
		if !unmarshalRequestBody(w, req, &pointWithMeasure) {
			return
		}

		endPoints := drawing.Points[pointNumber+1:]
		drawing.Points = drawing.Points[:pointNumber]

		drawingMeasures := drawing.Measures
		drawing.Measures = pointWithMeasure.Measures.ToFigureMeasures(drawing.Measures)

		if err := addRequestPointsToDrawing(drawing, &pointWithMeasure.PointCalculating); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		drawing.Measures = drawingMeasures
		drawing.Points = append(drawing.Points, endPoints...)

		if err := updater.UpdateDrawing(drawing); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			log.Print(err)
			return
		}
	}
}
