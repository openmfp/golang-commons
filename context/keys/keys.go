package keys

type ContextKey string

const (
	RequestIdCtxKey      = ContextKey("request-id")
	LoggerCtxKey         = ContextKey("logger")
	ConfigCtxKey         = ContextKey("config")
	SentryTagsCtxKey     = ContextKey("sentryTags")
	TracingHeadersCtxKey = ContextKey("tracingHeaders")
	TechnicalUserCtxKey  = ContextKey("technicalUser")
)
