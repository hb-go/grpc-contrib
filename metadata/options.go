package metadata

type options struct {
	headers  map[string]string // value为新的header key, value == ""使用原key
	prefixes map[string]string // value为前缀替换
}

type Option func(option *options)

func evaluateOptions(opts []Option) *options {
	opt := &options{
		headers:  make(map[string]string),
		prefixes: make(map[string]string),
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
			option.headers[header] = ""
		}
	}
}

func WithHeaderReplace(header, replace string) Option {
	return func(option *options) {
		if _, ok := option.headers[header]; ok {
			return
		} else {
			option.headers[header] = replace
		}
	}
}

func WithPrefix(prefix string) Option {
	return func(option *options) {
		if _, ok := option.prefixes[prefix]; ok {
			return
		} else {
			option.prefixes[prefix] = prefix
		}
	}
}

func WithPrefixReplace(prefix, replace string) Option {
	return func(option *options) {
		if _, ok := option.prefixes[prefix]; ok {
			return
		} else {
			option.prefixes[prefix] = replace
		}
	}
}
