package api

import (
	"context"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/maxsid/goCeilings/drawing/raster"
	"github.com/maxsid/goCeilings/value"
	"github.com/urfave/negroni"
	"image/png"
	"log"
	"net/http"
	"regexp"
	"time"
)

const (
	ctxKeyUser = ctxKey(iota)
	ctxKeyDrawing
	ctxKeyPointIndex
	ctxKeyPoint
)

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

const defaultAddress = "127.0.0.1:8081"

func Run(addr string, st Storage) error {
	if addr == "" {
		addr = defaultAddress
	}
	router := mux.NewRouter()
	addHandlersToRouter(router, st)
	addMiddlewaresToRouter(router, st)

	n := getNegroniHandler(router)

	log.Printf("API listening on %s...", addr)
	if err := http.ListenAndServe(addr, n); err != nil {
		return err
	}
	return nil
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

func addMiddlewaresToRouter(router *mux.Router, st Storage) {
	router.Use(authorizationMiddleware)
	router.Use(gettingDrawingMiddleware(st))
	router.Use(gettingPointMiddleware)
	router.Use(mux.CORSMethodMiddleware(router))
}

func addHandlersToRouter(router *mux.Router, st Storage) {
	router.HandleFunc("/login", loginHandler(st)).Methods(http.MethodPost)

	path := "/users"
	router.HandleFunc(path, getUsersListHandler(st)).Methods(http.MethodGet)
	router.HandleFunc(path, getUserCreatingHandler(st)).Methods(http.MethodPost)

	path = fmt.Sprintf("/users/{%s:[0-9]+}", pathVarUserID)
	router.HandleFunc(path, getUserGettingHandler(st)).Methods(http.MethodGet)
	router.HandleFunc(path, getUserUpdatingHandler(st)).Methods(http.MethodPut)
	router.HandleFunc(path, getUserRemovingHandler(st)).Methods(http.MethodDelete)

	path = "/drawings"
	router.HandleFunc(path, getDrawingsListGettingHandler(st)).Methods(http.MethodGet)
	router.HandleFunc(path, getDrawingCreatingHandler(st)).Methods(http.MethodPost)

	path = fmt.Sprintf("/drawings/{%s:[0-9]+}", pathVarDrawingID)
	router.HandleFunc(path, drawingGettingHandler).Methods(http.MethodGet)
	router.HandleFunc(path, getDrawingDeletingHandler(st)).Methods(http.MethodDelete)

	path = fmt.Sprintf("/drawings/{%s:[0-9]+}/image", pathVarDrawingID)
	router.HandleFunc(path, drawingImageHandler).Methods(http.MethodGet)

	path = fmt.Sprintf("/drawings/{%s:[0-9]+}/points", pathVarDrawingID)
	router.HandleFunc(path, drawingPointsGettingHandler).Methods(http.MethodGet)
	router.HandleFunc(path, getDrawingPointsAddingHandler(st)).Methods(http.MethodPost)

	path = fmt.Sprintf("/drawings/{%s:[0-9]+}/points/{%s:[0-9]+}", pathVarDrawingID, pathVarPointNumber)
	router.HandleFunc(path, drawingPointGettingHandler).Methods(http.MethodGet)
	router.HandleFunc(path, getDrawingPointUpdatingHandler(st)).Methods(http.MethodPut)
	router.HandleFunc(path, getDrawingPointDeletingHandler(st)).Methods(http.MethodDelete)
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
			ctx := context.WithValue(req.Context(), ctxKeyUser, userClaims.UserOpen)
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
			var drawingPathReg = regexp.MustCompile(`/drawings/[0-9]+.*`)

			if drawingPathReg.MatchString(req.URL.Path) {
				var err error
				user, drawingID, drawing := new(UserOpen), uint(0), new(Drawing)
				if err := readCtxValue(req.Context(), ctxKeyUser, user); writeError(w, err) {
					return
				}
				if err := parsePathValue(mux.Vars(req), pathVarDrawingID, &drawingID); writeError(w, err) {
					return
				}

				drawing, err = getter.GetDrawingOfUser(user.ID, drawingID)
				if writeError(w, err) {
					return
				}
				ctx := context.WithValue(req.Context(), ctxKeyDrawing, drawing)
				req = req.WithContext(ctx)
			}
			next.ServeHTTP(w, req)
		})
	}
}

func gettingPointMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		var pointPathReg = regexp.MustCompile(`/drawings/[0-9]+/points/[0-9]+`)

		if pointPathReg.MatchString(req.URL.Path) {
			pointIndex, drawing := 0, new(Drawing)
			if err := readCtxValue(req.Context(), ctxKeyDrawing, drawing); writeError(w, err) {
				return
			}
			if err := parsePathValue(mux.Vars(req), pathVarPointNumber, &pointIndex); writeError(w, err) {
				return
			}
			if pointIndex > drawing.Len() || pointIndex < 1 {
				_ = writeError(w, ErrPointNotFound)
				return
			}
			pointIndex--
			ctx := context.WithValue(req.Context(), ctxKeyPointIndex, pointIndex)
			ctx = context.WithValue(ctx, ctxKeyPoint, drawing.Points[pointIndex])
			req = req.WithContext(ctx)
		}
		next.ServeHTTP(w, req)
	})
}

// ======
// /login
// ======

func loginHandler(getter UserGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		var userAuth User
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

		token, err := createUserJWTToken(user.UserOpen, SigningSecret, time.Hour*24)
		if writeError(w, err) {
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintf(w, `{"token":"%s"}`, token)
	}
}

// ======
// /users
// ======

func getUsersListHandler(getter UsersListGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		amount, err := getter.UsersAmount()
		if writeError(w, err) {
			return
		}
		stat, err := readListStatData(req.URL.Query(), amount)
		if writeError(w, err) {
			return
		}

		respData := UsersListResponseData{ListStatData: *stat}
		respData.Users, err = getter.GetUsersList(respData.Page, respData.PageLimit)
		if writeError(w, err) {
			return
		}

		marshalAndWrite(w, &respData)
	}
}

func getUserCreatingHandler(creator UserCreator) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		var user User
		if err := unmarshalReaderContent(req.Body, &user); writeError(w, err) {
			return
		}
		if user.Login == "" || user.Password == "" {
			_ = writeError(w, ErrBadLoginOrPassword)
			return
		}
		if user.Permission == 0 {
			user.Permission = UserPermission
		}
		if err := creator.CreateUsers(&user); writeError(w, err) {
			return
		}
		w.Header().Add("Location", fmt.Sprintf("/users/%d", user.ID))
		w.WriteHeader(http.StatusCreated)
	}
}

// ===========
// /users/{id}
// ===========

func getUserGettingHandler(getter UserGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		userID := uint(0)
		if err := parsePathValue(mux.Vars(req), pathVarUserID, &userID); writeError(w, err) {
			return
		}

		user, err := getter.GetUserByID(userID)
		if writeError(w, err) {
			return
		}

		marshalAndWrite(w, &user.UserOpen)
	}
}

func getUserRemovingHandler(remover UserRemover) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		userID := uint(0)
		if err := parsePathValue(mux.Vars(req), pathVarUserID, &userID); writeError(w, err) {
			return
		}

		if err := remover.RemoveUser(userID); writeError(w, err) {
			return
		}
	}
}

func getUserUpdatingHandler(updater UserUpdater) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		userID := uint(0)
		if err := parsePathValue(mux.Vars(req), pathVarUserID, &userID); writeError(w, err) {
			return
		}

		var user User
		if err := unmarshalReaderContent(req.Body, &user); writeError(w, err) {
			return
		}

		user.ID = userID
		if err := updater.UpdateUser(&user); writeError(w, err) {
			return
		}
	}
}

// =========
// /drawings
// =========

func getDrawingsListGettingHandler(getter DrawingsListGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		user := new(UserOpen)
		if err := readCtxValue(req.Context(), ctxKeyUser, user); writeError(w, err) {
			return
		}

		amount, err := getter.DrawingsAmount(user.ID)
		if writeError(w, err) {
			return
		}
		stat, err := readListStatData(req.URL.Query(), amount)
		if writeError(w, err) {
			return
		}

		respData := DrawingsListResponseData{Drawings: make([]*DrawingOpen, 0), ListStatData: *stat}
		if amount > 0 {
			respData.Drawings, err = getter.GetDrawingsList(user.ID, respData.Page, respData.PageLimit)
			if writeError(w, err) {
				return
			}
		}

		marshalAndWrite(w, &respData)
	}
}

func getDrawingCreatingHandler(creator DrawingCreator) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		user := new(UserOpen)
		if err := readCtxValue(req.Context(), ctxKeyUser, user); writeError(w, err) {
			return
		}

		var requestData DrawingPostPutRequestData
		if err := unmarshalReaderContent(req.Body, &requestData); writeError(w, err) {
			return
		}
		if requestData.Name == "" {
			http.Error(w, "Bad Request: Wrong JSON body", http.StatusBadRequest)
			return
		}
		drawing := Drawing{DrawingOpen: requestData.DrawingOpen, GGDrawing: *raster.NewEmptyGGDrawing()}
		drawing.Measures = requestData.Measures.ToFigureMeasures(drawing.Measures)

		if err := drawing.AddPoints(getPointsFromRequestPoint(requestData.Points...)...); writeError(w, err) {
			return
		}

		if err := creator.CreateDrawings(user.ID, &drawing); err != nil {
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
	drawing := new(Drawing)
	if err := readCtxValue(req.Context(), ctxKeyDrawing, drawing); writeError(w, err) {
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

func getDrawingDeletingHandler(remover DrawingRemover) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		user, drawingID := new(UserOpen), uint(0)
		if err := readCtxValue(req.Context(), ctxKeyUser, user); writeError(w, err) {
			return
		}
		if err := parsePathValue(mux.Vars(req), pathVarDrawingID, &drawingID); writeError(w, err) {
			return
		}

		if err := remover.RemoveDrawingOfUser(user.ID, drawingID); writeError(w, err) {
			return
		}
	}
}

// ====================
// /drawings/{id}/image
// ====================

func drawingImageHandler(w http.ResponseWriter, req *http.Request) {
	drawing := new(Drawing)
	if err := readCtxValue(req.Context(), ctxKeyDrawing, drawing); writeError(w, err) {
		return
	}

	var err error
	drawDescription := false
	if err := parseURLParamValue(req.URL.Query(), urlParamInfo, &drawDescription); err != nil && !errors.Is(err, ErrNotFound) && writeError(w, err) {
		return
	}

	image, err := drawing.Draw(drawDescription)
	if writeError(w, err) {
		return
	}

	w.Header().Set("Content-Type", "image/png")
	if err = png.Encode(w, image); writeError(w, err) {
		return
	}
}

// =====================
// /drawings/{id}/points
// =====================

func drawingPointsGettingHandler(w http.ResponseWriter, req *http.Request) {
	drawing := new(Drawing)
	if err := readCtxValue(req.Context(), ctxKeyDrawing, drawing); writeError(w, err) {
		return
	}

	precision, measure := 2, drawing.Measures.Length
	if err := readLengthMeasureAndPrecision(req.URL.Query(), &measure, &precision); writeError(w, err) {
		return
	}
	respData := DrawingPointsGettingResponseData{
		DrawingOpen: drawing.DrawingOpen,
		Points:      drawing.GetPointsWithParams(measure, precision),
		Measure:     value.NameOfLengthMeasure(measure),
	}

	marshalAndWrite(w, &respData)
}

func getDrawingPointsAddingHandler(updater DrawingUpdater) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		drawing := new(Drawing)
		if err := readCtxValue(req.Context(), ctxKeyDrawing, drawing); writeError(w, err) {
			return
		}

		var reqData PointsCalculatingWithMeasures
		if err := unmarshalReaderContent(req.Body, &reqData); writeError(w, err) {
			return
		}

		dmCopy := drawing.Measures
		drawing.Measures = reqData.Measures.ToFigureMeasures(drawing.Measures)

		if err := drawing.AddPoints(getPointsFromRequestPoint(reqData.Points...)...); writeError(w, err) {
			return
		}

		respData := DrawingPointsGettingResponseData{
			DrawingOpen: drawing.DrawingOpen,
			Points:      drawing.GetPointsWithParams(drawing.Measures.Length, 2),
			Measure:     reqData.Measures.Length,
		}

		drawing.Measures = dmCopy

		if err := updater.UpdateDrawing(drawing); writeError(w, err) {
			return
		}

		marshalAndWrite(w, &respData)
	}
}

// =========================
// /drawings/{id}/points/{n}
// =========================

func drawingPointGettingHandler(w http.ResponseWriter, req *http.Request) {
	pointIndex, drawing := 0, new(Drawing)
	if err := readCtxValue(req.Context(), ctxKeyDrawing, drawing); writeError(w, err) {
		return
	}
	if err := readCtxValue(req.Context(), ctxKeyPointIndex, &pointIndex); writeError(w, err) {
		return
	}

	precision, measure := 2, drawing.Measures.Length
	if err := readLengthMeasureAndPrecision(req.URL.Query(), &measure, &precision); writeError(w, err) {
		return
	}
	point := drawing.Points[pointIndex]
	marshalAndWrite(w, PointWithMeasure{
		X:       value.ConvertFromOneRound(measure, point.X, precision),
		Y:       value.ConvertFromOneRound(measure, point.Y, precision),
		Measure: value.NameOfLengthMeasure(measure),
	})
}

func getDrawingPointDeletingHandler(updater DrawingUpdater) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		pointIndex, drawing := 0, new(Drawing)
		if err := readCtxValue(req.Context(), ctxKeyDrawing, drawing); writeError(w, err) {
			return
		}
		if err := readCtxValue(req.Context(), ctxKeyPointIndex, &pointIndex); writeError(w, err) {
			return
		}

		drawing.Points = append(drawing.Points[:pointIndex], drawing.Points[pointIndex+1:]...)

		if err := updater.UpdateDrawing(drawing); writeError(w, err) {
			return
		}
	}
}

func getDrawingPointUpdatingHandler(updater DrawingUpdater) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		pointIndex, drawing := 0, new(Drawing)
		if err := readCtxValue(req.Context(), ctxKeyDrawing, drawing); writeError(w, err) {
			return
		}
		if err := readCtxValue(req.Context(), ctxKeyPointIndex, &pointIndex); writeError(w, err) {
			return
		}

		var pointWithMeasure PointCalculatingWithMeasures
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

		if err := updater.UpdateDrawing(drawing); writeError(w, err) {
			return
		}
	}
}
