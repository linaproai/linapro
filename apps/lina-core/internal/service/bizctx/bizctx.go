package bizctx

import (
	"context"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gctx"

	"lina-core/internal/model"
)

// contextKey is the key for business context in request context.
const contextKey gctx.StrKey = "BizCtx"

// Service defines the bizctx service contract.
type Service interface {
	// Init initializes and injects business context into request.
	Init(r *ghttp.Request, ctx *model.Context)
	// Get retrieves business context from context.
	Get(ctx context.Context) *model.Context
	// SetUser sets user info into business context.
	SetUser(ctx context.Context, tokenId string, userId int, username string, status int)
}

var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct{}

// New creates and returns a new Service instance.
func New() Service {
	return &serviceImpl{}
}

// Init initializes and injects business context into request.
func (s *serviceImpl) Init(r *ghttp.Request, ctx *model.Context) {
	r.SetCtxVar(contextKey, ctx)
}

// Get retrieves business context from context.
func (s *serviceImpl) Get(ctx context.Context) *model.Context {
	value := ctx.Value(contextKey)
	if value == nil {
		return nil
	}
	if localCtx, ok := value.(*model.Context); ok {
		return localCtx
	}
	return nil
}

// SetUser sets user info into business context.
func (s *serviceImpl) SetUser(ctx context.Context, tokenId string, userId int, username string, status int) {
	if c := s.Get(ctx); c != nil {
		c.TokenId = tokenId
		c.UserId = userId
		c.Username = username
		c.Status = status
	}
}
