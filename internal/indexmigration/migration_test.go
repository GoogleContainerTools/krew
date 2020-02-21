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

package indexmigration

import (
	"os"
	"testing"

	"sigs.k8s.io/krew/internal/environment"
	"sigs.k8s.io/krew/internal/testutil"
)

func TestIsMigrated(t *testing.T) {
	if _, ok := os.LookupEnv("X_KREW_ENABLE_MULTI_INDEX"); !ok {
		t.Skip("Set X_KREW_ENABLE_MULTI_INDEX variable to run this test")
	}

	tests := []struct {
		name     string
		dirPath  string
		expected bool
	}{
		{
			name:     "Already migrated",
			dirPath:  "index/default",
			expected: true,
		},
		{
			name:     "Not migrated",
			dirPath:  "index",
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tmpDir, cleanup := testutil.NewTempDir(t)
			defer cleanup()

			_ = os.MkdirAll(tmpDir.Path(test.dirPath), os.ModePerm)

			newPaths := environment.NewPaths(tmpDir.Root())
			actual, err := Done(newPaths)
			if err != nil {
				t.Fatal(err)
			}
			if actual != test.expected {
				t.Errorf("Expected %v but found %v", test.expected, actual)
			}
		})
	}
}
