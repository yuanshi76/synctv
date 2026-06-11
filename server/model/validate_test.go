package model_test

import (
	"errors"
	"strings"
	"testing"

	model "github.com/synctv-org/synctv/server/model"
)

func TestCreateRoomReqValidate(t *testing.T) {
	tests := []struct {
		name    string
		room    string
		pass    string
		wantErr error
	}{
		{"valid no password", "movie night", "", nil},
		{"valid with han name", "电影之夜", "", nil},
		{"valid with password", "room1", "secret-123", nil},
		{"empty name", "", "", model.ErrEmptyRoomName},
		{"name too long", strings.Repeat("a", 33), "", model.ErrRoomNameTooLong},
		{"name newline injection", "room\nname", "", model.ErrRoomNameHasInvalidChar},
		{"password too long", "room", strings.Repeat("a", 33), model.ErrPasswordTooLong},
		{"password han not allowed", "room", "密码", model.ErrPasswordHasInvalidChar},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &model.CreateRoomReq{RoomName: tt.room, Password: tt.pass}
			err := req.Validate()
			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("Validate() = %v, want nil", err)
				}
				return
			}
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Validate() = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoginRoomReqValidate(t *testing.T) {
	valid := strings.Repeat("a", 32)
	tests := []struct {
		name    string
		roomID  string
		wantErr bool
	}{
		{"valid 32-char id", valid, false},
		{"empty id", "", true},
		{"too short id", "abc", true},
		{"too long id", strings.Repeat("a", 33), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &model.LoginRoomReq{RoomID: tt.roomID}
			if err := req.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() err = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRoomIDReqValidate(t *testing.T) {
	if err := (&model.RoomIDReq{ID: strings.Repeat("a", 32)}).Validate(); err != nil {
		t.Errorf("valid id should pass, got %v", err)
	}
	if err := (&model.RoomIDReq{ID: "short"}).Validate(); err == nil {
		t.Error("short id should fail")
	}
}

func TestLoginUserReqValidate(t *testing.T) {
	tests := []struct {
		name     string
		username string
		email    string
		password string
		wantErr  bool
	}{
		{"valid username login", "alice", "", "pass123", false},
		{"valid email login", "", "a@b.com", "pass123", false},
		{"both username and email", "alice", "a@b.com", "pass123", true},
		{"neither username nor email", "", "", "pass123", true},
		{"invalid email", "", "not-an-email", "pass123", true},
		{"username too long", strings.Repeat("a", 33), "", "pass123", true},
		{"empty password", "alice", "", "", true},
		{"password with newline", "alice", "", "pa\nss", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &model.LoginUserReq{Username: tt.username, Email: tt.email, Password: tt.password}
			if err := req.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() err = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUserSignupPasswordReqValidate(t *testing.T) {
	tests := []struct {
		name     string
		username string
		password string
		wantErr  bool
	}{
		{"valid", "bob", "secret", false},
		{"valid han username", "小明", "secret", false},
		{"empty username", "", "secret", true},
		{"username too long", strings.Repeat("a", 33), "secret", true},
		{"empty password", "bob", "", true},
		{"password too long", "bob", strings.Repeat("a", 33), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &model.UserSignupPasswordReq{Username: tt.username, Password: tt.password}
			if err := req.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() err = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSetUsernameReqValidate(t *testing.T) {
	if err := (&model.SetUsernameReq{Username: "valid_name"}).Validate(); err != nil {
		t.Errorf("valid username should pass, got %v", err)
	}
	if err := (&model.SetUsernameReq{Username: ""}).Validate(); err == nil {
		t.Error("empty username should fail")
	}
	if err := (&model.SetUsernameReq{Username: strings.Repeat("a", 33)}).Validate(); !errors.Is(err, model.ErrUsernameTooLong) {
		t.Errorf("long username should fail with ErrUsernameTooLong, got %v", err)
	}
}

func TestUserIDReqValidate(t *testing.T) {
	if err := (&model.UserIDReq{ID: strings.Repeat("a", 32)}).Validate(); err != nil {
		t.Errorf("valid id should pass, got %v", err)
	}
	if err := (&model.UserIDReq{ID: ""}).Validate(); err == nil {
		t.Error("empty id should fail")
	}
}

func TestEmailValidation(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr error
	}{
		{"valid", "user@example.com", nil},
		{"valid subdomain", "a.b@mail.example.co", nil},
		{"empty", "", errors.New("email is empty")},
		{"missing at", "userexample.com", model.ErrInvalidEmail},
		{"missing tld", "user@example", model.ErrInvalidEmail},
		{"too long", strings.Repeat("a", 120) + "@example.com", model.ErrEmailTooLong},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &model.UserSendBindEmailCaptchaReq{Email: tt.email, CaptchaID: "c", Answer: "a"}
			err := req.Validate()
			switch {
			case tt.wantErr == nil:
				if err != nil {
					t.Fatalf("Validate() = %v, want nil", err)
				}
			case errors.Is(tt.wantErr, model.ErrInvalidEmail) || errors.Is(tt.wantErr, model.ErrEmailTooLong):
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("Validate() = %v, want %v", err, tt.wantErr)
				}
			default:
				if err == nil {
					t.Errorf("Validate() = nil, want error")
				}
			}
		})
	}
}
