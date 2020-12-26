package storage

import (
	"github.com/maxsid/goCeilings/api"
	"github.com/maxsid/goCeilings/api/storage/generator"
	"github.com/maxsid/goCeilings/drawing/raster"
	"math/rand"
	"testing"
)

func TestUserStorage_checkPermission(t *testing.T) {
	createTempStorage()
	defer deleteTempStorage()

	type fields struct {
		Storage *Storage
		user    *api.UserBasic
	}
	type args struct {
		drawingID uint
		f         func(p *api.DrawingPermission) bool
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "allowed for admin",
			fields: fields{Storage: storage, user: &users[1].UserBasic},
			args:   args{1, func(p *api.DrawingPermission) bool { return true }},
		},
		{
			name:    "Not allowed for non-admin",
			fields:  fields{Storage: storage, user: &users[2].UserBasic},
			args:    args{4, func(p *api.DrawingPermission) bool { return true }},
			wantErr: true,
		},
		{
			name:    "Function not allows",
			fields:  fields{Storage: storage, user: &users[2].UserBasic},
			args:    args{1, func(p *api.DrawingPermission) bool { return false }},
			wantErr: true,
		},
		{
			name:   "Function allows",
			fields: fields{Storage: storage, user: &users[2].UserBasic},
			args:   args{1, func(p *api.DrawingPermission) bool { return true }},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &UserStorage{
				Storage: tt.fields.Storage,
				user:    tt.fields.user,
			}
			if err := u.checkPermission(tt.args.drawingID, tt.args.f); (err != nil) != tt.wantErr {
				t.Errorf("checkPermission() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUserStorage_GetUserStorage(t *testing.T) {
	createTempStorage()
	defer deleteTempStorage()

	type fields struct {
		Storage *Storage
		user    *api.UserBasic
	}
	type args struct {
		user *api.UserBasic
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "allowed for admin",
			fields: fields{Storage: storage, user: &users[1].UserBasic},
			args:   args{user: &users[2].UserBasic},
		},
		{
			name:    "Not allowed for non-admin",
			fields:  fields{Storage: storage, user: &users[3].UserBasic},
			args:    args{user: &users[2].UserBasic},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &UserStorage{
				Storage: tt.fields.Storage,
				user:    tt.fields.user,
			}
			if _, err := u.GetUserStorage(tt.args.user); (err != nil) != tt.wantErr {
				t.Errorf("GetUserStorage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestUserStorage_GetUser(t *testing.T) {
	createTempStorage()
	defer deleteTempStorage()

	type fields struct {
		Storage *Storage
		user    *api.UserBasic
	}
	type args struct {
		in0 string
		in1 string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:    "Not allowed for all",
			fields:  fields{Storage: storage, user: &users[1].UserBasic},
			args:    args{"maxim", "password1"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &UserStorage{
				Storage: tt.fields.Storage,
				user:    tt.fields.user,
			}
			if _, err := u.GetUser(tt.args.in0, tt.args.in1); (err != nil) != tt.wantErr {
				t.Errorf("GetUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestUserStorage_GetUserByID(t *testing.T) {
	createTempStorage()
	defer deleteTempStorage()

	type fields struct {
		Storage *Storage
		user    *api.UserBasic
	}
	type args struct {
		id uint
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "allowed for admin",
			fields: fields{Storage: storage, user: &users[1].UserBasic},
			args:   args{id: 2},
		},
		{
			name:   "allowed for a user with the same ID",
			fields: fields{Storage: storage, user: &users[2].UserBasic},
			args:   args{id: 2},
		},
		{
			name:    "Not allowed for non-admin",
			fields:  fields{Storage: storage, user: &users[3].UserBasic},
			args:    args{id: 2},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &UserStorage{
				Storage: tt.fields.Storage,
				user:    tt.fields.user,
			}
			_, err := u.GetUserByID(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetUserByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestUserStorage_CreateUsers(t *testing.T) {
	createTempStorage()
	defer deleteTempStorage()

	type fields struct {
		Storage *Storage
		user    *api.UserBasic
	}
	type args struct {
		users []*api.UserConfident
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "allowed for admin",
			fields: fields{Storage: storage, user: &users[1].UserBasic},
			args:   args{users: []*api.UserConfident{{UserBasic: api.UserBasic{Login: "maxim2", Role: api.RoleUser}, Password: "pass"}}},
		},
		{
			name:    "not allowed for non-admin",
			fields:  fields{Storage: storage, user: &users[3].UserBasic},
			args:    args{users: []*api.UserConfident{{UserBasic: api.UserBasic{Login: "maxim3", Role: api.RoleUser}, Password: "pass"}}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &UserStorage{
				Storage: tt.fields.Storage,
				user:    tt.fields.user,
			}
			if err := u.CreateUsers(tt.args.users...); (err != nil) != tt.wantErr {
				t.Errorf("CreateUsers() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUserStorage_UpdateUser(t *testing.T) {
	createTempStorage()
	defer deleteTempStorage()

	type fields struct {
		Storage *Storage
		user    *api.UserBasic
	}
	type args struct {
		user *api.UserConfident
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "allowed for admin",
			fields: fields{Storage: storage, user: &users[1].UserBasic},
			args:   args{&api.UserConfident{UserBasic: api.UserBasic{ID: 2, Login: "oleg4", Role: api.RoleUser}, Password: "password22"}},
		},
		{
			name:   "allowed for a user with the same ID",
			fields: fields{Storage: storage, user: &users[2].UserBasic},
			args:   args{&api.UserConfident{UserBasic: api.UserBasic{ID: 2, Login: "oleg8", Role: api.RoleUser}, Password: "password22"}},
		},
		{
			name:    "not allowed for non-admin",
			fields:  fields{Storage: storage, user: &users[3].UserBasic},
			args:    args{&api.UserConfident{UserBasic: api.UserBasic{ID: 2, Login: "oleg5", Role: api.RoleUser}, Password: "password22"}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &UserStorage{
				Storage: tt.fields.Storage,
				user:    tt.fields.user,
			}
			if err := u.UpdateUser(tt.args.user); (err != nil) != tt.wantErr {
				t.Errorf("UpdateUser() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUserStorage_RemoveUser(t *testing.T) {
	createTempStorage()
	defer deleteTempStorage()

	type fields struct {
		Storage *Storage
		user    *api.UserBasic
	}
	type args struct {
		id uint
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "allowed for admin",
			fields: fields{Storage: storage, user: &users[1].UserBasic},
			args:   args{id: 2},
		},
		{
			name:    "not allowed for non-admin",
			fields:  fields{Storage: storage, user: &users[3].UserBasic},
			args:    args{id: 1},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &UserStorage{
				Storage: tt.fields.Storage,
				user:    tt.fields.user,
			}
			if err := u.RemoveUser(tt.args.id); (err != nil) != tt.wantErr {
				t.Errorf("RemoveUser() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUserStorage_GetUsersList(t *testing.T) {
	createTempStorage()
	defer deleteTempStorage()

	type fields struct {
		Storage *Storage
		user    *api.UserBasic
	}
	type args struct {
		page      uint
		pageLimit uint
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "allowed for admin",
			fields: fields{Storage: storage, user: &users[1].UserBasic},
			args:   args{page: 1, pageLimit: 30},
		},
		{
			name:    "not allowed for non-admin",
			fields:  fields{Storage: storage, user: &users[3].UserBasic},
			args:    args{page: 1, pageLimit: 30},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &UserStorage{
				Storage: tt.fields.Storage,
				user:    tt.fields.user,
			}
			_, err := u.GetUsersList(tt.args.page, tt.args.pageLimit)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetUsersList() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUserStorage_UsersAmount(t *testing.T) {
	createTempStorage()
	defer deleteTempStorage()

	type fields struct {
		Storage *Storage
		user    *api.UserBasic
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name:   "allowed for admin",
			fields: fields{Storage: storage, user: &users[1].UserBasic},
		},
		{
			name:    "not allowed for non-admin",
			fields:  fields{Storage: storage, user: &users[3].UserBasic},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &UserStorage{
				Storage: tt.fields.Storage,
				user:    tt.fields.user,
			}
			_, err := u.UsersAmount()
			if (err != nil) != tt.wantErr {
				t.Errorf("UsersAmount() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUserStorage_CreateDrawings(t *testing.T) {
	createTempStorage()
	defer deleteTempStorage()

	type fields struct {
		Storage *Storage
		user    *api.UserBasic
	}
	type args struct {
		userID   uint
		drawings []*api.Drawing
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "allowed for admins",
			fields: fields{Storage: storage, user: &users[1].UserBasic},
			args: args{
				userID:   3,
				drawings: []*api.Drawing{randomDrawing(), randomDrawing(), randomDrawing()},
			},
		},
		{
			name:   "allowed for any user if IDs match",
			fields: fields{Storage: storage, user: &users[3].UserBasic},
			args: args{
				userID:   3,
				drawings: []*api.Drawing{randomDrawing(), randomDrawing(), randomDrawing()},
			},
		},
		{
			name: "not allowed for non-admins",
			fields: fields{
				Storage: storage,
				user:    &users[2].UserBasic,
			},
			args: args{
				userID:   3,
				drawings: []*api.Drawing{randomDrawing(), randomDrawing(), randomDrawing()},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &UserStorage{
				Storage: tt.fields.Storage,
				user:    tt.fields.user,
			}
			if err := u.CreateDrawings(tt.args.userID, tt.args.drawings...); (err != nil) != tt.wantErr {
				t.Errorf("CreateDrawings() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func randomDrawing() *api.Drawing {
	d := &api.Drawing{
		DrawingBasic: api.DrawingBasic{Name: generator.GeneratePassword(15, 0, 0, 0)},
		GGDrawing:    *raster.NewGGDrawing(),
	}
	for i, n := 0, rand.Intn(5); i < n; i++ {
		d.Description.PushBack(generator.GeneratePassword(7, 0, 0, 0), generator.GeneratePassword(10, 0, 0, 0))
	}
	for n := rand.Intn(30); d.Len() <= n; {
		switch rand.Intn(2) {
		case 0:
			_ = d.AddPointByAngle(rand.Float64()*150, rand.Float64()*360)
		case 1:
			_ = d.AddPointByDirection(rand.Float64()*150, rand.Float64()*360)
		}
	}
	return d
}

func TestUserStorage_GetDrawing(t *testing.T) {
	createTempStorage()
	defer deleteTempStorage()

	type fields struct {
		Storage *Storage
		user    *api.UserBasic
	}
	type args struct {
		id uint
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "allowed for admin",
			fields: fields{Storage: storage, user: &users[1].UserBasic},
			args:   args{id: 1},
		},
		{
			name:   "allowed for owner",
			fields: fields{Storage: storage, user: &users[2].UserBasic},
			args:   args{id: 1},
		},
		{
			name:   "allowed for user with GET permission",
			fields: fields{Storage: storage, user: &users[3].UserBasic},
			args:   args{id: 2},
		},
		{
			name:    "not allowed",
			fields:  fields{Storage: storage, user: &users[2].UserBasic},
			args:    args{id: 2},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &UserStorage{
				Storage: tt.fields.Storage,
				user:    tt.fields.user,
			}
			_, err := u.GetDrawing(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDrawing() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestUserStorage_GetDrawingOfUser(t *testing.T) {
	createTempStorage()
	defer deleteTempStorage()

	type fields struct {
		Storage *Storage
		user    *api.UserBasic
	}
	type args struct {
		userID    uint
		drawingID uint
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "allowed for admin",
			fields: fields{Storage: storage, user: &users[1].UserBasic},
			args:   args{userID: 3, drawingID: 2},
		},
		{
			name:   "allowed for owner",
			fields: fields{Storage: storage, user: &users[2].UserBasic},
			args:   args{userID: 2, drawingID: 1},
		},
		{
			name:   "allowed for user with GET permission",
			fields: fields{Storage: storage, user: &users[3].UserBasic},
			args:   args{userID: 3, drawingID: 2},
		},
		{
			name:   "allowed of GET permission from another user",
			fields: fields{Storage: storage, user: &users[3].UserBasic},
			args:   args{userID: 2, drawingID: 1},
		},
		{
			name:    "not allowed without permission, even if user IDs match",
			fields:  fields{Storage: storage, user: &users[2].UserBasic},
			args:    args{userID: 2, drawingID: 4},
			wantErr: true,
		},
		{
			name:    "not allowed without permission for different users",
			fields:  fields{Storage: storage, user: &users[2].UserBasic},
			args:    args{userID: 3, drawingID: 4},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &UserStorage{
				Storage: tt.fields.Storage,
				user:    tt.fields.user,
			}
			_, err := u.getDrawingOfUser(tt.args.userID, tt.args.drawingID)
			if (err != nil) != tt.wantErr {
				t.Errorf("getDrawingOfUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestUserStorage_UpdateDrawing(t *testing.T) {
	createTempStorage()
	defer deleteTempStorage()

	type fields struct {
		Storage *Storage
		user    *api.UserBasic
	}
	type args struct {
		drawing *api.Drawing
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "allowed for admin",
			fields: fields{Storage: storage, user: &users[1].UserBasic},
			args:   args{drawing: drawings[1]},
		},
		{
			name:   "allowed for owner",
			fields: fields{Storage: storage, user: &users[2].UserBasic},
			args:   args{drawing: drawings[1]},
		},
		{
			name:   "allowed by Change permission",
			fields: fields{Storage: storage, user: &users[3].UserBasic},
			args:   args{drawing: drawings[1]},
		},
		{
			name:    "not allowed",
			fields:  fields{Storage: storage, user: &users[3].UserBasic},
			args:    args{drawing: drawings[2]},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &UserStorage{
				Storage: tt.fields.Storage,
				user:    tt.fields.user,
			}
			if err := u.UpdateDrawing(tt.args.drawing); (err != nil) != tt.wantErr {
				t.Errorf("UpdateDrawing() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUserStorage_UpdateDrawingOfUser(t *testing.T) {
	createTempStorage()
	defer deleteTempStorage()

	type fields struct {
		Storage *Storage
		user    *api.UserBasic
	}
	type args struct {
		userID  uint
		drawing *api.Drawing
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "allowed for admin",
			fields: fields{Storage: storage, user: &users[1].UserBasic},
			args:   args{userID: 3, drawing: drawings[1]},
		},
		{
			name:   "allowed for owner",
			fields: fields{Storage: storage, user: &users[2].UserBasic},
			args:   args{userID: 2, drawing: drawings[1]},
		},
		{
			name:   "allowed by Change permission",
			fields: fields{Storage: storage, user: &users[2].UserBasic},
			args:   args{userID: 3, drawing: drawings[3]},
		},
		{
			name:    "not allowed",
			fields:  fields{Storage: storage, user: &users[3].UserBasic},
			args:    args{userID: 3, drawing: drawings[2]},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &UserStorage{
				Storage: tt.fields.Storage,
				user:    tt.fields.user,
			}
			if err := u.updateDrawingOfUser(tt.args.userID, tt.args.drawing); (err != nil) != tt.wantErr {
				t.Errorf("updateDrawingOfUser() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUserStorage_RemoveDrawing(t *testing.T) {
	createTempStorage()
	defer deleteTempStorage()

	type fields struct {
		Storage *Storage
		user    *api.UserBasic
	}
	type args struct {
		id uint
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "allowed for admin",
			fields: fields{Storage: storage, user: &users[1].UserBasic},
			args:   args{id: 4},
		},
		{
			name:   "allowed for owner",
			fields: fields{Storage: storage, user: &users[3].UserBasic},
			args:   args{id: 3},
		},
		{
			name:   "allowed by Delete permission",
			fields: fields{Storage: storage, user: &users[3].UserBasic},
			args:   args{id: 1},
		},
		{
			name:    "not allowed",
			fields:  fields{Storage: storage, user: &users[2].UserBasic},
			args:    args{id: 2},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &UserStorage{
				Storage: tt.fields.Storage,
				user:    tt.fields.user,
			}
			if err := u.RemoveDrawing(tt.args.id); (err != nil) != tt.wantErr {
				t.Errorf("RemoveDrawing() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUserStorage_RemoveDrawingOfUser(t *testing.T) {
	createTempStorage()
	defer deleteTempStorage()

	type fields struct {
		Storage *Storage
		user    *api.UserBasic
	}
	type args struct {
		userID    uint
		drawingID uint
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "allowed for admin",
			fields: fields{Storage: storage, user: &users[1].UserBasic},
			args:   args{userID: 3, drawingID: 4},
		},
		{
			name:   "allowed for owner",
			fields: fields{Storage: storage, user: &users[3].UserBasic},
			args:   args{userID: 3, drawingID: 3},
		},
		{
			name:   "allowed by Delete permission",
			fields: fields{Storage: storage, user: &users[3].UserBasic},
			args:   args{userID: 2, drawingID: 1},
		},
		{
			name:    "not allowed",
			fields:  fields{Storage: storage, user: &users[2].UserBasic},
			args:    args{userID: 2, drawingID: 2},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &UserStorage{
				Storage: tt.fields.Storage,
				user:    tt.fields.user,
			}
			if err := u.removeDrawingOfUser(tt.args.userID, tt.args.drawingID); (err != nil) != tt.wantErr {
				t.Errorf("removeDrawingOfUser() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUserStorage_GetDrawingsList(t *testing.T) {
	createTempStorage()
	defer deleteTempStorage()

	type fields struct {
		Storage *Storage
		user    *api.UserBasic
	}
	type args struct {
		userID    uint
		page      uint
		pageLimit uint
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "allowed for admin",
			fields: fields{Storage: storage, user: &users[1].UserBasic},
			args:   args{userID: 3, page: 1, pageLimit: 30},
		},
		{
			name:   "allowed by userID",
			fields: fields{Storage: storage, user: &users[3].UserBasic},
			args:   args{userID: 3, page: 1, pageLimit: 30},
		},
		{
			name:    "not allowed",
			fields:  fields{Storage: storage, user: &users[2].UserBasic},
			args:    args{userID: 3, page: 1, pageLimit: 30},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &UserStorage{
				Storage: tt.fields.Storage,
				user:    tt.fields.user,
			}
			_, err := u.GetDrawingsList(tt.args.userID, tt.args.page, tt.args.pageLimit)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDrawingsList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestUserStorage_DrawingsAmount(t *testing.T) {
	createTempStorage()
	defer deleteTempStorage()

	type fields struct {
		Storage *Storage
		user    *api.UserBasic
	}
	type args struct {
		userID uint
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "allowed for admin",
			fields: fields{Storage: storage, user: &users[1].UserBasic},
			args:   args{userID: 3},
		},
		{
			name:   "allowed by userID",
			fields: fields{Storage: storage, user: &users[3].UserBasic},
			args:   args{userID: 3},
		},
		{
			name:    "not allowed",
			fields:  fields{Storage: storage, user: &users[2].UserBasic},
			args:    args{userID: 3},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &UserStorage{
				Storage: tt.fields.Storage,
				user:    tt.fields.user,
			}
			_, err := u.DrawingsAmount(tt.args.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("DrawingsAmount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestUserStorage_GetDrawingPermission(t *testing.T) {
	createTempStorage()
	defer deleteTempStorage()

	type fields struct {
		Storage *Storage
		user    *api.UserBasic
	}
	type args struct {
		userID    uint
		drawingID uint
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "allowed for admin",
			fields: fields{Storage: storage, user: &users[1].UserBasic},
			args:   args{userID: 2, drawingID: 3},
		},
		{
			name:   "allowed for owner",
			fields: fields{Storage: storage, user: &users[3].UserBasic},
			args:   args{userID: 2, drawingID: 3},
		},
		{
			name:   "allowed by Share permission",
			fields: fields{Storage: storage, user: &users[3].UserBasic},
			args:   args{userID: 3, drawingID: 1},
		},
		{
			name:    "not allowed",
			fields:  fields{Storage: storage, user: &users[2].UserBasic},
			args:    args{userID: 2, drawingID: 4},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &UserStorage{
				Storage: tt.fields.Storage,
				user:    tt.fields.user,
			}
			_, err := u.GetDrawingPermission(tt.args.userID, tt.args.drawingID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDrawingPermission() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestUserStorage_GetDrawingsPermissionsOfDrawing(t *testing.T) {
	createTempStorage()
	defer deleteTempStorage()

	type fields struct {
		Storage *Storage
		user    *api.UserBasic
	}
	type args struct {
		drawingID uint
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "allowed for admin",
			fields: fields{Storage: storage, user: &users[1].UserBasic},
			args:   args{drawingID: 3},
		},
		{
			name:   "allowed for owner",
			fields: fields{Storage: storage, user: &users[2].UserBasic},
			args:   args{drawingID: 1},
		},
		{
			name:   "allowed by Share permission",
			fields: fields{Storage: storage, user: &users[3].UserBasic},
			args:   args{drawingID: 1},
		},
		{
			name:    "not allowed",
			fields:  fields{Storage: storage, user: &users[2].UserBasic},
			args:    args{drawingID: 4},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &UserStorage{
				Storage: tt.fields.Storage,
				user:    tt.fields.user,
			}
			_, err := u.GetDrawingsPermissionsOfDrawing(tt.args.drawingID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDrawingsPermissionsOfDrawing() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestUserStorage_GetDrawingsPermissionsOfUser(t *testing.T) {
	createTempStorage()
	defer deleteTempStorage()

	type fields struct {
		Storage *Storage
		user    *api.UserBasic
	}
	type args struct {
		userID uint
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "allowed for admin",
			fields: fields{Storage: storage, user: &users[1].UserBasic},
			args:   args{userID: 2},
		},
		{
			name:   "allowed for self",
			fields: fields{Storage: storage, user: &users[2].UserBasic},
			args:   args{userID: 2},
		},
		{
			name:    "not allowed",
			fields:  fields{Storage: storage, user: &users[3].UserBasic},
			args:    args{userID: 2},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &UserStorage{
				Storage: tt.fields.Storage,
				user:    tt.fields.user,
			}
			_, err := u.GetDrawingsPermissionsOfUser(tt.args.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDrawingsPermissionsOfUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestUserStorage_CreateDrawingPermission(t *testing.T) {
	createTempStorage()
	defer deleteTempStorage()

	type fields struct {
		Storage *Storage
		user    *api.UserBasic
	}
	type args struct {
		permission *api.DrawingPermission
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "allowed for admin",
			fields: fields{Storage: storage, user: &users[1].UserBasic},
			args: args{permission: &api.DrawingPermission{
				User:    &users[2].UserBasic,
				Drawing: &drawings[4].DrawingBasic,
				Get:     true,
			}},
		},
		{
			name:   "allowed for owner",
			fields: fields{Storage: storage, user: &users[3].UserBasic},
			args: args{permission: &api.DrawingPermission{
				User:    &users[1].UserBasic,
				Drawing: &drawings[3].DrawingBasic,
				Get:     true,
			}},
		},
		{
			name:   "allowed by Share permission",
			fields: fields{Storage: storage, user: &users[3].UserBasic},
			args: args{permission: &api.DrawingPermission{
				User:    &users[1].UserBasic,
				Drawing: &drawings[1].DrawingBasic,
				Get:     true,
			}},
		},
		{
			name:   "not allowed",
			fields: fields{Storage: storage, user: &users[2].UserBasic},
			args: args{permission: &api.DrawingPermission{
				User:    &users[1].UserBasic,
				Drawing: &drawings[1].DrawingBasic,
				Get:     true,
			}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &UserStorage{
				Storage: tt.fields.Storage,
				user:    tt.fields.user,
			}
			if err := u.CreateDrawingPermission(tt.args.permission); (err != nil) != tt.wantErr {
				t.Errorf("CreateDrawingPermission() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUserStorage_UpdateDrawingPermission(t *testing.T) {
	createTempStorage()
	defer deleteTempStorage()

	type fields struct {
		Storage *Storage
		user    *api.UserBasic
	}
	type args struct {
		permission *api.DrawingPermission
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "allowed for admin",
			fields: fields{Storage: storage, user: &users[1].UserBasic},
			args: args{permission: &api.DrawingPermission{
				User:    &users[2].UserBasic,
				Drawing: &drawings[3].DrawingBasic,
				Get:     true,
			}},
		},
		{
			name:   "allowed for owner",
			fields: fields{Storage: storage, user: &users[3].UserBasic},
			args: args{permission: &api.DrawingPermission{
				User:    &users[2].UserBasic,
				Drawing: &drawings[3].DrawingBasic,
				Get:     true,
				Share:   true,
			}},
		},
		{
			name:   "allowed by Share permission",
			fields: fields{Storage: storage, user: &users[2].UserBasic},
			args: args{permission: &api.DrawingPermission{
				User:    &users[3].UserBasic,
				Drawing: &drawings[1].DrawingBasic,
				Get:     true,
				Share:   true,
			}},
		},
		{
			name:   "not allowed",
			fields: fields{Storage: storage, user: &users[2].UserBasic},
			args: args{permission: &api.DrawingPermission{
				User:    &users[3].UserBasic,
				Drawing: &drawings[1].DrawingBasic,
				Get:     true,
				Share:   true,
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &UserStorage{
				Storage: tt.fields.Storage,
				user:    tt.fields.user,
			}
			if err := u.UpdateDrawingPermission(tt.args.permission); (err != nil) != tt.wantErr {
				t.Errorf("UpdateDrawingPermission() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUserStorage_RemoveDrawingPermission(t *testing.T) {
	createTempStorage()
	defer deleteTempStorage()

	type fields struct {
		Storage *Storage
		user    *api.UserBasic
	}
	type args struct {
		userID    uint
		drawingID uint
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "allowed for admin",
			fields: fields{Storage: storage, user: &users[1].UserBasic},
			args:   args{userID: 3, drawingID: 2},
		},
		{
			name:   "allowed for owner",
			fields: fields{Storage: storage, user: &users[3].UserBasic},
			args:   args{userID: 1, drawingID: 4},
		},
		{
			name:   "allowed by Share permission",
			fields: fields{Storage: storage, user: &users[3].UserBasic},
			args:   args{userID: 3, drawingID: 1},
		},
		{
			name:    "not allowed",
			fields:  fields{Storage: storage, user: &users[2].UserBasic},
			args:    args{userID: 2, drawingID: 3},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &UserStorage{
				Storage: tt.fields.Storage,
				user:    tt.fields.user,
			}
			if err := u.RemoveDrawingPermission(tt.args.userID, tt.args.drawingID); (err != nil) != tt.wantErr {
				t.Errorf("RemoveDrawingPermission() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
