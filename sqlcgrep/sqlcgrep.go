// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package sqlcgrep provides functionality for replace/regenerate sqlc code.
package sqlcgrep

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/danil/ggrep/regenerate"
	"github.com/goccy/go-yaml"
	"golang.org/x/tools/imports"
)

func New(yml []byte, opts ...Option) error {
	var cfg Configuration

	err := yaml.Unmarshal(yml, &cfg)
	if err != nil {
		return fmt.Errorf("unmarshal yaml: %w", err)
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	for _, pkg := range cfg.Packages {
		for _, replace := range pkg.Regenerates {
			f, err := os.OpenFile(filepath.Join(cfg.directory, replace.File), os.O_RDWR, os.ModePerm)
			if os.IsNotExist(err) {
				fmt.Fprintf(os.Stderr, "File %q not exist.\n", replace.File)
				continue

			} else if err != nil {
				return fmt.Errorf("regexp compile: %w", err)
			}

			var regenerates []regenerate.Option

			for _, rep := range replace.Strings {
				regenerates = append(regenerates,
					regenerate.ReplaceString(rep.Match, rep.Replacement),
				)
			}

			for _, rep := range replace.Regexps {
				re, err := regexp.Compile(rep.Match)
				if err != nil {
					return fmt.Errorf("regexp compile: %w", err)
				}

				regenerates = append(regenerates,
					regenerate.ReplaceRegexp(re, rep.Replacement),
				)
			}

			if cfg.gofmt != nil {
				regenerates = append(regenerates, regenerate.WithGofmt(cfg.gofmt))
			}

			err = regenerate.Pipe(f, f, regenerates...)
			if err != nil {
				return fmt.Errorf("regenerate pipe: %s", err)
			}
		}
	}

	return nil
}

type Configuration struct {
	Packages  []Package `yaml:"packages"`
	directory string
	gofmt     *imports.Options
}

type Package struct {
	Regenerates []Replace `yaml:"regenerates"`
}

type Replace struct {
	File    string   `yaml:"file"`
	Strings []String `yaml:"strings"`
	Regexps []Regexp `yaml:"regexps"`
}

type String struct {
	Match       string `yaml:"match"`
	Replacement string `yaml:"replacement"`
}

type Regexp struct {
	Match       string `yaml:"match"`
	Replacement string `yaml:"replacement"`
}

// Option changes configuration.
type Option func(*Configuration)

// WithDirectory sets directory.
func WithDirectory(dir string) Option {
	return func(c *Configuration) { c.directory = dir }
}

func WithGofmt(opts *imports.Options) Option {
	return func(c *Configuration) { c.gofmt = opts }
}
