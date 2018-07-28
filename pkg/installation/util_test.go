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
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	"github.com/GoogleContainerTools/krew/pkg/index"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_matchPlatformToSystemEnvs(t *testing.T) {
	matchingPlatform := index.Platform{
		Head: "A",
		Selector: &v1.LabelSelector{
			MatchLabels: map[string]string{
				"os": "foo",
			},
		},
		Files: nil,
	}

	type args struct {
		i index.Plugin
	}
	tests := []struct {
		name         string
		args         args
		wantPlatform index.Platform
		wantFound    bool
		wantErr      bool
	}{
		{
			name: "Test Matching Index",
			args: args{
				i: index.Plugin{
					Spec: index.PluginSpec{
						Platforms: []index.Platform{
							matchingPlatform, {
								Head: "B",
								Selector: &v1.LabelSelector{
									MatchLabels: map[string]string{
										"os": "None",
									},
								},
							},
						},
					},
				},
			},
			wantPlatform: matchingPlatform,
			wantFound:    true,
			wantErr:      false,
		}, {
			name: "Test Matching Index Not Found",
			args: args{
				i: index.Plugin{
					Spec: index.PluginSpec{
						Platforms: []index.Platform{
							{
								Head: "B",
								Selector: &v1.LabelSelector{
									MatchLabels: map[string]string{
										"os": "None",
									},
								},
							},
						},
					},
				},
			},
			wantPlatform: index.Platform{},
			wantFound:    false,
			wantErr:      false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPlatform, gotFound, err := matchPlatformToSystemEnvs(tt.args.i, "foo", "amdBar")
			if (err != nil) != tt.wantErr {
				t.Errorf("GetMatchingPlatform() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotPlatform, tt.wantPlatform) {
				t.Errorf("GetMatchingPlatform() gotPlatform = %v, want %v", gotPlatform, tt.wantPlatform)
			}
			if gotFound != tt.wantFound {
				t.Errorf("GetMatchingPlatform() gotFound = %v, want %v", gotFound, tt.wantFound)
			}
		})
	}
}

func Test_getPluginVersion(t *testing.T) {
	type args struct {
		p         index.Platform
		forceHEAD bool
	}
	tests := []struct {
		name        string
		args        args
		wantVersion string
		wantURI     string
		wantErr     bool
	}{
		{
			name: "Get Single Head",
			args: args{
				p: index.Platform{
					Head:   "https://head.git",
					URI:    "",
					Sha256: "",
				},
				forceHEAD: false,
			},
			wantVersion: "HEAD",
			wantURI:     "https://head.git",
		}, {
			name: "Get URI default",
			args: args{
				p: index.Platform{
					Head:   "https://head.git",
					URI:    "https://uri.git",
					Sha256: "deadbeef",
				},
				forceHEAD: false,
			},
			wantVersion: "deadbeef",
			wantURI:     "https://uri.git",
		}, {
			name: "Get HEAD force",
			args: args{
				p: index.Platform{
					Head:   "https://head.git",
					URI:    "https://uri.git",
					Sha256: "deadbeef",
				},
				forceHEAD: true,
			},
			wantVersion: "HEAD",
			wantURI:     "https://head.git",
		}, {
			name: "HEAD force fallback",
			args: args{
				p: index.Platform{
					Head:   "",
					URI:    "https://uri.git",
					Sha256: "deadbeef",
				},
				forceHEAD: true,
			},
			wantErr:     true,
			wantVersion: "",
			wantURI:     "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotVersion, gotURI, err := getPluginVersion(tt.args.p, tt.args.forceHEAD)
			if (err != nil) != tt.wantErr {
				t.Errorf("getPluginVersion() gotVersion = %v, want %v, got err = %v want err = %v", gotVersion, tt.wantVersion, err, tt.wantErr)
			}
			if gotVersion != tt.wantVersion {
				t.Errorf("getPluginVersion() gotVersion = %v, want %v", gotVersion, tt.wantVersion)
			}
			if gotURI != tt.wantURI {
				t.Errorf("getPluginVersion() gotURI = %v, want %v", gotURI, tt.wantURI)
			}
		})
	}
}

func Test_getDownloadTarget(t *testing.T) {
	matchingPlatform := index.Platform{
		Head:   "https://head.git",
		URI:    "https://uri.git",
		Sha256: "deadbeef",
		Selector: &v1.LabelSelector{
			MatchLabels: map[string]string{
				"os": runtime.GOOS,
			},
		},
		Files: nil,
	}
	type args struct {
		index     index.Plugin
		forceHEAD bool
	}
	tests := []struct {
		name        string
		args        args
		wantVersion string
		wantURI     string
		wantFos     []index.FileOperation
		wantErr     bool
	}{
		{
			name: "Find Matching Platform",
			args: args{
				forceHEAD: true,
				index: index.Plugin{
					Spec: index.PluginSpec{
						Platforms: []index.Platform{
							matchingPlatform,
							{
								Head: "https://wrong.com",
								Selector: &v1.LabelSelector{
									MatchLabels: map[string]string{
										"os": "None",
									},
								},
							},
						},
					},
				},
			},
			wantVersion: "HEAD",
			wantURI:     "https://head.git",
			wantFos:     nil,
			wantErr:     false,
		}, {
			name: "No Matching Platform",
			args: args{
				forceHEAD: true,
				index: index.Plugin{
					Spec: index.PluginSpec{
						Platforms: []index.Platform{
							{
								Head: "https://wrong.com",
								Selector: &v1.LabelSelector{
									MatchLabels: map[string]string{
										"os": "None",
									},
								},
							},
						},
					},
				},
			},
			wantVersion: "",
			wantURI:     "",
			wantFos:     nil,
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotVersion, gotURI, gotFos, err := getDownloadTarget(tt.args.index, tt.args.forceHEAD)
			if (err != nil) != tt.wantErr {
				t.Errorf("getDownloadTarget() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotVersion != tt.wantVersion {
				t.Errorf("getDownloadTarget() gotVersion = %v, want %v", gotVersion, tt.wantVersion)
			}
			if gotURI != tt.wantURI {
				t.Errorf("getDownloadTarget() gotURI = %v, want %v", gotURI, tt.wantURI)
			}
			if !reflect.DeepEqual(gotFos, tt.wantFos) {
				t.Errorf("getDownloadTarget() gotFos = %v, want %v", gotFos, tt.wantFos)
			}
		})
	}
}

func Test_findInstalledPluginVersion(t *testing.T) {
	type args struct {
		installPath string
		pluginName  string
	}
	tests := []struct {
		name          string
		args          args
		wantName      string
		wantInstalled bool
		wantErr       bool
	}{
		{
			name: "Find version",
			args: args{
				installPath: filepath.Join(testdataPath(t), "index"),
				pluginName:  "foo",
			},
			wantName:      "deadbeef",
			wantInstalled: true,
			wantErr:       false,
		}, {
			name: "No installed version",
			args: args{
				installPath: filepath.Join(testdataPath(t), "index"),
				pluginName:  "not-found",
			},
			wantName:      "",
			wantInstalled: false,
			wantErr:       false,
		}, {
			name: "Insecure name",
			args: args{
				installPath: filepath.Join(testdataPath(t), "index"),
				pluginName:  "../foo",
			},
			wantName:      "",
			wantInstalled: false,
			wantErr:       true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotName, gotInstalled, err := findInstalledPluginVersion(tt.args.installPath, tt.args.pluginName)
			if (err != nil) != tt.wantErr {
				t.Errorf("getOtherInstalledVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotName != tt.wantName {
				t.Errorf("getOtherInstalledVersion() gotName = %v, want %v", gotName, tt.wantName)
			}
			if gotInstalled != tt.wantInstalled {
				t.Errorf("getOtherInstalledVersion() gotInstalled = %v, want %v", gotInstalled, tt.wantInstalled)
			}
		})
	}
}

func Test_containsPluginDescriptors(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "test recursive contains plugin descriptor",
			args: args{
				path: filepath.Join(testdataPath(t), "index", "foo"),
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "test directly contains plugin descriptor",
			args: args{
				path: filepath.Join(testdataPath(t), "index", "foo", "deadbeef"),
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "test no plugin descriptor",
			args: args{
				path: filepath.Join(testdataPath(t), "index", "foo", "AAAnotplugin"),
			},
			want:    false,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := containsPluginDescriptors(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("containsPluginDescriptors() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("containsPluginDescriptors() = %v, want %v", got, tt.want)
			}
		})
	}
}

func testdataPath(t *testing.T) string {
	pwd, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Join(pwd, "testdata")
}
