package infrastructure

import "testing"

func TestVectorService_Search(t *testing.T) {
	svc := NewVectorService()
	res, err := svc.Search("hello world")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(res) == 0 {
		t.Fatalf("expected at least one result, got 0")
	}
}
