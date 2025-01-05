package config

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"

	"github.com/0wnperception/go-helpers/pkg/checks"
	"github.com/0wnperception/go-helpers/pkg/slice"
	"gopkg.in/yaml.v3"
)

var (
	ErrNilConfig                 = errors.New("config object could not be nil")
	ErrConfigObjectMustBePointer = errors.New("config object must be pointer to struct")
)

type FileNotFoundError struct {
	name      string
	locations string
}

func (f FileNotFoundError) Error() string {
	return fmt.Sprintf("Reader File %q Not Found in %q", f.name, f.locations)
}

type Reader interface {
	Read(cfg any) error
}

type cfg struct {
	r           io.Reader
	configFile  string
	configPaths []string
	envPrefix   string
	rawBytes    []byte
}

type Option func(c *cfg)

func WithEncPrefix(prefix string) Option {
	return func(c *cfg) {
		c.envPrefix = prefix
	}
}

func WithReader(r io.Reader) Option {
	return func(c *cfg) {
		c.r = r
	}
}

func WithConfigFile(f string) Option {
	return func(c *cfg) {
		c.configFile = f
	}
}

func WithConfigPath(p string) Option {
	return func(c *cfg) {
		if p != "" {
			abspath := absPathify(p)
			if !stringInSlice(abspath, c.configPaths) {
				c.configPaths = append(c.configPaths, abspath)
			}
		}
	}
}

func New(ops ...Option) Reader {
	c := new(cfg)

	c.configFile = "config.yaml"

	for _, o := range ops {
		o(c)
	}

	return c
}

func (c *cfg) readConfigFile() error {
	if c.configFile != "" {
		fn, err := c.findConfigFile()
		if err != nil {
			return err
		}

		in, err := os.ReadFile(fn)
		if err != nil {
			return fmt.Errorf("read config file error: %w", err)
		}

		c.rawBytes = in
	}

	return nil
}

func (c *cfg) readConfig(r io.Reader) error {
	buf := new(bytes.Buffer)

	if _, err := buf.ReadFrom(r); err != nil {
		return fmt.Errorf("read config error: %w", err)
	}

	c.rawBytes = buf.Bytes()

	return nil
}

func (c *cfg) Read(cfg any) error {
	if checks.IsNil(cfg) {
		return ErrNilConfig
	}

	if reflect.TypeOf(cfg).Kind() != reflect.Ptr {
		return ErrConfigObjectMustBePointer
	}

	if reflect.ValueOf(cfg).Elem().Type().Kind() != reflect.Struct {
		return ErrConfigObjectMustBePointer
	}

	var err error

	if c.r != nil {
		err = c.readConfig(c.r)
	} else {
		err = c.readConfigFile()
	}

	if err != nil {
		return fmt.Errorf("read config error: %w", err)
	}

	if err = SetDefaults(cfg); err != nil {
		return fmt.Errorf("set defaults error: %w", err)
	}

	if err = yaml.Unmarshal(c.rawBytes, cfg); err != nil {
		return fmt.Errorf("parse config error: %w", err)
	}

	if err = InjectFromEnv(c.envPrefix, cfg); err != nil {
		return fmt.Errorf("overwrite from env variables error: %w", err)
	}

	return nil
}

func userHomeDir() string {
	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}

		return home
	}

	return os.Getenv("HOME")
}

func stringInSlice(a string, list []string) bool {
	return slice.Contains(a, list)
}

func (c *cfg) findConfigFile() (string, error) {
	if len(c.configPaths) == 0 {
		if _, err := os.Stat(c.configFile); !os.IsNotExist(err) {
			return c.configFile, nil
		}
	} else {
		for _, cp := range c.configPaths {
			fn := filepath.Join(cp, c.configFile)

			if _, err := os.Stat(fn); !os.IsNotExist(err) {
				return fn, nil
			}
		}
	}

	return "", FileNotFoundError{name: c.configFile, locations: fmt.Sprintf("%s", c.configPaths)}
}

func absPathify(inPath string) string {
	if strings.HasPrefix(inPath, "$HOME") {
		inPath = userHomeDir() + inPath[5:]
	}

	if strings.HasPrefix(inPath, "$") {
		end := strings.Index(inPath, string(os.PathSeparator))
		inPath = os.Getenv(inPath[1:end]) + inPath[end:]
	}

	if filepath.IsAbs(inPath) {
		return filepath.Clean(inPath)
	}

	p, err := filepath.Abs(inPath)
	if err == nil {
		return filepath.Clean(p)
	}

	return ""
}
