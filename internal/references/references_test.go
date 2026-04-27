package references

import (
	"testing"

	"github.com/datran93/knowns/internal/models"
)

func TestParse_TaskSemanticRef(t *testing.T) {
	ref, ok := Parse("@task-rag001{blocked-by}")
	if !ok {
		t.Fatal("expected parse success")
	}
	if ref.Type != "task" || ref.Target != "rag001" {
		t.Fatalf("unexpected task ref: %+v", ref)
	}
	if ref.Relation != "blocked-by" || !ref.ExplicitRelation || !ref.ValidRelation {
		t.Fatalf("unexpected task relation parsing: %+v", ref)
	}
}

func TestParse_DocRefDefaultsToReferences(t *testing.T) {
	ref, ok := Parse("@doc/guides/setup")
	if !ok {
		t.Fatal("expected parse success")
	}
	if ref.Type != "doc" || ref.Target != "guides/setup" {
		t.Fatalf("unexpected doc ref: %+v", ref)
	}
	if ref.Relation != models.SemanticReferenceRelationReferences || ref.ExplicitRelation {
		t.Fatalf("expected default references relation, got %+v", ref)
	}
}

func TestParse_DocRefWithHeadingAndRelation(t *testing.T) {
	ref, ok := Parse("@doc/guides/setup#overview{implements}")
	if !ok {
		t.Fatal("expected parse success")
	}
	if ref.Target != "guides/setup" {
		t.Fatalf("target = %q, want guides/setup", ref.Target)
	}
	if ref.Fragment == nil || ref.Fragment.Heading != "overview" {
		t.Fatalf("expected heading fragment, got %+v", ref.Fragment)
	}
	if ref.Relation != "implements" {
		t.Fatalf("relation = %q, want implements", ref.Relation)
	}
}

func TestParse_DocRefWithLineRange(t *testing.T) {
	ref, ok := Parse("@doc/guides/setup:10-25{related}")
	if !ok {
		t.Fatal("expected parse success")
	}
	if ref.Fragment == nil || ref.Fragment.RangeStart != 10 || ref.Fragment.RangeEnd != 25 {
		t.Fatalf("expected line range fragment, got %+v", ref.Fragment)
	}
}

func TestParse_InvalidRelation(t *testing.T) {
	ref, ok := Parse("@memory-mem001{owns}")
	if !ok {
		t.Fatal("expected parse success")
	}
	if ref.ValidRelation {
		t.Fatalf("expected invalid relation, got %+v", ref)
	}
}

func TestExtract_MixedSemanticRefs(t *testing.T) {
	refs := Extract("See @doc/guides/setup{implements}, @task-rag001, and @memory-mem001{follows}.")
	if len(refs) != 3 {
		t.Fatalf("ref count = %d, want 3", len(refs))
	}
	if refs[1].Relation != models.SemanticReferenceRelationReferences {
		t.Fatalf("plain task ref should default to references, got %+v", refs[1])
	}
}
