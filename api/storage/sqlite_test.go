package storage

import (
	"errors"
	"github.com/go-test/deep"
	"github.com/maxsid/goCeilings/api"
	"github.com/maxsid/goCeilings/drawing/raster"
	"github.com/maxsid/goCeilings/figure"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

var (
	drawing1, _ = raster.NewGGDrawingWithPoints([]*figure.Point{{X: 0, Y: 0}, {X: 0, Y: 1.25}, {X: 0.27, Y: 1.25},
		{X: 0.2701, Y: 1.71}, {X: 2.2201, Y: 1.6998}, {X: 2.25, Y: 0}}...)
	drawing2, _ = raster.NewGGDrawingWithPoints([]*figure.Point{{X: 0, Y: 0}, {X: 0, Y: 1.55}, {X: 0.725, Y: 1.55},
		{X: 0.725, Y: 1.675}, {X: 0.125, Y: 1.6751}, {X: 0.1253, Y: 5.9751}, {X: 3.4252, Y: 5.9999}, {X: 3.45, Y: 0}}...)
)

type deleteStorageFunc func()

func createTempStorage() (*Storage, deleteStorageFunc, error) {
	dirName, err := ioutil.TempDir(os.TempDir(), "goCeiling-sqlite-test-*")
	if err != nil {
		return nil, nil, err
	}
	storage, err := createStorage(path.Join(dirName, "test.db"))
	if err != nil {
		return nil, nil, err
	}
	return storage, getDeleterOfDirectory(dirName), nil
}

func createStorage(filename string) (storage *Storage, err error) {
	storage, err = NewSQLiteStorage(filename)
	if err != nil {
		return nil, err
	}

	users := []*api.User{
		{UserOpen: api.UserOpen{ID: 1, Login: "maxim", Permission: api.AdminPermission}, Password: "password1"},
		{UserOpen: api.UserOpen{ID: 2, Login: "oleg", Permission: api.UserPermission}, Password: "password2"},
		{UserOpen: api.UserOpen{ID: 3, Login: "elena", Permission: api.UserPermission}, Password: "password3"},
	}
	if err := storage.CreateUsers(users...); err != nil {
		return nil, err
	}

	drawings := map[*api.User][]*api.Drawing{
		users[0]: {
			{DrawingOpen: api.DrawingOpen{ID: 2, Name: "Drawing 2"}, GGDrawing: *drawing2},
			{DrawingOpen: api.DrawingOpen{ID: 6, Name: "Drawing 6"}, GGDrawing: *drawing1},
			{DrawingOpen: api.DrawingOpen{ID: 8, Name: "Drawing 8"}, GGDrawing: *drawing1},
			{DrawingOpen: api.DrawingOpen{ID: 9, Name: "Drawing 9"}, GGDrawing: *drawing2},
		},
		users[1]: {
			{DrawingOpen: api.DrawingOpen{ID: 1, Name: "Drawing 1"}, GGDrawing: *drawing1},
		},
		users[2]: {
			{DrawingOpen: api.DrawingOpen{ID: 3, Name: "Drawing 3"}, GGDrawing: *drawing1},
			{DrawingOpen: api.DrawingOpen{ID: 4, Name: "Drawing 4"}, GGDrawing: *drawing2},
			{DrawingOpen: api.DrawingOpen{ID: 5, Name: "Drawing 5"}, GGDrawing: *drawing1},
			{DrawingOpen: api.DrawingOpen{ID: 7, Name: "Drawing 7"}, GGDrawing: *drawing2},
		},
	}

	for user, ds := range drawings {
		if err := storage.CreateDrawings(user.ID, ds...); err != nil {
			return nil, err
		}
	}
	return
}

func getDeleterOfDirectory(dirName string) deleteStorageFunc {
	return func() {
		_ = os.RemoveAll(dirName)
	}
}

func TestNewSQLiteDatabase(t *testing.T) {
	fileName := path.Join(os.TempDir(), "test.db")
	_ = os.Remove(fileName)
	_, err := createStorage(fileName)
	if err != nil {
		t.Error(err)
	}
}

func TestStorage_GetUser(t *testing.T) {
	storage, del, err := createTempStorage()
	if err != nil {
		t.Error(err)
		return
	}
	defer del()

	type args struct {
		login string
		pass  string
	}
	tests := []struct {
		name    string
		args    args
		want    api.UserOpen
		wantErr bool
	}{
		{
			name: "OK 1",
			args: args{
				login: "maxim",
				pass:  "password1",
			},
			want: api.UserOpen{ID: 1, Login: "maxim", Permission: api.AdminPermission},
		},
		{
			name: "OK 2",
			args: args{
				login: "oleg",
				pass:  "password2",
			},
			want: api.UserOpen{ID: 2, Login: "oleg", Permission: api.UserPermission},
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
			if diff := deep.Equal(got.UserOpen, tt.want); diff != nil {
				t.Errorf("GetUser() -> %v", diff)
			}
		})
	}
}

func TestStorage_GetUserByID(t *testing.T) {
	storage, del, err := createTempStorage()
	if err != nil {
		t.Error(err)
		return
	}
	defer del()

	type args struct {
		id uint
	}
	tests := []struct {
		name    string
		args    args
		want    api.UserOpen
		wantErr bool
	}{
		{
			name: "OK 1",
			args: args{1},
			want: api.UserOpen{
				ID:         1,
				Login:      "maxim",
				Permission: api.AdminPermission,
			},
		},
		{
			name: "OK 3",
			args: args{3},
			want: api.UserOpen{
				ID:         3,
				Login:      "elena",
				Permission: api.UserPermission,
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
			if diff := deep.Equal(got.UserOpen, tt.want); diff != nil {
				t.Errorf("GetUserByID() -> %v", diff)
			}

		})
	}
}

func TestStorage_CreateUser(t *testing.T) {
	storage, del, err := createTempStorage()
	if err != nil {
		t.Error(err)
		return
	}
	defer del()

	type args struct {
		user *api.User
	}
	tests := []struct {
		name    string
		args    args
		want    *api.User
		wantErr bool
	}{
		{
			name: "OK",
			args: args{&api.User{UserOpen: api.UserOpen{Login: "dmitry", Permission: api.UserPermission}, Password: "pass4"}},
			want: &api.User{UserOpen: api.UserOpen{ID: 4, Login: "dmitry", Permission: api.UserPermission}, Password: getHexHash("pass4", HashSalt)},
		},
		{
			name:    "User with this login already exist",
			args:    args{&api.User{UserOpen: api.UserOpen{Login: "maxim", Permission: api.UserPermission}, Password: "pass4"}},
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
					t.Error("CreateUser() must change User.ID, now the field is zero")
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
	storage, del, err := createTempStorage()
	if err != nil {
		t.Error(err)
		return
	}
	defer del()

	type args struct {
		user *api.User
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "OK",
			args: args{
				&api.User{UserOpen: api.UserOpen{ID: 1, Login: "maxim2", Permission: api.UserPermission}, Password: "password13"}},
		},
		{
			name:    "Not found",
			args:    args{&api.User{UserOpen: api.UserOpen{ID: 123, Login: "maxim2"}, Password: "password1"}},
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
	storage, del, err := createTempStorage()
	if err != nil {
		t.Error(err)
		return
	}
	defer del()

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
	storage, del, err := createTempStorage()
	if err != nil {
		t.Error(err)
		return
	}
	defer del()

	type args struct {
		page      uint
		pageLimit uint
	}
	tests := []struct {
		name    string
		args    args
		want    []*api.UserOpen
		wantErr bool
	}{
		{
			name: "OK page 1 pageLimit 10",
			args: args{page: 1, pageLimit: 10},
			want: []*api.UserOpen{
				{ID: 1, Login: "maxim", Permission: api.AdminPermission},
				{ID: 2, Login: "oleg", Permission: api.UserPermission},
				{ID: 3, Login: "elena", Permission: api.UserPermission},
			},
		},
		{
			name: "OK page 2 pageLimit 1",
			args: args{page: 2, pageLimit: 1},
			want: []*api.UserOpen{
				{ID: 2, Login: "oleg", Permission: api.UserPermission},
			},
		},
		{
			name: "OK page 2 pageLimit 0",
			args: args{page: 2, pageLimit: 0},
			want: []*api.UserOpen{},
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
	storage, del, err := createTempStorage()
	if err != nil {
		t.Error(err)
		return
	}
	defer del()

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
	storage, del, err := createTempStorage()
	if err != nil {
		t.Error(err)
		return
	}
	defer del()

	type args struct {
		id uint
	}
	tests := []struct {
		name    string
		args    args
		want    *api.Drawing
		wantErr bool
	}{
		{
			name: "OK",
			args: args{id: 1},
			want: &api.Drawing{DrawingOpen: api.DrawingOpen{ID: 1, Name: "Drawing 1"}, GGDrawing: *drawing1},
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
	storage, del, err := createTempStorage()
	if err != nil {
		t.Error(err)
		return
	}
	defer del()

	type args struct {
		userID    uint
		drawingID uint
	}
	tests := []struct {
		name    string
		args    args
		want    *api.Drawing
		wantErr bool
	}{
		{
			name: "OK",
			args: args{userID: 1, drawingID: 2},
			want: &api.Drawing{DrawingOpen: api.DrawingOpen{ID: 2, Name: "Drawing 2"}, GGDrawing: *drawing2},
		},
		{
			name:    "User has not access",
			args:    args{userID: 1, drawingID: 1},
			wantErr: true,
		},
		{
			name:    "User doesn't exist",
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
			got, err := storage.GetDrawingOfUser(tt.args.userID, tt.args.drawingID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDrawingOfUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := deep.Equal(got, tt.want); diff != nil {
				t.Errorf("GetDrawingOfUser() -> %v", diff)
			}
		})
	}
}

func checkUpdateDrawing(t *testing.T, storage *Storage, drawing *api.Drawing, userID uint, wantErr bool) {
	var err error
	if userID == 0 {
		err = storage.UpdateDrawing(drawing)
	} else {
		err = storage.UpdateDrawingOfUser(userID, drawing)
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
	storage, del, err := createTempStorage()
	if err != nil {
		t.Error(err)
		return
	}
	defer del()

	type args struct {
		drawing *api.Drawing
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "OK",
			args: args{&api.Drawing{DrawingOpen: api.DrawingOpen{ID: 9, Name: "Updated Drawing"}, GGDrawing: *drawing1}},
		},
		{
			name:    "Not found",
			args:    args{&api.Drawing{DrawingOpen: api.DrawingOpen{ID: 92, Name: "Not Updated Drawing"}, GGDrawing: *drawing1}},
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
	storage, del, err := createTempStorage()
	if err != nil {
		t.Error(err)
		return
	}
	defer del()

	type args struct {
		userID  uint
		drawing *api.Drawing
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
				drawing: &api.Drawing{DrawingOpen: api.DrawingOpen{ID: 9, Name: "Updated Drawing"}, GGDrawing: *drawing1},
			},
		},
		{
			name: "User don't have access",
			args: args{
				userID:  2,
				drawing: &api.Drawing{DrawingOpen: api.DrawingOpen{ID: 9, Name: "Updated Drawing"}, GGDrawing: *drawing1},
			},
			wantErr: true,
		},
		{
			name: "Not found drawing by ID",
			args: args{
				userID:  1,
				drawing: &api.Drawing{DrawingOpen: api.DrawingOpen{ID: 44, Name: "Updated Drawing"}, GGDrawing: *drawing1},
			},
			wantErr: true,
		},
		{
			name: "Not found user by ID",
			args: args{
				userID:  44,
				drawing: &api.Drawing{DrawingOpen: api.DrawingOpen{ID: 9, Name: "Updated Drawing"}, GGDrawing: *drawing1},
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
		if err := storage.RemoveDrawingOfUser(userID, drawingID); (err != nil) != wantErr {
			t.Errorf("RemoveDrawingOfUser() error = %v, wantErr %v", err, wantErr)
			return
		}
	}

	if !wantErr {
		if _, err := storage.GetDrawing(drawingID); !errors.Is(err, api.ErrDrawingNotFound) {
			t.Error("Drawing hasn't been deleted from storage")
		}
		// check count of drawings after removing
		count := int64(-1)
		err := storage.db.Find(&UserDrawingRelation{}, "drawing_id = ?", drawingID).Count(&count).Error
		if err != nil {
			t.Errorf("Getting a number of a number of the drawing relations error: %v", err)
		}
		if count != 0 {
			t.Errorf("Wrong a number of the drawing relations. Must be 0, got %v", count)
		}
	}
}

func TestStorage_RemoveDrawing(t *testing.T) {
	storage, del, err := createTempStorage()
	if err != nil {
		t.Error(err)
		return
	}
	defer del()

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
	storage, del, err := createTempStorage()
	if err != nil {
		t.Error(err)
		return
	}
	defer del()

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
			name:    "User hasn't access",
			args:    args{userID: 1, drawingID: 1},
			wantErr: true,
		},
		{
			name:    "User doesn't exist with this UserID",
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
	storage, del, err := createTempStorage()
	if err != nil {
		t.Error(err)
		return
	}
	defer del()

	type args struct {
		userID    uint
		page      uint
		pageLimit uint
	}
	tests := []struct {
		name    string
		args    args
		want    []*api.DrawingOpen
		wantErr bool
	}{
		{
			name: "OK 1: page 1, pageLimit 30",
			args: args{userID: 1, page: 1, pageLimit: 30},
			want: []*api.DrawingOpen{
				{ID: 2, Name: "Drawing 2"},
				{ID: 6, Name: "Drawing 6"},
				{ID: 8, Name: "Drawing 8"},
				{ID: 9, Name: "Drawing 9"},
			},
		},
		{
			name: "OK 2: page 2, pageLimit 2",
			args: args{userID: 1, page: 2, pageLimit: 2},
			want: []*api.DrawingOpen{
				{ID: 8, Name: "Drawing 8"},
				{ID: 9, Name: "Drawing 9"},
			},
		},
		{
			name: "OK 3: page 2, pageLimit 0",
			args: args{userID: 1, page: 2, pageLimit: 0},
			want: []*api.DrawingOpen{},
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
	storage, del, err := createTempStorage()
	if err != nil {
		t.Error(err)
		return
	}
	defer del()

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
			want: 4,
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
	storage, del, err := createTempStorage()
	if err != nil {
		t.Error(err)
		return
	}
	defer del()

	type args struct {
		userID  uint
		drawing *api.Drawing
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "OK",
			args: args{userID: 1, drawing: &api.Drawing{DrawingOpen: api.DrawingOpen{ID: 10, Name: "Drawing 10"}, GGDrawing: *drawing1}},
		},
		{
			name:    "Not found user",
			args:    args{userID: 123, drawing: &api.Drawing{DrawingOpen: api.DrawingOpen{ID: 10, Name: "Drawing 10"}, GGDrawing: *drawing1}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := storage.CreateDrawings(tt.args.userID, tt.args.drawing); (err != nil) != tt.wantErr {
				t.Errorf("CreateDrawings() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				d, err := storage.GetDrawing(tt.args.drawing.DrawingOpen.ID)
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
