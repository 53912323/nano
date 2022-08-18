package nano

import (
	"net/http"
	"time"

	"github.com/lonng/nano/metrics"
	"github.com/lonng/nano/session"

	"github.com/lonng/nano/cluster"
	"github.com/lonng/nano/component"
	"github.com/lonng/nano/internal/env"
	"github.com/lonng/nano/internal/log"
	"github.com/lonng/nano/internal/message"
	"github.com/lonng/nano/pipeline"
	"github.com/lonng/nano/serialize"
	"google.golang.org/grpc"
)

type Option func(*cluster.Options)

func WithPipeline(pipeline pipeline.Pipeline) Option {
	return func(opt *cluster.Options) {
		opt.Pipeline = pipeline
	}
}

func WithIncreaseCheck() Option {
	return func(opt *cluster.Options) {
		env.IncreaseCheck = true
	}
}

// WithAdvertiseAddr sets the advertise address option, it will be the listen address in
// master node and an advertise address which cluster member to connect
func WithAdvertiseAddr(addr string, retryInterval ...time.Duration) Option {
	return func(opt *cluster.Options) {
		opt.AdvertiseAddr = addr
		if len(retryInterval) > 0 {
			opt.RetryInterval = retryInterval[0]
		}
	}
}

// WithMemberAddr sets the listen address which is used to establish connection between
// cluster members. Will select an available port automatically if no member address
// setting and panic if no available port
func WithClientAddr(addr string) Option {
	return func(opt *cluster.Options) {
		opt.ClientAddr = addr
	}
}

// WithMaster sets the option to indicate whether the current node is master node
func WithMaster() Option {
	return func(opt *cluster.Options) {
		opt.IsMaster = true
	}
}

// WithGrpcOptions sets the grpc dial options
func WithGrpcOptions(opts ...grpc.DialOption) Option {
	return func(_ *cluster.Options) {
		env.GrpcOptions = append(env.GrpcOptions, opts...)
	}
}

// WithComponents sets the Components
func WithComponents(components *component.Components) Option {
	return func(opt *cluster.Options) {
		opt.Components = components
	}
}

// 前置处理函数
func WithBefore(funcBefore func(session *session.Session, route string, msg interface{}, ns int64) bool) Option {
	return func(opt *cluster.Options) {
		opt.FuncBefore = funcBefore
	}
}

// 后置处理函数
func WithAfter(funcAfter func(session *session.Session, route string, msg interface{}) bool) Option {
	return func(opt *cluster.Options) {
		opt.FuncAfter = funcAfter
	}
}

func WithRoutable(fn func(session *session.Session, route string) bool) Option {
	return func(opt *cluster.Options) {
		opt.Routable = fn
	}
}

// 流量控制
func WithRateLimit(limit int, interval time.Duration) Option {
	return func(_ *cluster.Options) {
		if limit > 0 {
			env.RateLimit = env.NewRateLimitingMaker(limit, interval)
		}
	}
}

// WithHeartbeatInterval sets Heartbeat time interval
func WithHeartbeatInterval(d time.Duration) Option {
	return func(_ *cluster.Options) {
		env.Heartbeat = d
	}
}

// WithCheckOriginFunc sets the function that check `Origin` in http headers
func WithCheckOriginFunc(fn func(*http.Request) bool) Option {
	return func(opt *cluster.Options) {
		env.CheckOrigin = fn
	}
}

// WithDebugMode let 'nano' to run under Debug mode.
func WithDebugMode() Option {
	return func(_ *cluster.Options) {
		env.Debug = true
	}
}

// WithProtoMode
func WithProtoRoute() Option {
	return func(_ *cluster.Options) {
		env.ProtoRoute = true
	}
}

// WithTestTcp
func WithTestTcp() Option {
	return func(_ *cluster.Options) {
		env.TestTcp = true
	}
}

// WithTestTcp
func WithTcpAddr(addr string) Option {
	return func(_ *cluster.Options) {
		env.TcpAddr = addr
	}
}

// SetDictionary sets routes map
func WithDictionary(dict map[string]uint16) Option {
	return func(_ *cluster.Options) {
		env.RouteDict = dict
		message.SetDictionary(dict)
	}
}

func WithWSPath(path string) Option {
	return func(_ *cluster.Options) {
		env.WSPath = path
	}
}

// SetTimerPrecision sets the ticker precision, and time precision can not less
// than a Millisecond, and can not change after application running. The default
// precision is time.Second
func WithTimerPrecision(precision time.Duration) Option {
	if precision < time.Millisecond {
		panic("time precision can not less than a Millisecond")
	}
	return func(_ *cluster.Options) {
		env.TimerPrecision = precision
	}
}

// WithSerializer customizes application serializer, which automatically Marshal
// and UnMarshal handler payload
func WithSerializer(serializer serialize.Serializer) Option {
	return func(opt *cluster.Options) {
		env.Serializer = serializer
	}
}

// WithLabel sets the current node label in cluster
func WithLabel(label string) Option {
	return func(opt *cluster.Options) {
		opt.Label = label
	}
}

// WithIsWebsocket indicates whether current node WebSocket is enabled
func WithIsWebsocket(enableWs bool) Option {
	return func(opt *cluster.Options) {
		opt.IsWebsocket = enableWs
	}
}

// WithTSLConfig sets the `key` and `certificate` of TSL
func WithTSLConfig(certificate, key string) Option {
	return func(opt *cluster.Options) {
		opt.TSLCertificate = certificate
		opt.TSLKey = key
	}
}

// WithLogger overrides the default logger
func WithLogger(l log.Logger) Option {
	return func(opt *cluster.Options) {
		log.SetLogger(l)
	}
}

// Metrics
func WithMetrics(reporters []metrics.Reporter, period time.Duration) Option {
	return func(opt *cluster.Options) {
		if len(reporters) > 0 {
			opt.MetricsReporters = reporters
			opt.MetricsPeriod = period
		}
	}
}
