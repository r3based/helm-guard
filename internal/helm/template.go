package helm

import (
	"bytes"
	"fmt"
	"os/exec"
)

type TemplateParams struct {
	ChartPath string
	Release   string
	Namespace string
	Values    []string
	Set       []string
}

func Template(p TemplateParams) ([]byte, error) {
	if p.Release == "" {
		p.Release = "release"
	}

	args := []string{"template", p.Release, p.ChartPath}

	if p.Namespace != "" {
		args = append(args, "--namespace", p.Namespace)
	}

	for _, vf := range p.Values {
		args = append(args, "-f", vf)
	}

	for _, s := range p.Set {
		args = append(args, "--set", s)
	}

	cmd := exec.Command("helm", args...)

	var out bytes.Buffer
	var errBuf bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errBuf

	if err := cmd.Run(); err != nil {
		if errBuf.Len() > 0 {
			return nil, fmt.Errorf("helm template failed: %s", errBuf.String())
		}
		return nil, fmt.Errorf("helm template failed: %w", err)
	}

	return out.Bytes(), nil
}
