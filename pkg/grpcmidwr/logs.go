package grpcmidwr

import (
	"context"
	"fmt"

	"github.com/0wnperception/go-helpers/pkg/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

func WithLogUnaryInterceptor(logCtx context.Context) grpc.ServerOption {
	return grpc.UnaryInterceptor(func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		var consumer string

		netInfo, ok := peer.FromContext(ctx)
		if ok {
			consumer = netInfo.Addr.String()
		}

		l := log.FromContext(logCtx)

		logCtx := l.Inject(ctx)

		logCtx = log.WithFields(logCtx,
			log.String("consumer", consumer),
			log.String("method", info.FullMethod),
			log.Reflect("request", req),
		)

		resp, err := handler(logCtx, req)

		st, ok := status.FromError(err)
		if ok {
			log.Debug(logCtx, fmt.Sprintf("status '%d'", st.Code()), log.Reflect("response", resp))
		} else {
			log.Debug(logCtx, "status 'uncknown'", log.Reflect("response", resp))
		}

		return resp, err
	})
}

func WithLogStreamInterceptor(logCtx context.Context) grpc.ServerOption {
	return grpc.StreamInterceptor(func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		streamCtx := ss.Context()

		var consumer string

		netInfo, ok := peer.FromContext(streamCtx)
		if ok {
			consumer = netInfo.Addr.String()
		}

		l := log.FromContext(logCtx)

		streamLogCtx := l.Inject(streamCtx)

		streamLogCtx = log.WithFields(streamLogCtx,
			log.String("consumer", consumer),
			log.String("method", info.FullMethod),
		)

		log.Debug(streamLogCtx, "new stream connection")

		newStream := WrapServerStream(ss)
		newStream.LogContext = streamLogCtx

		err := handler(srv, newStream)

		log.Debug(streamLogCtx, "stream connection is closed")

		return err
	})
}

func WithLogInterceptors(logCtx context.Context) []grpc.ServerOption {
	return []grpc.ServerOption{
		WithLogUnaryInterceptor(logCtx),
		WithLogStreamInterceptor(logCtx),
	}
}

// WrappedServerStream is a thin wrapper around grpc.ServerStream that allows modifying context.
type StreamLogWrapper struct {
	grpc.ServerStream
	// WrappedContext is the wrapper's own Context. You can assign it.
	LogContext context.Context
}

// Context returns the wrapper's WrappedContext, overwriting the nested grpc.ServerStream.Context()
func (w *StreamLogWrapper) Context() context.Context {
	return w.LogContext
}

func (w *StreamLogWrapper) SendMsg(m interface{}) error {
	ctx := log.WithFields(w.LogContext, log.Reflect("update", m))

	log.Debug(ctx, "send to stream")

	if err := w.ServerStream.SendMsg(m); err != nil {
		log.Err(ctx, "sending update error", log.Error(err))

		return err
	}

	return nil
}

// WrapServerStream returns a ServerStream that has the ability to overwrite context.
func WrapServerStream(stream grpc.ServerStream) *StreamLogWrapper {
	if existing, ok := stream.(*StreamLogWrapper); ok {
		return existing
	}
	return &StreamLogWrapper{ServerStream: stream, LogContext: stream.Context()}
}
