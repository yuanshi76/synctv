package model_test

import (
	"testing"

	"github.com/synctv-org/synctv/internal/model"
)

func TestRoomMemberPermissionBits(t *testing.T) {
	p := model.NoPermission
	if p.Has(model.PermissionAddMovie) {
		t.Fatal("empty permission should not have AddMovie")
	}

	p = p.Add(model.PermissionAddMovie).Add(model.PermissionEditMovie)
	if !p.Has(model.PermissionAddMovie) || !p.Has(model.PermissionEditMovie) {
		t.Fatal("added permissions should be present")
	}
	if p.Has(model.PermissionDeleteMovie) {
		t.Fatal("non-added permission should be absent")
	}

	p = p.Remove(model.PermissionAddMovie)
	if p.Has(model.PermissionAddMovie) {
		t.Fatal("removed permission should be absent")
	}
	if !p.Has(model.PermissionEditMovie) {
		t.Fatal("untouched permission should remain")
	}

	if !model.AllPermissions.Has(model.PermissionSetCurrentMovie) {
		t.Fatal("AllPermissions should include every permission")
	}
}

func TestRoomAdminPermissionBits(t *testing.T) {
	p := model.NoAdminPermission.
		Add(model.PermissionBanRoomMember).
		Add(model.PermissionDeleteRoom)
	if !p.Has(model.PermissionBanRoomMember) || !p.Has(model.PermissionDeleteRoom) {
		t.Fatal("added admin permissions should be present")
	}
	if p.Has(model.PermissionSetRoomPassword) {
		t.Fatal("non-added admin permission should be absent")
	}

	p = p.Remove(model.PermissionDeleteRoom)
	if p.Has(model.PermissionDeleteRoom) {
		t.Fatal("removed admin permission should be absent")
	}
}

func TestRoomMemberRoleHierarchy(t *testing.T) {
	if !model.RoomMemberRoleCreator.IsAdmin() {
		t.Error("creator should be admin")
	}
	if !model.RoomMemberRoleCreator.IsMember() {
		t.Error("creator should be member")
	}
	if !model.RoomMemberRoleAdmin.IsMember() {
		t.Error("admin should be member")
	}
	if model.RoomMemberRoleMember.IsAdmin() {
		t.Error("member should not be admin")
	}
	if model.RoomMemberRoleUnknown.IsMember() {
		t.Error("unknown should not be member")
	}
}

func TestHasPermission(t *testing.T) {
	tests := []struct {
		name   string
		member model.RoomMember
		perm   model.RoomMemberPermission
		want   bool
	}{
		{
			name: "admin bypasses status and explicit perms",
			member: model.RoomMember{
				Role:   model.RoomMemberRoleAdmin,
				Status: model.RoomMemberStatusBanned,
			},
			perm: model.PermissionAddMovie,
			want: true,
		},
		{
			name: "active member with permission",
			member: model.RoomMember{
				Role:        model.RoomMemberRoleMember,
				Status:      model.RoomMemberStatusActive,
				Permissions: model.PermissionAddMovie,
			},
			perm: model.PermissionAddMovie,
			want: true,
		},
		{
			name: "active member without permission",
			member: model.RoomMember{
				Role:        model.RoomMemberRoleMember,
				Status:      model.RoomMemberStatusActive,
				Permissions: model.PermissionGetMovieList,
			},
			perm: model.PermissionAddMovie,
			want: false,
		},
		{
			name: "member with permission but not active",
			member: model.RoomMember{
				Role:        model.RoomMemberRoleMember,
				Status:      model.RoomMemberStatusPending,
				Permissions: model.PermissionAddMovie,
			},
			perm: model.PermissionAddMovie,
			want: false,
		},
		{
			name: "non-member role",
			member: model.RoomMember{
				Role:   model.RoomMemberRoleUnknown,
				Status: model.RoomMemberStatusActive,
			},
			perm: model.PermissionAddMovie,
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.member.HasPermission(tt.perm); got != tt.want {
				t.Errorf("HasPermission() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHasAdminPermission(t *testing.T) {
	tests := []struct {
		name   string
		member model.RoomMember
		perm   model.RoomAdminPermission
		want   bool
	}{
		{
			name: "creator bypasses status and explicit perms",
			member: model.RoomMember{
				Role:   model.RoomMemberRoleCreator,
				Status: model.RoomMemberStatusBanned,
			},
			perm: model.PermissionDeleteRoom,
			want: true,
		},
		{
			name: "active admin with admin permission",
			member: model.RoomMember{
				Role:             model.RoomMemberRoleAdmin,
				Status:           model.RoomMemberStatusActive,
				AdminPermissions: model.PermissionBanRoomMember,
			},
			perm: model.PermissionBanRoomMember,
			want: true,
		},
		{
			name: "active admin without admin permission",
			member: model.RoomMember{
				Role:             model.RoomMemberRoleAdmin,
				Status:           model.RoomMemberStatusActive,
				AdminPermissions: model.PermissionBanRoomMember,
			},
			perm: model.PermissionDeleteRoom,
			want: false,
		},
		{
			name: "admin not active",
			member: model.RoomMember{
				Role:             model.RoomMemberRoleAdmin,
				Status:           model.RoomMemberStatusPending,
				AdminPermissions: model.PermissionBanRoomMember,
			},
			perm: model.PermissionBanRoomMember,
			want: false,
		},
		{
			name: "plain member has no admin permission",
			member: model.RoomMember{
				Role:   model.RoomMemberRoleMember,
				Status: model.RoomMemberStatusActive,
			},
			perm: model.PermissionBanRoomMember,
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.member.HasAdminPermission(tt.perm); got != tt.want {
				t.Errorf("HasAdminPermission() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRoomMemberStatusHelpers(t *testing.T) {
	if !model.RoomMemberStatusActive.IsActive() {
		t.Error("active status IsActive should be true")
	}
	if model.RoomMemberStatusActive.IsNotActive() {
		t.Error("active status IsNotActive should be false")
	}
	if !model.RoomMemberStatusPending.IsPending() {
		t.Error("pending status IsPending should be true")
	}
	if !model.RoomMemberStatusBanned.IsBanned() {
		t.Error("banned status IsBanned should be true")
	}
	if model.RoomMemberStatusActive.String() != "active" {
		t.Errorf("status string = %q, want active", model.RoomMemberStatusActive.String())
	}
}
