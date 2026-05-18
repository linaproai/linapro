// This file handles command-line flag parsing for image-builder.

package main

import "flag"

// parseOptions reads flags and records which values were explicitly set.
func parseOptions() (cliOptions, map[string]bool) {
	opts := cliOptions{}
	flag.StringVar(&opts.ConfigPath, "config", "hack/config.yaml", "Repository tool config path")
	flag.BoolVar(&opts.BuildOnly, "build-only", false, "Prepare image build artifacts without running docker build")
	flag.BoolVar(&opts.Preflight, "preflight", false, "Validate image build request without checking artifacts or running docker build")
	flag.BoolVar(&opts.PrintBuildEnv, "print-build-env", false, "Print normalized build config as shell assignments")
	flag.StringVar(&opts.Image, "image", "", "Override image repository name")
	flag.StringVar(&opts.Tag, "tag", "", "Override image tag")
	flag.StringVar(&opts.Registry, "registry", "", "Override image registry prefix")
	flag.StringVar(&opts.Push, "push", "", "Override push behavior")
	flag.StringVar(&opts.Platforms, "platforms", "", "Override build target platforms")
	flag.StringVar(&opts.CGOEnabled, "cgo-enabled", "", "Override CGO build behavior")
	flag.StringVar(&opts.OutputDir, "output-dir", "", "Override build output directory")
	flag.StringVar(&opts.BinaryName, "binary-name", "", "Override build binary filename")
	flag.StringVar(&opts.BaseImage, "base-image", "", "Override Docker base image")
	flag.StringVar(&opts.Verbose, "verbose", "", "Show child command output")
	flag.Parse()

	specified := map[string]bool{}
	flag.Visit(func(item *flag.Flag) {
		specified[item.Name] = true
	})
	return opts, specified
}
