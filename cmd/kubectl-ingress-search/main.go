package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"regexp"

	"github.com/mozillazg/kubectl-ingress-search/pkg/ingress"
	"github.com/mozillazg/kubectl-ingress-search/pkg/process"
	"github.com/mozillazg/kubectl-ingress-search/pkg/render"
	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

type Option struct {
	caseSensitive              bool
	highlightDuplicateServices bool
	noHeader                   bool
	autoMergeTable             bool
	kubeconfig                 string
	showVersion                bool

	allNamespace bool
	namespace    string
	name         string
	service      string
	host         string
	path         string
	backend      string
	matchMode    string
}

var (
	opt                     = &Option{}
	matchModeRegexp         = "regexp"
	matchModeExact          = "exact"
	version                 = ""
	rootCmdDescriptionShort = "Search Ingress resources"
	rootCmdDescriptionLong  = `kubectl-ingress-search search Ingress resources to match given pattern(s).
It prints matched Ingress in table format.
More info: https://github.com/mozilazg/kubectl-ingress-search`

	rootCmdExamples = `
$ kubectl ingress search
$ kubectl ingress search --service test
$ kubectl ingress search --name test
$ kubectl ingress search --host example.com
$ kubectl ingress search --path /test
$ kubectl ingress search --no-header
$ kubectl ingress search --match-mode exact
$ kubectl ingress search --match-mode regexp
$ kubectl ingress search --case-sensitive
`
)

var rootCmd = &cobra.Command{
	Use:     "kubectl-ingress-search",
	Short:   rootCmdDescriptionShort,
	Long:    rootCmdDescriptionLong,
	Example: rootCmdExamples,
	Run:     run,
}

func init() {
	rootCmd.Flags().BoolVarP(&opt.showVersion, "version", "V", false, "show version and exit")
	rootCmd.Flags().StringVarP(&opt.matchMode, "match-mode", "m", "exact", "exact match or regexp match. One of (exact|regexp)")
	rootCmd.Flags().BoolVarP(&opt.caseSensitive, "case-sensitive", "I", false, "case sensitive pattern match")
	rootCmd.Flags().BoolVarP(&opt.highlightDuplicateServices, "highlight-duplicate-service", "H", false, "highlight duplicate service")
	rootCmd.Flags().BoolVar(&opt.noHeader, "no-header", false, "don't show table header")
	rootCmd.Flags().BoolVar(&opt.autoMergeTable, "auto-merge", false, "auto merge table cells")
	rootCmd.Flags().StringVarP(&opt.namespace, "namespace", "n", "default", "if present, the namespace scope for this CLI request")
	rootCmd.Flags().BoolVarP(&opt.allNamespace, "all-namespaces", "A", false, "if present, list the requested object(s) across all namespaces. Namespace in current context is ignored even if specified with --namespace")
	rootCmd.Flags().StringVar(&opt.name, "name", "", "search by name")
	rootCmd.Flags().StringVarP(&opt.service, "service", "s", "", "search by service")
	rootCmd.Flags().StringVarP(&opt.backend, "backend", "b", "", "search by backend")
	rootCmd.Flags().StringVar(&opt.host, "host", "", "search by host")
	rootCmd.Flags().StringVarP(&opt.path, "path", "p", "", "search by path")
	rootCmd.Flags().StringVar(&opt.kubeconfig, "kubeconfig", "", "path to the kubeconfig file to use for CLI requests")
}

func buildRegexp(value string) (*regexp.Regexp, error) {
	if opt.matchMode == matchModeExact {
		value = regexp.QuoteMeta(value)
	}
	if opt.caseSensitive {
		return regexp.Compile(value)
	}
	return regexp.Compile(`(?i)` + value)
}

func buildFilters() ([]process.Filter, error) {
	var filters []process.Filter
	if opt.name != "" {
		re, err := buildRegexp(opt.name)
		if err != nil {
			return nil, err
		}
		filters = append(filters, process.FieldValueFilter{Name: "name", Exp: re})
	}
	if opt.service != "" {
		exp := opt.service
		re, err := buildRegexp(exp)
		if err != nil {
			return nil, err
		}
		filters = append(filters, process.FieldValueFilter{Name: "backend", Exp: regexp.MustCompile(`^Service/`), NoColor: true})
		filters = append(filters, process.FieldValueFilter{Name: "backend", Exp: re})
	}
	if opt.host != "" {
		re, err := buildRegexp(opt.host)
		if err != nil {
			return nil, err
		}
		filters = append(filters, process.FieldValueFilter{Name: "host", Exp: re})
	}
	if opt.path != "" {
		re, err := buildRegexp(opt.path)
		if err != nil {
			return nil, err
		}
		filters = append(filters, process.FieldValueFilter{Name: "path", Exp: re})
	}
	if opt.backend != "" {
		re, err := buildRegexp(opt.backend)
		if err != nil {
			return nil, err
		}
		filters = append(filters, process.FieldValueFilter{Name: "backend", Exp: re})
	}
	if opt.highlightDuplicateServices {
		filters = append(filters, process.HighlightDupServiceFilter{})
	}
	return filters, nil
}

func run(cmd *cobra.Command, args []string) {
	if opt.showVersion {
		fmt.Println(version)
		return
	}
	namespace := opt.namespace
	if opt.allNamespace {
		namespace = ""
	}
	filters, err := buildFilters()
	if err != nil {
		exitWitError(err)
	}
	client, err := getClient()
	if err != nil {
		exitWitError(err)
	}
	s := ingress.NewSearcher(client)
	items, err := s.ListIngresses(context.Background(), namespace, v1.ListOptions{})
	if err != nil {
		exitWitError(err)
	}
	var rules []ingress.Rule
	for _, item := range items.Items {
		rules = append(rules, ingress.ParseRules(item)...)
	}
	for _, f := range filters {
		rules = f.Filter(rules)
	}

	r := render.TableRender{NoHeader: opt.noHeader, AutoMerge: opt.autoMergeTable}
	out := r.Render(rules)
	fmt.Println(out)
}

func getClient() (kubernetes.Interface, error) {
	if opt.kubeconfig != "" {
		flag.Parse()
	}
	conf, err := config.GetConfig()
	if err != nil {
		return nil, err
	}
	client := kubernetes.NewForConfigOrDie(conf)
	return client, nil
}

func exitWitError(err error) {
	fmt.Fprintf(os.Stderr, "%s\n", err)
	os.Exit(1)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		exitWitError(err)
	}
}
