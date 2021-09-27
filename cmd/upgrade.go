package cmd

import (
	"github.com/cic-sap/helmize/pkg"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"k8s.io/helm/pkg/helm"
	"os"
)

type upgradeCmd struct {
	release                  string
	chart                    string
	chartVersion             string
	chartRepo                string
	client                   helm.Interface
	detailedExitCode         bool
	devel                    bool
	disableValidation        bool
	disableOpenAPIValidation bool
	dryRun                   bool
	namespace                string // namespace to assume the release to be installed into. Defaults to the current kube config namespace.
	valueFiles               []string
	values                   []string
	stringValues             []string
	fileValues               []string
	reuseValues              bool
	resetValues              bool
	allowUnreleased          bool
	noHooks                  bool
	includeTests             bool
	suppressedKinds          []string
	outputContext            int
	showSecrets              bool
	postRenderer             string
	output                   string
	install                  *bool
	stripTrailingCR          bool
	normalizeManifests       bool
}

// upgradeCmd represents the upgrade command

func newChartCommand() *cobra.Command {

	if isHelm2() {
		panic(errors.New("No longer supports helm2"))
	}

	diff := upgradeCmd{}

	cmd := &cobra.Command{
		Use:   "upgrade",
		Short: "This command upgrades a release to a new version of a chart",
		Long: `The upgrade arguments must be a release and chart. The chart
argument can be either: a chart reference('example/mariadb'), a path to a chart directory,
a packaged chart, or a fully qualified URL. For chart references, the latest
version will be specified unless the '--version' flag is set.

To override values in a chart, use either the '--values' flag and pass in a file
or use the '--set' flag and pass configuration from the command line, to force string
values, use '--set-string'. In case a value is large and therefore
you want not to use neither '--values' nor '--set', use '--set-file' to read the
single large value from file.

You can specify the '--values'/'-f' flag multiple times. The priority will be given to the
last (right-most) file specified. For example, if both myvalues.yaml and override.yaml
contained a key called 'Test', the value set in override.yaml would take precedence:

    $ helm upgrade -f myvalues.yaml -f override.yaml redis ./redis

You can specify the '--set' flag multiple times. The priority will be given to the
last (right-most) set specified. For example, if both 'bar' and 'newbar' values are
set for a key called 'foo', the 'newbar' value would take precedence:

    $ helm upgrade --set foo=bar --set foo=newbar redis ./redis`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return pkg.Inject(args[0], args[1])
		},
	}

	f := cmd.Flags()
	var kubeconfig string
	f.StringVar(&kubeconfig, "kubeconfig", "", "This flag is ignored, to allow passing of this top level flag to helm")
	f.StringVar(&diff.chartVersion, "version", "", "specify the exact chart version to use. If this is not specified, the latest version is used")
	f.StringVar(&diff.chartRepo, "repo", "", "specify the chart repository url to locate the requested chart")
	f.BoolVar(&diff.detailedExitCode, "detailed-exitcode", false, "return a non-zero exit code when there are changes")
	f.BoolP("suppress-secrets", "q", false, "suppress secrets in the output")
	f.BoolVar(&diff.showSecrets, "show-secrets", false, "do not redact secret values in the output")
	f.StringArrayVar(&diff.valueFiles, "values", []string{}, "specify values in a YAML file (can specify multiple)")
	f.StringArrayVar(&diff.values, "set", []string{}, "set values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)")
	f.StringArrayVar(&diff.stringValues, "set-string", []string{}, "set STRING values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)")
	f.StringArrayVar(&diff.fileValues, "set-file", []string{}, "set values from respective files specified via the command line (can specify multiple or separate values with commas: key1=path1,key2=path2)")
	f.BoolVar(&diff.reuseValues, "reuse-values", false, "reuse the last release's values and merge in any new values. If '--reset-values' is specified, this is ignored")
	f.BoolVar(&diff.resetValues, "reset-values", false, "reset the values to the ones built into the chart and merge in any new values")
	f.BoolVar(&diff.allowUnreleased, "allow-unreleased", false, "enables diffing of releases that are not yet deployed via Helm")
	diff.install = f.BoolP("install", "i", false, "enables diffing of releases that are not yet deployed via Helm (equivalent to --allow-unreleased, added to match \"helm upgrade --install\" command")
	f.BoolVar(&diff.noHooks, "no-hooks", false, "disable diffing of hooks")
	f.BoolVar(&diff.includeTests, "include-tests", false, "enable the diffing of the helm test hooks")
	f.BoolVar(&diff.devel, "devel", false, "use development versions, too. Equivalent to version '>0.0.0-0'. If --version is set, this is ignored.")
	f.StringArrayVar(&diff.suppressedKinds, "suppress", []string{}, "allows suppression of the values listed in the diff output")
	f.IntVarP(&diff.outputContext, "context", "C", -1, "output NUM lines of context around changes")
	f.BoolVar(&diff.disableValidation, "disable-validation", false, "disables rendered templates validation against the Kubernetes cluster you are currently pointing to. This is the same validation performed on an install")
	f.BoolVar(&diff.disableOpenAPIValidation, "disable-openapi-validation", false, "disables rendered templates validation against the Kubernetes OpenAPI Schema")
	f.BoolVar(&diff.dryRun, "dry-run", false, "disables cluster access and show diff as if it was install. Implies --install, --reset-values, and --disable-validation")
	f.StringVar(&diff.postRenderer, "post-renderer", "", "the path to an executable to be used for post rendering. If it exists in $PATH, the binary will be used, otherwise it will try to look for the executable at the given path")
	f.StringVar(&diff.output, "output", "diff", "Possible values: diff, simple, json, template. When set to \"template\", use the env var HELM_DIFF_TPL to specify the template.")
	f.BoolVar(&diff.stripTrailingCR, "strip-trailing-cr", false, "strip trailing carriage return on input")
	f.BoolVar(&diff.normalizeManifests, "normalize-manifests", false, "normalize manifests before running diff to exclude style differences from the output")

	return cmd
}

func isHelm2() bool {
	return os.Getenv("TILLER_HOST") != ""
}

func init() {
	rootCmd.AddCommand(newChartCommand())

}
