package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"text/template"
	"unicode"
	"unicode/utf8"

	"github.com/b5/outline/lib"
	bp "github.com/gohugoio/hugo/bufferpool"
	"github.com/spf13/cobra"
	parseutil "gopkg.in/src-d/go-parse-utils.v1"
)

// PackageCmd extracts and execute outline documents from a go package against a template",
var PackageCmd = &cobra.Command{
	Use:     "package",
	Aliases: []string{"pkg"},
	Short:   "exctract and execute outline documents from a go package against a template",
	Long:    ``,
	Run: func(cmd *cobra.Command, args []string) {
		funcMap := template.FuncMap{
			"split":          strings.Split,
			"trim":           strings.TrimSpace,
			"replace_all":    strings.ReplaceAll,
			"sanitizeAnchor": sanitizeAnchor,
		}

		t := template.Must(template.New("mdIndex").Funcs(funcMap).Parse(mdIndex))

		str, err := cmd.Flags().GetString("template")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		ctx, err := cmd.Flags().GetString("context-dir")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		if str != "" {
			b, err := ioutil.ReadFile(str)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			t, err = template.New("mdIndex").Funcs(funcMap).Parse(string(b))
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}

		docs := map[string]*lib.Doc{}
		for _, pkg := range args {
			pkg, err := parseutil.PackageAST(pkg)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			for _, f := range pkg.Files {
				for _, c := range f.Comments {
					buf := strings.NewReader(c.Text())
					read, err := lib.Parse(buf)
					if err != nil {
						fmt.Println(err.Error())
						os.Exit(1)
					}

					for _, doc := range read {
						if found, ok := docs[doc.Name]; ok {
							merge(found, doc)
							continue
						}

						docs[doc.Name] = doc
					}
				}
			}
		}

		noSort, err := cmd.Flags().GetBool("no-sort")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		var list lib.Docs
		for _, doc := range docs {
			list = append(list, doc)
		}

		if !noSort {
			list.Sort()
		}

		if ctx != "" {
			os.Chdir(ctx)
		}

		if err := t.Execute(os.Stdout, list); err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
	},
}

func merge(a, b *lib.Doc) {
	if a.Description == "" {
		a.Description = b.Description
	}

	if a.Path == "" {
		a.Path = b.Path
	}

	a.Types = append(a.Types, b.Types...)
	a.Functions = append(a.Functions, b.Functions...)
}

func init() {
	PackageCmd.Flags().StringP("template", "t", "", "template file to load. overrides preset")
	PackageCmd.Flags().Bool("no-sort", false, "don't alpha-sort fields & outline documents")
	PackageCmd.Flags().StringP("context-dir", "d", "", "context dir")
}

func sanitizeAnchor(input string) string {
	b := []byte(input)
	buf := bp.GetBuffer()

	for len(b) > 0 {
		r, size := utf8.DecodeRune(b)
		switch {
		case r == '-' || r == ' ':
			buf.WriteRune('-')
		case isAlphaNumeric(r):
			buf.WriteRune(unicode.ToLower(r))
		default:
		}

		b = b[size:]
	}

	result := make([]byte, buf.Len())
	copy(result, buf.Bytes())

	bp.PutBuffer(buf)

	return string(result)
}

func isAlphaNumeric(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}
