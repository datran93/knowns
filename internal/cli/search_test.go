package cli

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/howznguyen/knowns/internal/models"
)

func TestSprintPlainRetrieval(t *testing.T) {
	resp := &models.RetrievalResponse{
		Query: "rag retrieval foundation",
		Mode:  "keyword",
		Candidates: []models.RetrievalCandidate{
			{
				Type:        "doc",
				ID:          "specs/rag-retrieval-foundation",
				Title:       "RAG Retrieval Foundation",
				Score:       1,
				DirectMatch: true,
				Citation: models.Citation{
					Type: "doc",
					Path: "specs/rag-retrieval-foundation",
				},
				Snippet: "Specification for retrieval foundation across docs, tasks, and memories",
			},
		},
		ContextPack: models.ContextPack{
			Items: []models.ContextItem{
				{
					Type:        "doc",
					ID:          "specs/rag-retrieval-foundation",
					Title:       "RAG Retrieval Foundation",
					DirectMatch: true,
					Citation: models.Citation{
						Type: "doc",
						Path: "specs/rag-retrieval-foundation",
					},
					Content: "Build a shared retrieval foundation for Knowns.",
				},
			},
		},
	}

	got := sprintPlainRetrieval(resp)
	for _, want := range []string{
		"Query: rag retrieval foundation",
		"Candidates: 1",
		"[DOC] RAG Retrieval Foundation (specs/rag-retrieval-foundation)",
		"Citation: doc:specs/rag-retrieval-foundation",
		"Context Pack:",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, got)
		}
	}
}

func TestExtractONNXLibTarGzRegularFile(t *testing.T) {
	archivePath := filepath.Join(t.TempDir(), "onnxruntime.tgz")
	writeTarGzFixture(t, archivePath, []tarFixtureEntry{
		{
			name:     "onnxruntime-linux-x64-1.24.3/lib/libonnxruntime.so",
			body:     []byte("regular-so"),
			typeflag: tar.TypeReg,
		},
	})

	destPath := filepath.Join(t.TempDir(), "libonnxruntime.so")
	if err := extractONNXLib(archivePath, "libonnxruntime.so", destPath); err != nil {
		t.Fatalf("extractONNXLib returned error: %v", err)
	}

	got, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("read dest file: %v", err)
	}
	if string(got) != "regular-so" {
		t.Fatalf("expected regular file bytes, got %q", string(got))
	}
}

func TestExtractONNXLibTarGzSymlinkChain(t *testing.T) {
	archivePath := filepath.Join(t.TempDir(), "onnxruntime.tgz")
	writeTarGzFixture(t, archivePath, []tarFixtureEntry{
		{
			name:     "onnxruntime-linux-x64-1.24.3/lib/libonnxruntime.so",
			linkname: "libonnxruntime.so.1",
			typeflag: tar.TypeSymlink,
		},
		{
			name:     "onnxruntime-linux-x64-1.24.3/lib/libonnxruntime.so.1",
			linkname: "libonnxruntime.so.1.24.3",
			typeflag: tar.TypeSymlink,
		},
		{
			name:     "onnxruntime-linux-x64-1.24.3/lib/libonnxruntime.so.1.24.3",
			body:     []byte("versioned-so"),
			typeflag: tar.TypeReg,
		},
	})

	destPath := filepath.Join(t.TempDir(), "libonnxruntime.so")
	if err := extractONNXLib(archivePath, "libonnxruntime.so", destPath); err != nil {
		t.Fatalf("extractONNXLib returned error: %v", err)
	}

	got, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("read dest file: %v", err)
	}
	if string(got) != "versioned-so" {
		t.Fatalf("expected symlink target bytes, got %q", string(got))
	}
}

type tarFixtureEntry struct {
	name     string
	body     []byte
	linkname string
	typeflag byte
}

func writeTarGzFixture(t *testing.T, archivePath string, entries []tarFixtureEntry) {
	t.Helper()

	file, err := os.Create(archivePath)
	if err != nil {
		t.Fatalf("create archive: %v", err)
	}
	defer file.Close()

	gz := gzip.NewWriter(file)
	defer gz.Close()

	tw := tar.NewWriter(gz)
	defer tw.Close()

	for _, entry := range entries {
		hdr := &tar.Header{
			Name:     entry.name,
			Typeflag: entry.typeflag,
			Mode:     0755,
			Linkname: entry.linkname,
			Size:     int64(len(entry.body)),
		}
		if entry.typeflag != tar.TypeReg && entry.typeflag != tar.TypeRegA {
			hdr.Size = 0
		}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatalf("write header %s: %v", entry.name, err)
		}
		if len(entry.body) > 0 {
			if _, err := io.Copy(tw, bytes.NewReader(entry.body)); err != nil {
				t.Fatalf("write body %s: %v", entry.name, err)
			}
		}
	}
}
