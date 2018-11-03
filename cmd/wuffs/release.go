// Copyright 2018 The Wuffs Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	cf "github.com/google/wuffs/cmd/commonflags"
)

func genrelease(wuffsRoot string, langs []string, v cf.Version) error {
	revision := findRevision(wuffsRoot)
	for _, lang := range langs {
		suffix := lang
		if suffix == "c" {
			suffix = "h"
		}

		filename, contents, err := genreleaseLang(wuffsRoot, revision, v, lang, suffix)
		if err != nil {
			return err
		}
		if err := writeFile(filename, contents); err != nil {
			return err
		}
	}
	return nil
}

func genreleaseLang(wuffsRoot string, revision string, v cf.Version, lang string, suffix string) (filename string, contents []byte, err error) {
	// During a transitional period (starting in November 2018), remove any
	// existing wuffs-c generated files that use the old naming scheme (e.g.
	// "base.h" or "std/gif.h") instead of the new naming scheme (e.g.
	// "wuffs-base.h" or "wuffs-std-gif.h").
	if lang == "c" {
		baseDir := filepath.Join(wuffsRoot, "gen", lang)
		os.Remove(filepath.Join(baseDir, "base.h"))
		os.RemoveAll(filepath.Join(baseDir, "std"))
	}

	qualFilenames, err := findFiles(filepath.Join(wuffsRoot, "gen", lang), "."+suffix)
	if err != nil {
		return "", nil, err
	}

	command := "wuffs-" + lang
	args := []string(nil)
	args = append(args, "genrelease", "-revision", revision, "-version", v.String())
	args = append(args, qualFilenames...)
	stdout := &bytes.Buffer{}

	cmd := exec.Command(command, args...)
	cmd.Stdout = stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err == nil {
		// No-op.
	} else if _, ok := err.(*exec.ExitError); ok {
		return "", nil, fmt.Errorf("%s: failed", command)
	} else {
		return "", nil, err
	}

	base := "wuffs-unsupported-snapshot"
	if v.Major != 0 || v.Minor != 0 {
		base = fmt.Sprintf("wuffs-v%d.%d", v.Major, v.Minor)
	}
	ext := lang
	if ext == "c" {
		ext = "h"
	}
	return filepath.Join(wuffsRoot, "release", lang, base+"."+ext), stdout.Bytes(), nil
}

func findRevision(wuffsRoot string) string {
	// Assume that we're using git.

	head, err := ioutil.ReadFile(filepath.Join(wuffsRoot, ".git", "HEAD"))
	if err != nil {
		return ""
	}
	refPrefix := []byte("ref: ")
	if !bytes.HasPrefix(head, refPrefix) {
		return ""
	}
	head = head[len(refPrefix):]
	if len(head) == 0 || head[len(head)-1] != '\n' {
		return ""
	}
	head = head[:len(head)-1]

	ref, err := ioutil.ReadFile(filepath.Join(wuffsRoot, ".git", string(head)))
	if err != nil {
		return ""
	}
	if len(ref) == 0 || ref[len(ref)-1] != '\n' {
		return ""
	}
	ref = ref[:len(ref)-1]

	return string(ref)
}
