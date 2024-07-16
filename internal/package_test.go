package internal_test

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/posit-dev/envx/internal"
)

func TestExpandVars(t *testing.T) {
	m := map[string]string{
		"GRILLED_CHEESE": "sandwich",
		"TACO":           "pocket",
		"HOT_DOG":        "sandwich",
		"BURRITO":        "tube",
		"SANDWICHES":     "$GRILLED_CHEESE ${HOT_DOG}",
	}

	em, err := internal.ExpandVars(context.Background(), m)
	if err != nil {
		t.Errorf("failed to expand: %v", err)
		return
	}

	if em["TACO"] != "pocket" {
		t.Errorf("TACO != %q: %q", "pocket", em["TACO"])
	}

	if em["SANDWICHES"] != "sandwich sandwich" {
		t.Errorf("SANDWICHES != %q: %q", "sandwich sandwich", em["SANDWICHES"])
	}
}

func TestExpandURLs(t *testing.T) {
	secrets := map[string]string{
		"combo": fmt.Sprintf("dog-%v-elf-%v-bop", rand.Int63(), rand.Int63()),
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if v, ok := secrets[strings.TrimPrefix(req.URL.Path, "/")]; ok {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "%s", v)
			return
		}

		w.WriteHeader(http.StatusNotFound)
		return
	}))

	t.Cleanup(ts.Close)

	t.Run("typical http", func(t *testing.T) {
		m := map[string]string{
			"COMBINATION":          fmt.Sprintf("envx+%s/combo", ts.URL),
			"DOORKEEPER_SENTIMENT": fmt.Sprintf("all are welcome (%v)", rand.Int63()),
		}

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		t.Cleanup(cancel)

		em, err := internal.ExpandURLs(ctx, m)
		if err != nil {
			t.Errorf("failed to expand: %v", err)
			return
		}

		if em["COMBINATION"] != secrets["combo"] {
			t.Errorf("COMBINATION != %q: %q", secrets["combo"], em["COMBINATION"])
		}

		if em["DOORKEEPER_SENTIMENT"] != m["DOORKEEPER_SENTIMENT"] {
			t.Error("non-secret DOORKEEPER_SENTIMENT was not retained")
		}
	})

	t.Run("nonexistent http", func(t *testing.T) {
		m := map[string]string{
			"RABBITS": fmt.Sprintf("envx+%s/rabbits", ts.URL),
		}

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		t.Cleanup(cancel)

		if _, err := internal.ExpandURLs(ctx, m); err == nil {
			t.Error("expansion did not fail")
		}
	})

	plampts := fmt.Sprintf("ivy %v\ntree %v\nbushies %v\n", rand.Int63(), rand.Int63(), rand.Int63())

	tmpDir := t.TempDir()
	if err := os.WriteFile(
		filepath.Join(tmpDir, "plampts.txt"),
		[]byte(plampts),
		0644,
	); err != nil {
		t.Errorf("failed to write file: %v", err)
		return
	}

	t.Run("typical file", func(t *testing.T) {
		m := map[string]string{
			"PLAMPTS": fmt.Sprintf("envx+file://%s/plampts.txt", tmpDir),
		}

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		t.Cleanup(cancel)

		em, err := internal.ExpandURLs(ctx, m)
		if err != nil {
			t.Errorf("failed to expand: %v", err)
			return
		}

		if em["PLAMPTS"] != plampts {
			t.Errorf("PLAMPTS != %q: %q", plampts, em["PLAMPTS"])
		}
	})

	t.Run("nonexistent file", func(t *testing.T) {
		m := map[string]string{
			"ROMCKS": fmt.Sprintf("envx+file://%s/romcks.txt", tmpDir),
		}

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		t.Cleanup(cancel)

		if _, err := internal.ExpandURLs(ctx, m); err == nil {
			t.Error("expansion did not fail")
		}
	})
}
