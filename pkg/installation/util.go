// Copyright © 2018 Google Inc.
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

package installation

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/GoogleContainerTools/krew/pkg/index"
	"github.com/golang/glog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

// GetMatchingPlatform TODO(lbb)
func GetMatchingPlatform(i index.Plugin) (index.Platform, bool, error) {
	return matchPlatformToSystemEnvs(i, runtime.GOOS, runtime.GOARCH)
}

func matchPlatformToSystemEnvs(i index.Plugin, os, arch string) (index.Platform, bool, error) {
	envLabels := labels.Set{
		"os":   os,
		"arch": arch,
	}
	glog.V(2).Infof("Matching platform for labels(%v)", envLabels)
	for i, platform := range i.Spec.Platforms {
		sel, err := metav1.LabelSelectorAsSelector(platform.Selector)
		if err != nil {
			return index.Platform{}, false, fmt.Errorf("failed to compile label selector, err: %v", err)
		}
		if sel.Matches(envLabels) {
			glog.V(2).Infof("Found matching platform with index (%d)", i)
			return platform, true, nil
		}
	}
	return index.Platform{}, false, nil
}

func findInstalledPluginVersion(installPath, binDir, pluginName string) (name string, installed bool, err error) {
	if !index.IsSafePluginName(pluginName) {
		return "", false, fmt.Errorf("the plugin name %q is not allowed", pluginName)
	}
	glog.V(3).Infof("Searching for installed versions of %s in %q", pluginName, installPath)
	fis, err := ioutil.ReadDir(filepath.Join(installPath, pluginName))
	if os.IsNotExist(err) {
		return "", false, nil
	} else if err != nil {
		return "", false, fmt.Errorf("could not read direcory, err: %v", err)
	}
	for _, fi := range fis {
		if !fi.IsDir() {
			continue
		}
		if ok, err := containsPluginDescriptors(filepath.Join(installPath, pluginName, fi.Name())); err != nil {
			return "", false, fmt.Errorf("failed to find plugin descriptors in path %q, err: %v", filepath.Join(installPath, pluginName, fi.Name()), err)
		} else if ok {
			return fi.Name(), ok, nil
		}
	}
	return "", false, nil
}

// containsPluginDescriptors will recursively check a path if it contains any plugin descriptors that can be found by kubectl.
func containsPluginDescriptors(path string) (bool, error) {
	var contains bool
	glog.V(4).Infof("Checking path %q for plugin descriptors", path)
	return contains, filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if contains {
			return filepath.SkipDir
		}
		if info.Name() == "plugin.yaml" && !info.IsDir() {
			contains = true
			return filepath.SkipDir
		}
		return nil
	})
}

func getPluginVersion(p index.Platform, forceHEAD bool) (version, uri string, err error) {
	if (forceHEAD && p.Head != "") || (p.Head != "" && p.Sha256 == "" && p.URI == "") {
		return headVersion, p.Head, nil
	}
	if forceHEAD && p.Head == "" {
		return "", "", fmt.Errorf("can't force HEAD, with no HEAD specified")
	}
	return strings.ToLower(p.Sha256), p.URI, nil
}

func getDownloadTarget(index index.Plugin, forceHEAD bool) (version, uri string, fos []index.FileOperation, err error) {
	p, ok, err := GetMatchingPlatform(index)
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to get matching platforms, err: %v", err)
	}
	if !ok {
		return "", "", nil, fmt.Errorf("no matching platform found")
	}
	version, uri, err = getPluginVersion(p, forceHEAD)
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to get the plugin version, err: %v", err)
	}
	glog.V(4).Infof("Matching plugin version is %s", version)

	return version, uri, p.Files, nil
}

// ListInstalledPlugins returns a list of all name:version for all plugins.
func ListInstalledPlugins(installDir string) (map[string]string, error) {
	installed := make(map[string]string)
	plugins, err := ioutil.ReadDir(installDir)
	if err != nil {
		return installed, fmt.Errorf("failed to read install dir, err: %v", err)
	}
	for _, plugin := range plugins {
		version, ok, err := findInstalledPluginVersion(installDir, plugin.Name())
		if err != nil {
			return installed, fmt.Errorf("failed to get plugin version, err: %v", err)
		}
		if ok {
			installed[plugin.Name()] = version
		}
	}
	return installed, nil
}
