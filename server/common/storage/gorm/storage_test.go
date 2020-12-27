package gorm

import (
	"errors"
	"fmt"
	"gorm.io/driver/sqlite"
	"os"
	"path"
	"testing"
	"time"

	"github.com/go-test/deep"
	"github.com/maxsid/goCeilings/drawing"
	"github.com/maxsid/goCeilings/drawing/raster"
	"github.com/maxsid/goCeilings/figure"
	"github.com/maxsid/goCeilings/server/api"
	"github.com/maxsid/goCeilings/server/common"
	"github.com/maxsid/goCeilings/value"
)

var drawings = map[int]*common.Drawing{
	1: {
		DrawingBasic: common.DrawingBasic{ID: 1, Name: "First"},
		GGDrawing: raster.GGDrawing{
			Polygon: figure.Polygon{Points: []*figure.Point{{X: 0, Y: 0}, {X: 0, Y: 1.25}, {X: 0.27, Y: 1.25},
				{X: 0.2701, Y: 1.71}, {X: 2.2201, Y: 1.6998}, {X: 2.25, Y: 0}}},
			Description: &drawing.Description{{"Material", "Satin"}, {"Address", "Lenin's St."}},
			Measures:    value.NewFigureMeasures(),
		},
	},
	2: {
		DrawingBasic: common.DrawingBasic{ID: 2, Name: "Second"},
		GGDrawing: raster.GGDrawing{
			Polygon: figure.Polygon{Points: []*figure.Point{{X: 0, Y: 0}, {X: 0, Y: 1.55}, {X: 0.725, Y: 1.55},
				{X: 0.725, Y: 1.675}, {X: 0.125, Y: 1.6751}, {X: 0.1253, Y: 5.9751}, {X: 3.4252, Y: 5.9999}, {X: 3.45, Y: 0}}},
			Description: &drawing.Description{{"colour", "red"}, {"client", "Ivan"}},
			Measures:    value.NewFigureMeasures(),
		},
	},
	3: {
		DrawingBasic: common.DrawingBasic{ID: 3, Name: "Third"},
		GGDrawing: raster.GGDrawing{
			Polygon: figure.Polygon{Points: []*figure.Point{
				{X: 0, Y: 0},
				{Calculator: &figure.DirectionCalculator{Direction: value.ConvertToOne(value.Degree, 90), Distance: 1.42}},
				{Calculator: &figure.AngleCalculator{Angle: value.ConvertToOne(value.Degree, 90), Distance: 4.23}},
				{Calculator: &figure.AngleCalculator{Angle: value.ConvertToOne(value.Degree, 90), Distance: 1.42}},
			}},
			Description: &drawing.Description{{"material", "m"}, {"client", "Sergey"}},
			Measures:    value.NewFigureMeasures(),
		},
	},
	4: {
		DrawingBasic: common.DrawingBasic{ID: 4, Name: "Fourth (empty)"},
		GGDrawing: raster.GGDrawing{
			Polygon:     figure.Polygon{Points: []*figure.Point{}},
			Description: &drawing.Description{{"address", "K. Marx St."}},
			Measures:    value.NewFigureMeasures(),
		},
	},
}

var users = map[int]*common.UserConfident{
	1: {UserBasic: common.UserBasic{ID: 1, Login: "maxim", Role: common.RoleAdmin}, Password: "password1"},
	2: {UserBasic: common.UserBasic{ID: 2, Login: "oleg", Role: common.RoleUser}, Password: "password2"},
	3: {UserBasic: common.UserBasic{ID: 3, Login: "elena", Role: common.RoleUser}, Password: "password3"},
}

var permissions = []*common.DrawingPermission{
	{User: &users[1].UserBasic, Drawing: &drawings[2].DrawingBasic, Owner: true},
	{User: &users[1].UserBasic, Drawing: &drawings[4].DrawingBasic, Get: true},

	{User: &users[2].UserBasic, Drawing: &drawings[1].DrawingBasic, Owner: true},
	{User: &users[2].UserBasic, Drawing: &drawings[3].DrawingBasic, Get: true, Change: true},

	{User: &users[3].UserBasic, Drawing: &drawings[3].DrawingBasic, Owner: true},
	{User: &users[3].UserBasic, Drawing: &drawings[4].DrawingBasic, Owner: true},
	{User: &users[3].UserBasic, Drawing: &drawings[1].DrawingBasic, Get: true, Change: true, Delete: true, Share: true},
	{User: &users[3].UserBasic, Drawing: &drawings[2].DrawingBasic, Get: true},
}

var (
	previousStorageFile string
	storage             *Storage
)

func createTempStorage() {
	deleteTempStorage()
	var err error
	previousStorageFile = path.Join(os.TempDir(), fmt.Sprintf("goCeiling-sqlite-test-%d", time.Now().UnixNano()))
	storage, err = NewDatabase(sqlite.Open(previousStorageFile))
	if err != nil {
		panic(err)
	}
	if err = createStorageData(storage); err != nil {
		panic(err)
	}
}

func deleteTempStorage() {
	if previousStorageFile == "" {
		return
	}
	if err := os.RemoveAll(previousStorageFile); err != nil {
		panic(err)
	}
	previousStorageFile = ""
}

func createStorageData(storage *Storage) error {
	for _, d := range drawings {
		if err := d.CalculatePoints(); err != nil {
			return err
		}
	}
	initTime := time.Now()
	storageUsers := make([]*userModel, len(users))
	for i, u := range users {
		i--
		storageUsers[i] = &userModel{}
		storageUsers[i].FromAPI(u)
		storageUsers[i].Password = getHexHash(u.Password, HashSalt)
		t := initTime.Add(time.Duration(i) * time.Second)
		storageUsers[i].CreatedAt, storageUsers[i].UpdatedAt = t, t
	}
	if err := storage.db.Create(&storageUsers).Error; err != nil {
		return err
	}

	storageDrawings := make([]*drawingModel, len(drawings))
	for i, d := range drawings {
		i--
		storageDrawings[i] = &drawingModel{}
		storageDrawings[i].FromAPI(d)
		t := initTime.Add(time.Duration(i) * 5 * time.Second)
		storageDrawings[i].CreatedAt, storageDrawings[i].UpdatedAt = t, t
	}
	if err := storage.db.Create(&storageDrawings).Error; err != nil {
		return err
	}

	storagePs := make([]*drawingPermissionModel, len(permissions))
	for i, p := range permissions {
		storagePs[i] = &drawingPermissionModel{}
		storagePs[i].FromAPI(p)
		var t time.Time
		if p.Owner {
			for _, d := range storageDrawings {
				if d.ID == p.Drawing.ID {
					t = d.CreatedAt
					break
				}
			}
		} else {
			t = initTime.Add(time.Hour*24 + (time.Duration(i)*10)*time.Second)
		}
		storagePs[i].CreatedAt, storagePs[i].UpdatedAt = t, t
	}
	if err := storage.db.Create(&storagePs).Error; err != nil {
		return err
	}
	return nil
}

func TestStorage_GetUser(t *testing.T) {
	createTempStorage()
	defer deleteTempStorage()

	type args struct {
		login string
		pass  string
	}
	tests := []struct {
		name    string
		args    args
		want    common.UserBasic
		wantErr bool
	}{
		{
			name: "OK 1",
			args: args{
				login: "maxim",
				pass:  "password1",
			},
			want: common.UserBasic{ID: 1, Login: "maxim", Role: common.RoleAdmin},
		},
		{
			name: "OK 2",
			args: args{
				login: "oleg",
				pass:  "password2",
			},
			want: common.UserBasic{ID: 2, Login: "oleg", Role: common.RoleUser},
		},
		{
			name: "Not found 1",
			args: args{
				login: "maxim2",
				pass:  "password1",
			},
			wantErr: true,
		},
		{
			name: "Not found 2",
			args: args{
				login: "maxim",
				pass:  "password2",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := storage.GetUser(tt.args.login, tt.args.pass)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr || err != nil {
				return
			}
			if diff := deep.Equal(got.UserBasic, tt.want); diff != nil {
				t.Errorf("GetUser() -> %v", diff)
			}
		})
	}
}

func TestStorage_GetUserByID(t *testing.T) {
	createTempStorage()
	defer deleteTempStorage()

	type args struct {
		id uint
	}
	tests := []struct {
		name    string
		args    args
		want    common.UserBasic
		wantErr bool
	}{
		{
			name: "OK 1",
			args: args{1},
			want: common.UserBasic{
				ID:    1,
				Login: "maxim",
				Role:  common.RoleAdmin,
			},
		},
		{
			name: "OK 3",
			args: args{3},
			want: common.UserBasic{
				ID:    3,
				Login: "elena",
				Role:  common.RoleUser,
			},
		},
		{
			name:    "Not found",
			args:    args{654},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := storage.GetUserByID(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetUserByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr || err == nil {
				return
			}
			if got == nil {
				t.Error("GetUserByID() got nil")
				return
			}
			if diff := deep.Equal(got.UserBasic, tt.want); diff != nil {
				t.Errorf("GetUserByID() -> %v", diff)
			}

		})
	}
}

func TestStorage_CreateUser(t *testing.T) {
	createTempStorage()
	defer deleteTempStorage()

	type args struct {
		user *common.UserConfident
	}
	tests := []struct {
		name    string
		args    args
		want    *common.UserConfident
		wantErr bool
	}{
		{
			name: "OK",
			args: args{&common.UserConfident{UserBasic: common.UserBasic{Login: "dmitry", Role: common.RoleUser}, Password: "pass4"}},
			want: &common.UserConfident{UserBasic: common.UserBasic{ID: 4, Login: "dmitry", Role: common.RoleUser}, Password: getHexHash("pass4", HashSalt)},
		},
		{
			name:    "UserConfident with this login already exist",
			args:    args{&common.UserConfident{UserBasic: common.UserBasic{Login: "maxim", Role: common.RoleUser}, Password: "pass4"}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := storage.CreateUsers(tt.args.user)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if tt.want.ID == 0 {
					t.Error("CreateUser() must change UserConfident.ID, now the field is zero")
					return
				}
				if _, err = storage.GetUserByID(tt.want.ID); err != nil {
					t.Errorf("GetUserByID() after CreateUser() has error %v", err)
				}
			}
		})
	}
}

func TestStorage_UpdateUser(t *testing.T) {
	createTempStorage()
	defer deleteTempStorage()

	type args struct {
		user *common.UserConfident
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "OK",
			args: args{
				&common.UserConfident{UserBasic: common.UserBasic{ID: 1, Login: "maxim2", Role: common.RoleUser}, Password: "password13"}},
		},
		{
			name:    "Not found",
			args:    args{&common.UserConfident{UserBasic: common.UserBasic{ID: 123, Login: "maxim2"}, Password: "password1"}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := storage.UpdateUser(tt.args.user); (err != nil) != tt.wantErr {
				t.Errorf("UpdateUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			} else if err == nil {
				user, err := storage.GetUserByID(tt.args.user.ID)
				if err != nil {
					t.Errorf("GetUserByID() after UpdateUser() got error = %v", err)
					return
				}
				if diff := deep.Equal(tt.args.user, user); diff != nil {
					t.Errorf("UpdateUser() -> %v", diff)
				}
			}
		})
	}
}

func TestStorage_RemoveUser(t *testing.T) {
	createTempStorage()
	defer deleteTempStorage()

	type args struct {
		id uint
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "OK",
			args: args{1},
		},
		{
			name:    "Not found",
			args:    args{132},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := storage.RemoveUser(tt.args.id); (err != nil) != tt.wantErr {
				t.Errorf("RemoveUser() error = %v, wantErr %v", err, tt.wantErr)
			} else if err == nil {
				if _, err = storage.GetUserByID(tt.args.id); err == nil {
					t.Errorf("RemoveUser() hasn't deleted the user id = %v", tt.args.id)
					return
				}
			}
		})
	}
}

func TestStorage_GetUsersList(t *testing.T) {
	createTempStorage()
	defer deleteTempStorage()

	type args struct {
		page      uint
		pageLimit uint
	}
	tests := []struct {
		name    string
		args    args
		want    []*common.UserBasic
		wantErr bool
	}{
		{
			name: "OK page 1 pageLimit 10",
			args: args{page: 1, pageLimit: 10},
			want: []*common.UserBasic{
				{ID: 1, Login: "maxim", Role: common.RoleAdmin},
				{ID: 2, Login: "oleg", Role: common.RoleUser},
				{ID: 3, Login: "elena", Role: common.RoleUser},
			},
		},
		{
			name: "OK page 2 pageLimit 1",
			args: args{page: 2, pageLimit: 1},
			want: []*common.UserBasic{
				{ID: 2, Login: "oleg", Role: common.RoleUser},
			},
		},
		{
			name: "OK page 2 pageLimit 0",
			args: args{page: 2, pageLimit: 0},
			want: []*common.UserBasic{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := storage.GetUsersList(tt.args.page, tt.args.pageLimit)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetUsersList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := deep.Equal(got, tt.want); diff != nil {
				t.Errorf("GetUsersList() -> %v", diff)
			}
		})
	}
}

func TestStorage_UsersAmount(t *testing.T) {
	createTempStorage()
	defer deleteTempStorage()

	tests := []struct {
		name    string
		want    uint
		wantErr bool
	}{
		{
			name: "OK",
			want: 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := storage.UsersAmount()
			if (err != nil) != tt.wantErr {
				t.Errorf("UsersAmount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("UsersAmount() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStorage_GetDrawing(t *testing.T) {
	createTempStorage()
	defer deleteTempStorage()

	type args struct {
		id uint
	}
	tests := []struct {
		name    string
		args    args
		want    *common.Drawing
		wantErr bool
	}{
		{
			name: "OK",
			args: args{id: 1},
			want: drawings[1],
		},
		{
			name:    "Not found",
			args:    args{id: 312},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := storage.GetDrawing(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDrawing() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if diff := deep.Equal(got, tt.want); diff != nil {
				t.Errorf("GetDrawing() -> %v", diff)
			}
		})
	}
}

func TestStorage_GetDrawingOfUser(t *testing.T) {
	createTempStorage()
	defer deleteTempStorage()

	type args struct {
		userID    uint
		drawingID uint
	}
	tests := []struct {
		name    string
		args    args
		want    *common.Drawing
		wantErr bool
	}{
		{
			name: "OK",
			args: args{userID: 1, drawingID: 2},
			want: drawings[2],
		},
		{
			name:    "UserConfident has not access",
			args:    args{userID: 1, drawingID: 1},
			wantErr: true,
		},
		{
			name:    "UserConfident doesn't exist",
			args:    args{userID: 44, drawingID: 1},
			wantErr: true,
		},
		{
			name:    "Drawing doesn't exist",
			args:    args{userID: 1, drawingID: 44},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := storage.getDrawingOfUser(tt.args.userID, tt.args.drawingID)
			if (err != nil) != tt.wantErr {
				t.Errorf("getDrawingOfUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := deep.Equal(got, tt.want); diff != nil {
				t.Errorf("getDrawingOfUser() -> %v", diff)
			}
		})
	}
}

func checkUpdateDrawing(t *testing.T, storage *Storage, drawing *common.Drawing, userID uint, wantErr bool) {
	var err error
	if userID == 0 {
		err = storage.UpdateDrawing(drawing)
	} else {
		err = storage.updateDrawingOfUser(userID, drawing)
	}
	if (err != nil) != wantErr {
		t.Errorf("error = %v, wantErr %v", err, wantErr)
		return
	}
	if !wantErr {
		got, err := storage.GetDrawing(drawing.ID)
		if err != nil {
			t.Errorf("GetDrawing() got error: %v", err)
			return
		}
		if diff := deep.Equal(drawing, got); diff != nil {
			t.Errorf("Got not equal drawings -> %v", diff)
		}
	}
}

func TestStorage_UpdateDrawing(t *testing.T) {
	createTempStorage()
	defer deleteTempStorage()

	type args struct {
		drawing *common.Drawing
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "OK",
			args: args{&common.Drawing{DrawingBasic: common.DrawingBasic{ID: 1, Name: "Updated Drawing"}, GGDrawing: drawings[2].GGDrawing}},
		},
		{
			name:    "Not found",
			args:    args{&common.Drawing{DrawingBasic: common.DrawingBasic{ID: 92, Name: "Not Updated Drawing"}, GGDrawing: drawings[2].GGDrawing}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkUpdateDrawing(t, storage, tt.args.drawing, 0, tt.wantErr)
		})
	}
}

func TestStorage_UpdateDrawingOfUser(t *testing.T) {
	createTempStorage()
	defer deleteTempStorage()

	type args struct {
		userID  uint
		drawing *common.Drawing
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "OK",
			args: args{
				userID:  1,
				drawing: &common.Drawing{DrawingBasic: common.DrawingBasic{ID: 2, Name: "Updated Drawing"}, GGDrawing: drawings[3].GGDrawing},
			},
		},
		{
			name: "UserConfident don't have access",
			args: args{
				userID:  2,
				drawing: &common.Drawing{DrawingBasic: common.DrawingBasic{ID: 9, Name: "Updated Drawing"}, GGDrawing: drawings[3].GGDrawing},
			},
			wantErr: true,
		},
		{
			name: "Not found drawing by ID",
			args: args{
				userID:  1,
				drawing: &common.Drawing{DrawingBasic: common.DrawingBasic{ID: 44, Name: "Updated Drawing"}, GGDrawing: drawings[3].GGDrawing},
			},
			wantErr: true,
		},
		{
			name: "Not found user by ID",
			args: args{
				userID:  44,
				drawing: &common.Drawing{DrawingBasic: common.DrawingBasic{ID: 9, Name: "Updated Drawing"}, GGDrawing: drawings[3].GGDrawing},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkUpdateDrawing(t, storage, tt.args.drawing, tt.args.userID, tt.wantErr)
		})
	}
}

func checkDrawingRemoving(t *testing.T, storage *Storage, drawingID, userID uint, wantErr bool) {
	if userID == 0 {
		if err := storage.RemoveDrawing(drawingID); (err != nil) != wantErr {
			t.Errorf("RemoveDrawing() error = %v, wantErr %v", err, wantErr)
			return
		}
	} else {
		if err := storage.removeDrawingOfUser(userID, drawingID); (err != nil) != wantErr {
			t.Errorf("removeDrawingOfUser() error = %v, wantErr %v", err, wantErr)
			return
		}
	}

	if !wantErr {
		if _, err := storage.GetDrawing(drawingID); !errors.Is(err, api.ErrDrawingNotFound) {
			t.Error("Drawing hasn't been deleted from storage")
		}
		// check count of drawings after removing
		count := int64(-1)
		err := storage.db.Find(&drawingPermissionModel{}, "drawing_id = ?", drawingID).Count(&count).Error
		if err != nil {
			t.Errorf("Getting a number of a number of the drawing relations error: %v", err)
		}
		if count != 0 {
			t.Errorf("Wrong a number of the drawing relations. Must be 0, got %v", count)
		}
	}
}

func TestStorage_RemoveDrawing(t *testing.T) {
	createTempStorage()
	defer deleteTempStorage()

	type args struct {
		id uint
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "OK",
			args: args{id: 1},
		},
		{
			name:    "Not found",
			args:    args{id: 132},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkDrawingRemoving(t, storage, tt.args.id, 0, tt.wantErr)
		})
	}
}

func TestStorage_RemoveDrawingOfUser(t *testing.T) {
	createTempStorage()
	defer deleteTempStorage()

	type args struct {
		userID    uint
		drawingID uint
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "OK",
			args: args{userID: 2, drawingID: 1},
		},
		{
			name:    "UserConfident hasn't access",
			args:    args{userID: 1, drawingID: 1},
			wantErr: true,
		},
		{
			name:    "UserConfident doesn't exist with this UserID",
			args:    args{userID: 123, drawingID: 1},
			wantErr: true,
		},
		{
			name:    "Drawing doesn't exist with this DrawingID",
			args:    args{userID: 1, drawingID: 123},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkDrawingRemoving(t, storage, tt.args.drawingID, tt.args.userID, tt.wantErr)
		})
	}
}

func TestStorage_GetDrawingsList(t *testing.T) {
	createTempStorage()
	defer deleteTempStorage()

	type args struct {
		userID    uint
		page      uint
		pageLimit uint
	}
	tests := []struct {
		name    string
		args    args
		want    []*common.DrawingBasic
		wantErr bool
	}{
		{
			name: "OK 1: page 1, pageLimit 30",
			args: args{userID: 1, page: 1, pageLimit: 30},
			want: []*common.DrawingBasic{
				&drawings[4].DrawingBasic,
				&drawings[2].DrawingBasic,
			},
		},
		{
			name: "OK 2: page 2, pageLimit 2",
			args: args{userID: 3, page: 2, pageLimit: 2},
			want: []*common.DrawingBasic{
				&drawings[4].DrawingBasic,
				&drawings[3].DrawingBasic,
			},
		},
		{
			name: "OK 3: page 2, pageLimit 0",
			args: args{userID: 1, page: 2, pageLimit: 0},
			want: []*common.DrawingBasic{},
		},
		{
			name:    "Not found user",
			args:    args{userID: 123, page: 1, pageLimit: 10},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := storage.GetDrawingsList(tt.args.userID, tt.args.page, tt.args.pageLimit)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDrawingsList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := deep.Equal(got, tt.want); diff != nil {
				t.Errorf("GetDrawingsList() -> %v", diff)
			}
		})
	}
}

func TestStorage_DrawingsAmount(t *testing.T) {
	createTempStorage()
	defer deleteTempStorage()

	type args struct {
		userID uint
	}
	tests := []struct {
		name    string
		args    args
		want    uint
		wantErr bool
	}{
		{
			name: "OK",
			args: args{userID: 1},
			want: 2,
		},
		{
			name: "Not found",
			args: args{userID: 9},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := storage.DrawingsAmount(tt.args.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("DrawingsAmount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("DrawingsAmount() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStorage_CreateDrawings(t *testing.T) {
	createTempStorage()
	defer deleteTempStorage()

	type args struct {
		userID  uint
		drawing *common.Drawing
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "OK",
			args: args{userID: 1, drawing: &common.Drawing{DrawingBasic: common.DrawingBasic{ID: 10, Name: "Drawing 10"}, GGDrawing: drawings[2].GGDrawing}},
		},
		{
			name:    "Not found user",
			args:    args{userID: 123, drawing: &common.Drawing{DrawingBasic: common.DrawingBasic{ID: 10, Name: "Drawing 10"}, GGDrawing: drawings[2].GGDrawing}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := storage.CreateDrawings(tt.args.userID, tt.args.drawing); (err != nil) != tt.wantErr {
				t.Errorf("CreateDrawings() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				d, err := storage.GetDrawing(tt.args.drawing.DrawingBasic.ID)
				if err != nil {
					t.Errorf("GetDrawing() error: %v", err)
					return
				}
				if diff := deep.Equal(tt.args.drawing, d); diff != nil {
					t.Errorf("Drawings aren't equal -> %v", diff)
				}
			}
		})
	}
}

func TestStorage_CreateDrawingPermission(t *testing.T) {
	createTempStorage()
	defer deleteTempStorage()

	type args struct {
		permission *common.DrawingPermission
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "OK",
			args: args{permission: &common.DrawingPermission{
				User:    &common.UserBasic{ID: 1, Login: "maxim", Role: common.RoleAdmin},
				Drawing: &common.DrawingBasic{ID: 1},
				Get:     true,
				Change:  false,
				Delete:  false,
				Share:   false,
				Owner:   false,
			}},
		},
		{
			name: "Exist",
			args: args{permission: &common.DrawingPermission{
				User:    &common.UserBasic{ID: 1, Login: "maxim", Role: common.RoleAdmin},
				Drawing: &common.DrawingBasic{ID: 2},
				Get:     true,
				Change:  false,
				Delete:  false,
				Share:   false,
				Owner:   false,
			}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := storage.CreateDrawingPermission(tt.args.permission); (err != nil) != tt.wantErr {
				t.Errorf("CreateDrawingPermission() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStorage_GetDrawingPermission(t *testing.T) {
	createTempStorage()
	defer deleteTempStorage()

	type args struct {
		userID    uint
		drawingID uint
	}
	tests := []struct {
		name    string
		args    args
		want    *common.DrawingPermission
		wantErr bool
	}{
		{
			name: "OK",
			args: args{userID: 1, drawingID: 2},
			want: &common.DrawingPermission{
				User:    &common.UserBasic{ID: 1, Login: "maxim", Role: common.RoleAdmin},
				Drawing: &common.DrawingBasic{ID: 2, Name: "Second"},
				Owner:   true,
			},
		},
		{
			name:    "Not found",
			args:    args{userID: 1, drawingID: 1},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := storage.GetDrawingPermission(tt.args.userID, tt.args.drawingID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDrawingPermission() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := deep.Equal(got, tt.want); diff != nil {
				t.Errorf("GetDrawingPermission() -> %v", diff)
			}
		})
	}
}

func TestStorage_UpdateDrawingPermission(t *testing.T) {
	createTempStorage()
	defer deleteTempStorage()

	type args struct {
		permission *common.DrawingPermission
	}
	tests := []struct {
		name       string
		args       args
		wantErr    bool
		wantRemove bool
	}{
		{
			name: "OK",
			args: args{permission: &common.DrawingPermission{User: &users[3].UserBasic, Drawing: &drawings[2].DrawingBasic, Get: true, Change: true, Delete: true}},
		},
		{
			name:    "Owner cannot be changed",
			args:    args{permission: &common.DrawingPermission{User: &users[3].UserBasic, Drawing: &drawings[3].DrawingBasic, Owner: false}},
			wantErr: true,
		},
		{
			name:    "Not found",
			args:    args{permission: &common.DrawingPermission{User: &users[2].UserBasic, Drawing: &drawings[2].DrawingBasic, Get: true}},
			wantErr: true,
		},
		{
			name: "Off owner in update",
			args: args{permission: &common.DrawingPermission{User: &users[3].UserBasic, Drawing: &drawings[2].DrawingBasic, Get: true, Change: true, Delete: true, Owner: true}},
		},
		{
			name:       "Auto delete",
			args:       args{permission: &common.DrawingPermission{User: &users[3].UserBasic, Drawing: &drawings[1].DrawingBasic}}, // full false
			wantRemove: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := storage.UpdateDrawingPermission(tt.args.permission); (err != nil) != tt.wantErr {
				t.Errorf("UpdateDrawingPermission() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				p, err := storage.GetDrawingPermission(tt.args.permission.User.ID, tt.args.permission.Drawing.ID)
				if err != nil && errors.Is(err, api.ErrNotFound) != tt.wantRemove {
					t.Errorf("GetDrawingPermission() error = %v", err)
					return
				} else if err != nil {
					return
				}
				tt.args.permission.Owner = false
				if diff := deep.Equal(p, tt.args.permission); diff != nil {
					t.Errorf("Got %v", diff)
				}
			}
		})
	}
}

func TestStorage_RemoveDrawingPermission(t *testing.T) {
	createTempStorage()
	defer deleteTempStorage()

	type args struct {
		userID    uint
		drawingID uint
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "OK",
			args: args{userID: 1, drawingID: 4},
		},
		{
			name:    "Not found",
			args:    args{userID: 2, drawingID: 2},
			wantErr: true,
		},
		{
			name:    "Owner cannot be removed",
			args:    args{userID: 3, drawingID: 3},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := storage.RemoveDrawingPermission(tt.args.userID, tt.args.drawingID); (err != nil) != tt.wantErr {
				t.Errorf("RemoveDrawingPermission() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				_, err := storage.GetDrawingPermission(tt.args.userID, tt.args.drawingID)
				if err != nil && errors.Is(err, api.ErrNotFound) {
					return
				}
				t.Errorf("GetDrawingPermission() must be ErrNotFound, got %v", err)
			}
		})
	}
}

func TestStorage_GetDrawingsPermissionsOfDrawing(t *testing.T) {
	createTempStorage()
	defer deleteTempStorage()

	type args struct {
		drawingID uint
	}
	tests := []struct {
		name    string
		args    args
		want    []*common.DrawingPermission
		wantErr bool
	}{
		{
			name: "OK",
			args: args{drawingID: 1},
			want: []*common.DrawingPermission{
				permissions[6],
				permissions[2],
			},
		},
		{
			name: "Not found",
			args: args{drawingID: 10},
			want: []*common.DrawingPermission{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := storage.GetDrawingsPermissionsOfDrawing(tt.args.drawingID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDrawingsPermissionsOfDrawing() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := deep.Equal(got, tt.want); diff != nil {
				t.Errorf("GetDrawingsPermissionsOfDrawing() -> %v", diff)
			}
		})
	}
}

func TestStorage_GetDrawingsPermissionsOfUser(t *testing.T) {
	createTempStorage()
	defer deleteTempStorage()

	type args struct {
		userID uint
	}
	tests := []struct {
		name    string
		args    args
		want    []*common.DrawingPermission
		wantErr bool
	}{
		{
			name: "OK",
			args: args{userID: 3},
			want: []*common.DrawingPermission{
				permissions[7],
				permissions[6],
				permissions[5],
				permissions[4],
			},
		},
		{
			name: "Not found",
			args: args{userID: 30},
			want: []*common.DrawingPermission{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := storage.GetDrawingsPermissionsOfUser(tt.args.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDrawingsPermissionsOfUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := deep.Equal(got, tt.want); diff != nil {
				t.Errorf("GetDrawingsPermissionsOfUser() -> %v", diff)
			}
		})
	}
}

func TestStorage_CreateAdmin(t *testing.T) {
	createTempStorage()
	defer deleteTempStorage()
	_ = storage.db.Delete(&userModel{}, "1 = 1")

	type args struct {
		force bool
	}
	tests := []struct {
		name         string
		args         args
		wantIncrease bool
		wantErr      bool
	}{
		{
			name:         "First run",
			args:         args{false},
			wantIncrease: true,
		},
		{
			name:         "Found admins",
			args:         args{false},
			wantIncrease: false,
		},
		{
			name:         "Force",
			args:         args{true},
			wantIncrease: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := storage.UsersAmount()
			if err := storage.CreateAdmin(tt.args.force); (err != nil) != tt.wantErr {
				t.Errorf("CreateAdmin() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			nc, _ := storage.UsersAmount()
			if tt.wantIncrease != (nc == c+1) {
				t.Errorf("wantIncrease = %t. Users number is %d, was %d", tt.wantIncrease, nc, c)
			}
		})
	}
}
