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

package cmd

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/kubectl/pkg/util/slice"

	"sigs.k8s.io/krew/pkg/constants"
	"sigs.k8s.io/krew/pkg/installation"
)

type Entry struct {
	Name    string
	Version string
}

type Entries []Entry

func sortByName(e Entries) Entries {
	sort.Slice(e, func(a, b int) bool {
		return e[a].Name < e[b].Name
	})
	return e
}

// Consume produces a junk GroupVersionKind for obj.GetObjectKind().GroupVersionKind().Empty() check to eat in PrintObj()
type Consume struct {
	APIVersion string
	Kind       string
}

func (c Consume) SetGroupVersionKind(kind schema.GroupVersionKind) {
	c.APIVersion, c.Kind = kind.ToAPIVersionAndKind()
}

func (c Consume) GroupVersionKind() schema.GroupVersionKind {
	return schema.FromAPIVersionAndKind(c.APIVersion, c.Kind)
}

func (e Entries) GetObjectKind() schema.ObjectKind {
	return Consume{constants.CurrentAPIVersion, constants.PluginKind}
}

func (e Entry) GetObjectKind() schema.ObjectKind {
	return Consume{constants.CurrentAPIVersion, constants.PluginKind}
}

func (e Entries) DeepCopyObject() runtime.Object {
	return append(Entries{}, e...)
}

func (e Entry) DeepCopyObject() runtime.Object {
	return Entry{e.Name, e.Version}
}

type ListFlags struct {
	JSONYamlPrintFlags *genericclioptions.JSONYamlPrintFlags
	OutputFormat       *string
}

func NewListFlags() *ListFlags {
	outputFormat := ""
	return &ListFlags{
		JSONYamlPrintFlags: genericclioptions.NewJSONYamlPrintFlags(),
		OutputFormat:       &outputFormat,
	}
}

func (f *ListFlags) AllowedFormats() []string {
	return f.JSONYamlPrintFlags.AllowedFormats()
}

func (f *ListFlags) ToPrinter() (printers.ResourcePrinter, error) {
	outputFormat := *f.OutputFormat
	var printer printers.ResourcePrinter

	switch outputFormat {
	case "json":
		printer = &printers.JSONPrinter{}
	case "yaml":
		printer = &printers.YAMLPrinter{}
	default:
		return nil, genericclioptions.NoCompatiblePrinterError{OutputFormat: f.OutputFormat, AllowedFormats: f.AllowedFormats()}
	}

	return printer, nil
}

func init() {
	outputFormat := ""
	output := NewListFlags()

	// listCmd represents the list command
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List installed kubectl plugins",
		Long: `Show a list of installed kubectl plugins and their versions.

Remarks:
  Redirecting the output of this command to a program or file will only print
  the names of the plugins installed. This output can be piped back to the
  "install" command.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			plugins, err := installation.ListInstalledPlugins(paths.InstallReceiptsPath())
			if err != nil {
				return errors.Wrap(err, "failed to find all installed versions")
			}

			// return sorted list of plugin names when piped to other commands or file
			if *output.OutputFormat == "name" || (*output.OutputFormat == "" && !isTerminal(os.Stdout)) {
				var names []string
				for name := range plugins {
					names = append(names, name)
				}
				sort.Strings(names)
				fmt.Fprintln(os.Stdout, strings.Join(names, "\n"))
				return nil
			}

			if *output.OutputFormat == "wide" || (*output.OutputFormat == "" && isTerminal(os.Stdout)) {
				// print table
				var rows [][]string
				for p, version := range plugins {
					rows = append(rows, []string{p, version})
				}
				rows = sortByFirstColumn(rows)
				return printTable(os.Stdout, []string{"PLUGIN", "VERSION"}, rows)
			}

			if slice.ContainsString(output.AllowedFormats(), *output.OutputFormat, nil) {
				objs := make(Entries, 0, len(plugins))
				for plugin, version := range plugins {
					obj := Entry{plugin, version}
					objs = append(objs, obj)
				}
				objs = sortByName(objs)
				p, err := output.ToPrinter()
				if err != nil {
					return err
				}
				err = p.PrintObj(objs, os.Stdout)
				if err != nil {
					return err
				}
				return nil
			}

			return errors.New("unsupported output format specified")
		},
		PreRunE: checkIndex,
	}

	listCmd.Flags().StringVarP(&outputFormat, "output", "o", "", "Output format. One of json, yaml, wide, name")

	output.OutputFormat = &outputFormat
	rootCmd.AddCommand(listCmd)
}

func printTable(out io.Writer, columns []string, rows [][]string) error {
	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	fmt.Fprint(w, strings.Join(columns, "\t"))
	fmt.Fprintln(w)
	for _, values := range rows {
		fmt.Fprint(w, strings.Join(values, "\t"))
		fmt.Fprintln(w)
	}
	return w.Flush()
}

func sortByFirstColumn(rows [][]string) [][]string {
	sort.Slice(rows, func(a, b int) bool {
		return rows[a][0] < rows[b][0]
	})
	return rows
}
