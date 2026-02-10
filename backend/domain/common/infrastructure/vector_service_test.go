package infrastructure

import "testing"

func TestVectorService_Search(t *testing.T) {
	svc := NewVectorService("")
	_ = svc.Upsert("doc1", "hello world", nil)
	res, err := svc.Search("hello world")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(res) == 0 {
		t.Fatalf("expected at least one result, got 0")
	}
}
