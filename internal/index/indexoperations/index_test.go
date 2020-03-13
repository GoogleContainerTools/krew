// Copyright 2019 The Kubernetes Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package indexoperations

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"

	"sigs.k8s.io/krew/internal/environment"
	"sigs.k8s.io/krew/internal/gitutil"
	"sigs.k8s.io/krew/internal/testutil"
)

func TestListIndexes(t *testing.T) {
	tmpDir, cleanup := testutil.NewTempDir(t)
	defer cleanup()

	wantIndexes := []Index{
		{
			Name: "custom",
			URL:  "https://github.com/custom/index.git",
		},
		{
			Name: "default",
			URL:  "https://github.com/default/index.git",
		},
	}

	for _, index := range wantIndexes {
		path := tmpDir.Path("index/" + index.Name)
		if err := os.MkdirAll(path, os.ModePerm); err != nil {
			t.Fatalf("cannot create directory %q: %s", filepath.Dir(path), err)
		}
		_, err := gitutil.Exec(path, "init")
		if err != nil {
			t.Fatalf("error initializing git repo: %s", err)
		}
		_, err = gitutil.Exec(path, "remote", "add", "origin", index.URL)
		if err != nil {
			t.Fatalf("error setting remote origin: %s", err)
		}
	}

	gotIndexes, err := ListIndexes(environment.NewPaths(tmpDir.Root()).IndexBase())
	if err != nil {
		t.Errorf("error listing indexes: %v", err)
	}
	if diff := cmp.Diff(wantIndexes, gotIndexes); diff != "" {
		t.Errorf("output does not match: %s", diff)
	}
}