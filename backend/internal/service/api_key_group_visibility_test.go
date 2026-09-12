//go:build unit

package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type visibilityUserRepo struct {
	UserRepository
	user *User
	err  error
}

func (r *visibilityUserRepo) GetByID(context.Context, int64) (*User, error) { return r.user, r.err }

type visibilitySubRepo struct {
	UserSubscriptionRepository
	subscriptions []UserSubscription
	err           error
	calls         int
}

func (r *visibilitySubRepo) ListActiveByUserID(_ context.Context, userID int64) ([]UserSubscription, error) {
	r.calls++
	// 对照真实仓储的用户、状态与到期条件，过期记录不进入可绑定集合。
	active := make([]UserSubscription, 0)
	for _, sub := range r.subscriptions {
		if sub.UserID == userID && sub.IsActive() {
			active = append(active, sub)
		}
	}
	return active, r.err
}

type visibilityGroupRepo struct {
	GroupRepository
	groups []Group
	err    error
}

func (r *visibilityGroupRepo) ListActive(context.Context) ([]Group, error) { return r.groups, r.err }

func TestAvailableGroupVisibilityIncludesActiveSubscriptions(t *testing.T) {
	for _, restricted := range []bool{false, true} {
		t.Run(map[bool]string{false: "unrestricted", true: "restricted"}[restricted], func(t *testing.T) {
			now := time.Now()
			subs := &visibilitySubRepo{subscriptions: []UserSubscription{
				{UserID: 1, GroupID: 42, Status: SubscriptionStatusActive, ExpiresAt: now.Add(time.Hour)},
				{UserID: 1, GroupID: 43, Status: SubscriptionStatusActive, ExpiresAt: now.Add(-time.Hour)},
				{UserID: 1, GroupID: 44, Status: "expired", ExpiresAt: now.Add(time.Hour)},
				{UserID: 2, GroupID: 45, Status: SubscriptionStatusActive, ExpiresAt: now.Add(time.Hour)},
			}}
			svc := &APIKeyService{
				userRepo:    &visibilityUserRepo{user: &User{ID: 1, AllowedGroups: []int64{7}, RestrictPublicGroups: restricted}},
				userSubRepo: subs,
				groupRepo: &visibilityGroupRepo{groups: []Group{
					{ID: 7, IsExclusive: true, SubscriptionType: "standard"},
					{ID: 42, IsExclusive: true, SubscriptionType: "subscription"},
					{ID: 43, IsExclusive: true, SubscriptionType: "subscription"},
					{ID: 44, IsExclusive: true, SubscriptionType: "subscription"},
					{ID: 45, IsExclusive: true, SubscriptionType: "subscription"},
					{ID: 46, SubscriptionType: "standard"},
				}},
			}
			available, err := svc.GetAvailableGroups(context.Background(), 1)
			require.NoError(t, err)
			ids := make([]int64, 0, len(available))
			for _, group := range available {
				ids = append(ids, group.ID)
			}
			expected := []int64{7, 42}
			if !restricted {
				expected = append(expected, 46)
			}
			require.Equal(t, expected, ids)
			require.Equal(t, 1, subs.calls)
		})
	}
}

func TestAvailableGroupVisibilityEmptyAndErrors(t *testing.T) {
	failure := errors.New("repository unavailable")
	for _, tc := range []struct {
		name                      string
		userErr, groupErr, subErr error
	}{
		{name: "empty"},
		{name: "user failure", userErr: failure},
		{name: "group failure", groupErr: failure},
		{name: "subscription failure", subErr: failure},
	} {
		t.Run(tc.name, func(t *testing.T) {
			subs := &visibilitySubRepo{err: tc.subErr}
			svc := &APIKeyService{
				userRepo:    &visibilityUserRepo{user: &User{ID: 1}, err: tc.userErr},
				groupRepo:   &visibilityGroupRepo{err: tc.groupErr},
				userSubRepo: subs,
			}
			got, err := svc.GetAvailableGroups(context.Background(), 1)
			if tc.userErr != nil || tc.groupErr != nil || tc.subErr != nil {
				require.ErrorIs(t, err, failure)
				require.Nil(t, got, "仓储错误不得转成匿名可见性")
			} else {
				require.NoError(t, err)
				require.NotNil(t, got, "空登录用户不得转成匿名")
				require.Empty(t, got)
			}
			if tc.userErr != nil || tc.groupErr != nil {
				require.Zero(t, subs.calls)
			}
		})
	}
}
