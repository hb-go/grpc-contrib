package metadata

type options struct {
	headers  map[string]bool
	prefixes map[string]bool
}

type Option func(option *options)

func evaluateOptions(opts []Option) *options {
	opt := &options{
		headers:  make(map[string]bool),
		prefixes: make(map[string]bool),
	}
	for _, o := range opts {
		o(opt)
	}

	return opt
}

func WithHeader(header string) Option {
	return func(option *options) {
		if _, ok := option.headers[header]; ok {
			return
		} else {
			option.headers[header] = true
		}
	}
}

func WithPrefix(header string) Option {
	return func(option *options) {
		if _, ok := option.headers[header]; ok {
			return
		} else {
			option.headers[header] = true
		}
	}
}
