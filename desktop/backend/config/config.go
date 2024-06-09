package config

type Config struct {
	Opts
}

type Opts struct {
	Debug  bool
	DBFile string
}

type OptFunc func(*Opts)

func defaultOpts() Opts {
	return Opts{
		Debug:  false,
		DBFile: "",
	}
}

func WithDebug(d bool) OptFunc {
	return func(opts *Opts) {
		opts.Debug = d
	}
}

func WithDBFile(d string) OptFunc {
	return func(opts *Opts) {
		opts.DBFile = d
	}
}

func NewConfig(opts ...OptFunc) Config {
	o := defaultOpts()
	for _, fn := range opts {
		fn(&o)
	}

	return Config{Opts: o}
}
