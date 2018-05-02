package helper

import (
	"context"
	"net/url"

	"github.com/vmware/govmomi/session"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/types"
)

func CreateSession(ctx context.Context, c *vim25.Client, u *url.Userinfo) (*types.UserSession, error) {
	manager := session.NewManager(c)

	err := manager.Login(ctx, u)
	if err != nil {
		return nil, err
	}

	return manager.UserSession(ctx)
}
