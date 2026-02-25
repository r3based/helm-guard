package helm

import (
	"bytes"
	"fmt"
	"os/exec"
)

func Template(chartPath string, values []string) ([]byte, error) {
	args := []string{"template", "release", chartPath}

	for _, v := range values {
		args = append(args, "-f", v)
	}

	cmd := exec.Command("helm", args...)

	var out bytes.Buffer
	var stderr bytes.Buffer

	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("helm template failed: %s", stderr.String())
	}

	return out.Bytes(), nil
}
