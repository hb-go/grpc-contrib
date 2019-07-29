package registry

type Option func(options *Options)

type Options struct {
	Version string
	Addr    string
}

func WithVersion(version string) Option {
	return func(options *Options) {
		options.Version = version
	}
}

func WithAddr(addr string) Option {
	return func(options *Options) {
		options.Addr = addr
	}
}
