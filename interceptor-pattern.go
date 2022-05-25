package cloudy

import "context"

type BeforeInterceptor[T any] interface {
	BeforeAction(ctx context.Context, item *T) (*T, error)
}

type AfterInterceptor[T any] interface {
	AfterAction(ctx context.Context, item *T) (*T, error)
}
