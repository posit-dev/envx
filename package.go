package envx

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/posit-dev/envx/internal"
)

const (
	usageString = `Usage: envx <command> [command arguments]

EXpand the Env and eXec the command!

Environment variables with known 'expansions' will be 'expanded' and
then passed to the <command> with any [command arguments].`

	envxURLPrefix = "envx+"
)

var (
	errUsage      = errors.New("usage error")
	errHTTPStatus = errors.New("http status error")
)

func usage(w io.Writer) {
	fmt.Fprintln(w, usageString)
}

func Run(argv []string, env []string) error {
	if len(argv) <= 1 {
		usage(os.Stderr)
		return errUsage
	}

	if argv[1] == "-h" || argv[1] == "--help" {
		usage(os.Stderr)
		return nil
	}

	envMap := sliceToMap(os.Environ())

	timeout := 10 * time.Second
	if v, ok := envMap["ENVX_TIMEOUT"]; ok {
		dv, err := time.ParseDuration(v)

		if err != nil {
			return fmt.Errorf("failed to parse ENVX_TIMEOUT duration: %w", err)
		}

		timeout = dv
	}

	if v, ok := envMap["ENVX_HTTP_TIMEOUT"]; ok {
		dv, err := time.ParseDuration(v)

		if err != nil {
			return fmt.Errorf("failed to parse ENVX_HTTP_TIMEOUT duration: %w", err)
		}

		internal.HTTPClient.Timeout = dv
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	for _, f := range internal.Expansions {
		m, err := f(ctx, envMap)

		if err != nil {
			return fmt.Errorf("expansion failed: %w", err)
		}

		envMap = m
	}

	if err := syscall.Exec(argv[1], argv[1:], mapToSlice(envMap)); err != nil {
		return fmt.Errorf("exec failed: %w", err)
	}

	return nil
}

func sliceToMap(sl []string) map[string]string {
	m := map[string]string{}

	for _, entry := range sl {
		if key, value, ok := strings.Cut(entry, "="); ok {
			m[key] = value
		}
	}

	return m
}

func mapToSlice(m map[string]string) []string {
	keys := []string{}

	for key := range m {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	sl := []string{}

	for _, key := range keys {
		sl = append(sl, key+"="+m[key])
	}

	return sl
}
