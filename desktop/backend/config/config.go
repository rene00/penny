package config

import (
	"os"
	"path"
	"path/filepath"
)

const AppName string = "pennyapp"

type Config struct {
	Opts
}

type Opts struct {
	Debug bool

	// The path of the sqlite3 db file.
	DBFile string

	// The path of the app home dir.
	DataHome string

	// The path of the import dir where imported files are temporarily stored.
	ImportDir string

	// The path of the temp dir.
	TmpDir string
}

type OptFunc func(*Opts)

func defaultOpts() Opts {
	return Opts{
		Debug:     false,
		DBFile:    "",
		DataHome:  "",
		ImportDir: "",
	}
}

func WithDebug(d bool) OptFunc {
	return func(opts *Opts) {
		opts.Debug = d
	}
}

func WithDataHome(d string) OptFunc {
	return func(opts *Opts) {
		opts.DataHome = d
	}
}

func WithImportDir(d string) OptFunc {
	return func(opts *Opts) {
		opts.ImportDir = d
	}
}

func WithDBFile(d string) OptFunc {
	return func(opts *Opts) {
		opts.DBFile = d
	}
}

func NewConfig(opts ...OptFunc) (Config, error) {
	var err error

	c := Config{}
	o := defaultOpts()
	for _, fn := range opts {
		fn(&o)
	}
	c.Opts = o

	if err = c.setDataHome(); err != nil {
		return c, err
	}

	if err = c.setDBFile(); err != nil {
		return c, err
	}

	if err = c.setImportDir(); err != nil {
		return c, err
	}

	if err = c.setTmpDir(); err != nil {
		return c, err
	}

	return c, nil
}

func (c *Config) setDataHome() error {
	var dataHome string
	if c.DataHome != "" {
		return nil
	}
	xdgDataHome, ok := os.LookupEnv("XDG_DATA_HOME")
	if ok {
		dataHome = path.Join(xdgDataHome, AppName)
	} else {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		dataHome = path.Join(homeDir, ".local", "share", AppName)
	}
	c.DataHome = dataHome
	return nil
}

func (c *Config) setImportDir() error {
	if c.ImportDir == "" {
		c.ImportDir = path.Join(c.DataHome, "import")
	}
	return os.MkdirAll(c.ImportDir, os.ModePerm)
}

func (c *Config) setDBFile() error {
	if c.DBFile != "" {
		return os.MkdirAll(filepath.Dir(c.DBFile), os.ModePerm)
	}
	c.DBFile = path.Join(c.DataHome, "pennyapp.db")
	return os.MkdirAll(filepath.Dir(c.DBFile), os.ModePerm)
}

func (c *Config) setTmpDir() error {
	if c.TmpDir == "" {
		c.TmpDir = path.Join(c.DataHome, "tmp")
	}
	return os.MkdirAll(c.TmpDir, os.ModePerm)
}
