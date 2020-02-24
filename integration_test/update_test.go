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

package integrationtest

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"sigs.k8s.io/krew/internal/environment"
	"sigs.k8s.io/krew/pkg/constants"
)

func TestKrewUpdate(t *testing.T) {
	skipShort(t)

	test, cleanup := NewTest(t)
	defer cleanup()

	// nb do not call WithIndex() here
	updateOut := string(test.Krew("update").RunOrFailOutput())
	if strings.Contains(updateOut, "New plugins available:") {
		t.Fatalf("clean index fetch should not show 'new plugins available': %s", updateOut)
	}
	plugins := lines(test.Krew("search").RunOrFailOutput())
	if len(plugins) < 10 {
		// the first line is the header
		t.Errorf("Less than %d plugins found, `krew update` most likely failed unless TestKrewSearchAll also failed", len(plugins)-1)
	}

	if err := test.Krew("update").Run(); err != nil {
		t.Fatal("re-run of 'update' must succeed")
	}
}

func TestKrewUpdateListsNewPlugins(t *testing.T) {
	skipShort(t)
	test, cleanup := NewTest(t)
	defer cleanup()

	test = test.WithIndex()

	pluginManifest := filepath.Join(environment.NewPaths(test.Root()).IndexPluginsPath("default"), validPlugin+constants.ManifestExtension)
	if err := os.Remove(pluginManifest); err != nil {
		t.Fatalf("failed to delete manifest of an existing plugin: %v", err)
	}

	out := string(test.Krew("update").RunOrFailOutput())
	if !strings.Contains(out, "New plugins available:") {
		t.Fatalf("output doesn't list new plugins available; output=%s", out)
	}
	if !strings.Contains(out, validPlugin) {
		t.Fatalf("output doesn't list the new plugin (%s) is available: %s", validPlugin, out)
	}
}

func TestKrewUpdateListsUpgradesAvailable(t *testing.T) {
	skipShort(t)

	test, cleanup := NewTest(t)
	defer cleanup()
	test = test.WithIndex()

	// set version of some manifests to v0.0.0
	pluginManifest := filepath.Join(environment.NewPaths(test.Root()).IndexPluginsPath("default"), validPlugin+constants.ManifestExtension)
	modifyManifestVersion(t, pluginManifest, "v0.0.0")

	test.Krew("install", validPlugin, "--no-update-index").RunOrFail()  // has updates available
	test.Krew("install", validPlugin2, "--no-update-index").RunOrFail() // no updates available

	out := string(test.Krew("update").RunOrFailOutput())
	if !strings.Contains(out, "Upgrades available for installed plugins:") {
		t.Fatalf("output doesn't list upgrades available; output=%s", out)
	}
	if !strings.Contains(out, validPlugin+" v") {
		t.Fatalf("output doesn't mention update available for %q; output=%s", validPlugin, out)
	}
	if strings.Contains(out, validPlugin2+" v") {
		t.Fatalf("output should not mention update available for %q; output=%s", validPlugin2, out)
	}
}

func modifyManifestVersion(t *testing.T, file, version string) {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	r := regexp.MustCompile(`(?m)(\bversion:\s)(.*)$`) // patch "version:" field
	b = r.ReplaceAll(b, []byte(fmt.Sprintf("${1}%s", version)))
	if err := ioutil.WriteFile(file, b, 0); err != nil {
		t.Fatal(err)
	}
}
