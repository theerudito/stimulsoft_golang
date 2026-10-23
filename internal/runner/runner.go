package runner

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

func Execute(ctx context.Context, path string, input io.Reader) ([]byte, error) {
	cmd := exec.CommandContext(ctx, path)
	cmd.Stdin = input
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); ctx.Err() != nil {
		return nil, ctx.Err()
	} else if err != nil {
		detail := strings.TrimSpace(stderr.String())
		if detail != "" {
			return nil, fmt.Errorf("stimulsoft runner failed: %w: %s", err, detail)
		}
		return nil, fmt.Errorf("stimulsoft runner failed: %w", err)
	}
	return stdout.Bytes(), nil
}
