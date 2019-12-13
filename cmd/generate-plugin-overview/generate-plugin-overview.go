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

// plugin-overview reads the manifests in a directory and creates a markdown overview page
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"

	"sigs.k8s.io/krew/internal/index/indexscanner"
	"sigs.k8s.io/krew/pkg/index"
)

const (
	separator  = " | "
	pageHeader = `## Available kubectl plugins

To install these kubectl plugins:

1. [Install Krew](https://github.com/kubernetes-sigs/krew#installation)
2. Run ` + "`kubectl krew install PLUGIN_NAME`" + ` to install a plugin via Krew.

The following kubectl plugins are currently available on
[Krew plugin index](https://sigs.k8s.io/krew-index). Note that this table may be
outdated. For the most up-to-date list of plugins, visit the
[krew-index](https://github.com/kubernetes-sigs/krew-index/tree/master/plugins)
repository or run <code>kubectl krew search</code>.
`

	pageFooter = `

---

_This page is generated by running the
[generate-plugin-overview](http://sigs.k8s.io/krew/cmd/generate-plugin-overview)
tool._
`
)

var (
	githubRepoPattern = regexp.MustCompile(`.*github\.com/([^/]+/[^/#]+)`)
)

func main() {
	pluginsDir := flag.String("plugins-dir", "", "The directory containing the plugin manifests")
	flag.Parse()

	if *pluginsDir == "" {
		flag.Usage()
		return
	}

	plugins, err := indexscanner.LoadPluginListFromFS(*pluginsDir)
	if err != nil {
		log.Fatal(err)
	}

	out := os.Stdout

	_, _ = fmt.Fprintln(out, pageHeader)

	printTableHeader(out)
	for _, p := range plugins {
		printTableRowForPlugin(out, &p)
	}

	_, _ = fmt.Fprintln(out, pageFooter)
}

func printTableHeader(out io.Writer) {
	printRow(out, "Name", "Description", "Stars")
	printRow(out, "----", "-----------", "-----")
}

func printTableRowForPlugin(out io.Writer, p *index.Plugin) {
	// 1st column
	name := p.Name
	if homepage := p.Spec.Homepage; homepage != "" {
		name = fmt.Sprintf("[%s](%s)", strings.TrimSpace(name), homepage)
	}

	// 2nd column
	description := strings.TrimSpace(p.Spec.ShortDescription)

	// 3rd column
	shield := makeGithubShield(p.Spec.Homepage)

	printRow(out, name, description, shield)
}

func makeGithubShield(homePage string) string {
	repo := findRepo(homePage)
	if repo == "" {
		return ""
	}
	return "![GitHub stars](https://img.shields.io/github/stars/" + repo + ".svg?label=stars&logo=github)"
}

func findRepo(homePage string) string {
	if matches := githubRepoPattern.FindStringSubmatch(homePage); matches != nil {
		return matches[1]
	}

	knownHomePages := map[string]string{
		`https://sigs.k8s.io/krew`:                                   "kubernetes-sigs/krew",
		`https://kubernetes.github.io/ingress-nginx/kubectl-plugin/`: "kubernetes/ingress-nginx",
		`https://kudo.dev/`:                                          "kudobuilder/kudo",
	}
	return knownHomePages[homePage]
}

func printRow(w io.Writer, cols ...string) {
	_, _ = fmt.Fprintln(w, strings.Join(cols, separator))
}
