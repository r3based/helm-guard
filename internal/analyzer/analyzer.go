package analyzer

import (
	"io"
	"os"
	"strings"

	"github.com/r3based/helm-guard/internal/helm"
	"github.com/r3based/helm-guard/internal/kube"
	"github.com/r3based/helm-guard/internal/model"
	"github.com/r3based/helm-guard/internal/profile"
	"github.com/r3based/helm-guard/internal/rules"
	"github.com/r3based/helm-guard/internal/rules/builtin"
)

// Config holds all inputs for a single analyze run.
type Config struct {
	ChartPath    string
	Release      string
	Namespace    string
	Values       []string
	Set          []string
	FailOn       string
	Disable      []string
	ProfileName  string
	ProfileFile  string   // path to profile YAML; used only if ProfileReader is nil
	ProfileReader io.Reader // if set, used instead of opening ProfileFile (e.g. for tests)
}

// Result is the output of Run.
type Result struct {
	Model             model.Model
	Findings          []rules.Finding
	RenderedCount     int
	EffectiveProfile  profile.EffectiveProfile
}

// Run renders the chart, builds the model, resolves the profile, runs rules, and returns the result.
// Caller is responsible for reporting (e.g. report.Pretty) and exit code (e.g. rules.ExitCode).
func Run(cfg Config) (Result, error) {
	rendered, err := helm.Template(helm.TemplateParams{
		ChartPath: cfg.ChartPath,
		Release:   cfg.Release,
		Namespace: cfg.Namespace,
		Values:    cfg.Values,
		Set:       cfg.Set,
	})
	if err != nil {
		return Result{}, err
	}

	objs, err := kube.ParseManifests(rendered)
	if err != nil {
		return Result{}, err
	}

	m := model.Build(objs)

	disableIDs := map[string]bool{}
	for _, x := range cfg.Disable {
		for _, id := range strings.Split(x, ",") {
			id = strings.TrimSpace(id)
			if id != "" {
				disableIDs[id] = true
			}
		}
	}

	allRules := builtin.All()
	builtinProfiles := profile.Builtin()

	var fileProfPtr *profile.Profile
	if cfg.ProfileReader != nil {
		pf, err := profile.LoadFromYAML(cfg.ProfileReader)
		if err != nil {
			return Result{}, err
		}
		fileProfPtr = &pf
	} else if cfg.ProfileFile != "" {
		f, err := os.Open(cfg.ProfileFile)
		if err != nil {
			return Result{}, err
		}
		defer f.Close()
		pf, err := profile.LoadFromYAML(f)
		if err != nil {
			return Result{}, err
		}
		fileProfPtr = &pf
	}

	resolvedProfile, err := profile.ResolveProfile(cfg.ProfileName, fileProfPtr, builtinProfiles)
	if err != nil {
		return Result{}, err
	}

	effective := profile.BuildEffectiveProfile(resolvedProfile, allRules)
	engineOpts := profile.ToEngineOptions(effective, disableIDs)

	engine := rules.New(allRules, engineOpts)
	findings := engine.Run(m)

	return Result{
		Model:            m,
		Findings:         findings,
		RenderedCount:    len(objs),
		EffectiveProfile: effective,
	}, nil
}
