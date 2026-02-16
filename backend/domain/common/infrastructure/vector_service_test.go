package infrastructure

import "testing"

func TestVectorService_Search_Synonyms(t *testing.T) {
	svc := NewVectorService("")
	_ = svc.Upsert("ac1", "This is an Air Conditioner in my room", nil)
	_ = svc.Upsert("tv1", "Smart Television in the living room", nil)
	_ = svc.Upsert("lamp1", "Desk Lamp", nil)

	t.Run("Match AC", func(t *testing.T) {
		res, _ := svc.Search("Turn on the AC")
		if len(res) == 0 || res[0] != "ac1" {
			t.Errorf("expected ac1, got %v", res)
		}
	})

	t.Run("Match TV", func(t *testing.T) {
		res, _ := svc.Search("Matikan TV")
		if len(res) == 0 || res[0] != "tv1" {
			t.Errorf("expected tv1, got %v", res)
		}
	})

	t.Run("Match Lamp from Light", func(t *testing.T) {
		res, _ := svc.Search("Nyalakan Light")
		if len(res) == 0 || res[0] != "lamp1" {
			t.Errorf("expected lamp1, got %v", res)
		}
	})
}
