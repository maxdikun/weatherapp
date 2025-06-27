package handlers

import (
	"context"
	"time"

	"github.com/maxdikun/weatherapp/internal/handlers/gen"
	"github.com/maxdikun/weatherapp/internal/services"
)

type ApiHandler struct {
	userSvc *services.UserService
}

var _ gen.StrictServerInterface = (*ApiHandler)(nil)

// Register implements gen.StrictServerInterface.
func (api *ApiHandler) Register(ctx context.Context, request gen.RegisterRequestObject) (gen.RegisterResponseObject, error) {
	res, err := api.userSvc.Register(ctx, request.Body.Login, request.Body.Password)
	if err != nil {
		return gen.Register500JSONResponse{
			Code:      "INTERNAL_ERROR",
			Timestamp: time.Now(),
			Message:   "Internal service error occurred, try later",
		}, nil
	}

	return gen.Register200JSONResponse{
		AccessToken:           res.Access,
		RefreshToken:          res.Refresh,
		RefreshTokenExpiresAt: res.RefreshExpiresAt,
	}, nil
}
