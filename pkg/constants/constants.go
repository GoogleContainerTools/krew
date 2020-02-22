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

package constants

const (
	CurrentAPIVersion = "krew.googlecontainertools.github.com/v1alpha2"
	PluginKind        = "Plugin"
	ManifestExtension = ".yaml"
	KrewPluginName    = "krew" // plugin name of krew itself

	// IndexURI points to the upstream index.
	IndexURI = "https://github.com/kubernetes-sigs/krew-index.git"
	// DefaultIndexName is a magic string that's used for a plugin name specified without an index.
	DefaultIndexName = "default"
	// EnableMultiIndexFlag is the name of the environment variable that needs to be set to use
	// the features around multiple indexes (this will be removed later on).
	EnableMultiIndexFlag = "X_KREW_ENABLE_MULTI_INDEX"
)
