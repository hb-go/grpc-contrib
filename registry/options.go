package registry

type Option func(options *Options)

type Options struct {
	Versions []string
	Addr     string
}

func WithVersion(version ...string) Option {
	return func(options *Options) {
		options.Versions = append(options.Versions, version...)
	}
}

func WithAddr(addr string) Option {
	return func(options *Options) {
		options.Addr = addr
	}
}
