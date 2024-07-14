package logic

import (
	"context"

	"zeroim/imrpc/imrpc"
	"zeroim/imrpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type LogoutLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLogoutLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LogoutLogic {
	return &LogoutLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *LogoutLogic) Logout(in *imrpc.LogoutRequest) (*imrpc.LogoutResponse, error) {
	// todo: add your logic here and delete this line

	return &imrpc.LogoutResponse{}, nil
}
