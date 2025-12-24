package openinghours

import (
	"testing"
)

func TestIsEqualTo_IdenticalStrings(t *testing.T) {
	oh1, err := New("Mo-Fr 09:00-17:00")
	if err != nil {
		t.Fatalf("failed to parse oh1: %v", err)
	}

	oh2, err := New("Mo-Fr 09:00-17:00")
	if err != nil {
		t.Fatalf("failed to parse oh2: %v", err)
	}

	if !oh1.IsEqualTo(oh2) {
		t.Error("identical strings should be equal")
	}
}

func TestIsEqualTo_SemanticallyEqual(t *testing.T) {
	// Same meaning, different representation
	oh1, err := New("Mo-Fr 09:00-17:00")
	if err != nil {
		t.Fatalf("failed to parse oh1: %v", err)
	}

	oh2, err := New("Mo,Tu,We,Th,Fr 09:00-17:00")
	if err != nil {
		t.Fatalf("failed to parse oh2: %v", err)
	}

	if !oh1.IsEqualTo(oh2) {
		t.Error("semantically equal values should be equal")
	}
}

func TestIsEqualTo_DifferentTimeRanges(t *testing.T) {
	oh1, err := New("Mo-Fr 09:00-17:00")
	if err != nil {
		t.Fatalf("failed to parse oh1: %v", err)
	}

	oh2, err := New("Mo-Fr 09:00-18:00")
	if err != nil {
		t.Fatalf("failed to parse oh2: %v", err)
	}

	if oh1.IsEqualTo(oh2) {
		t.Error("different time ranges should not be equal")
	}
}

func TestIsEqualTo_DifferentDays(t *testing.T) {
	oh1, err := New("Mo-Fr 09:00-17:00")
	if err != nil {
		t.Fatalf("failed to parse oh1: %v", err)
	}

	oh2, err := New("Mo-Sa 09:00-17:00")
	if err != nil {
		t.Fatalf("failed to parse oh2: %v", err)
	}

	if oh1.IsEqualTo(oh2) {
		t.Error("different day ranges should not be equal")
	}
}

func TestIsEqualTo_24_7Variants(t *testing.T) {
	oh1, err := New("24/7")
	if err != nil {
		t.Fatalf("failed to parse oh1: %v", err)
	}

	oh2, err := New("00:00-24:00")
	if err != nil {
		t.Fatalf("failed to parse oh2: %v", err)
	}

	if !oh1.IsEqualTo(oh2) {
		t.Error("24/7 and 00:00-24:00 should be equal")
	}
}

func TestIsEqualTo_OffVariants(t *testing.T) {
	oh1, err := New("off")
	if err != nil {
		t.Fatalf("failed to parse oh1: %v", err)
	}

	oh2, err := New("closed")
	if err != nil {
		t.Fatalf("failed to parse oh2: %v", err)
	}

	if !oh1.IsEqualTo(oh2) {
		t.Error("off and closed should be equal")
	}
}

func TestIsEqualTo_MultipleTimeRanges(t *testing.T) {
	oh1, err := New("Mo-Fr 09:00-12:00,13:00-17:00")
	if err != nil {
		t.Fatalf("failed to parse oh1: %v", err)
	}

	oh2, err := New("Mo-Fr 09:00-12:00,13:00-17:00")
	if err != nil {
		t.Fatalf("failed to parse oh2: %v", err)
	}

	if !oh1.IsEqualTo(oh2) {
		t.Error("identical multiple time ranges should be equal")
	}
}

func TestIsEqualTo_ReorderedTimeRanges(t *testing.T) {
	// Order shouldn't matter for equality
	oh1, err := New("Mo-Fr 09:00-12:00,14:00-17:00")
	if err != nil {
		t.Fatalf("failed to parse oh1: %v", err)
	}

	oh2, err := New("Mo-Fr 14:00-17:00,09:00-12:00")
	if err != nil {
		t.Fatalf("failed to parse oh2: %v", err)
	}

	if !oh1.IsEqualTo(oh2) {
		t.Error("reordered time ranges should be equal")
	}
}

func TestIsEqualTo_WithComments(t *testing.T) {
	// Comments are part of the value, so different comments = not equal
	oh1, err := New("Mo-Fr 09:00-17:00 \"Main hours\"")
	if err != nil {
		t.Fatalf("failed to parse oh1: %v", err)
	}

	oh2, err := New("Mo-Fr 09:00-17:00 \"Different comment\"")
	if err != nil {
		t.Fatalf("failed to parse oh2: %v", err)
	}

	if oh1.IsEqualTo(oh2) {
		t.Error("different comments should not be equal")
	}
}

func TestIsEqualTo_SameComment(t *testing.T) {
	oh1, err := New("Mo-Fr 09:00-17:00 \"Open\"")
	if err != nil {
		t.Fatalf("failed to parse oh1: %v", err)
	}

	oh2, err := New("Mo-Fr 09:00-17:00 \"Open\"")
	if err != nil {
		t.Fatalf("failed to parse oh2: %v", err)
	}

	if !oh1.IsEqualTo(oh2) {
		t.Error("same comments should be equal")
	}
}

func TestIsEqualTo_SplitRules(t *testing.T) {
	// Rules split across semicolons vs combined
	oh1, err := New("Mo-Fr 09:00-17:00; Sa 10:00-14:00")
	if err != nil {
		t.Fatalf("failed to parse oh1: %v", err)
	}

	oh2, err := New("Mo-Fr 09:00-17:00; Sa 10:00-14:00")
	if err != nil {
		t.Fatalf("failed to parse oh2: %v", err)
	}

	if !oh1.IsEqualTo(oh2) {
		t.Error("identical multi-rule values should be equal")
	}
}

func TestIsEqualTo_EquivalentSplitRules(t *testing.T) {
	// Same result from different rule structures
	oh1, err := New("Mo 09:00-17:00; Tu 09:00-17:00")
	if err != nil {
		t.Fatalf("failed to parse oh1: %v", err)
	}

	oh2, err := New("Mo-Tu 09:00-17:00")
	if err != nil {
		t.Fatalf("failed to parse oh2: %v", err)
	}

	if !oh1.IsEqualTo(oh2) {
		t.Error("equivalent split rules should be equal")
	}
}

func TestIsEqualTo_Nil(t *testing.T) {
	oh1, err := New("Mo-Fr 09:00-17:00")
	if err != nil {
		t.Fatalf("failed to parse oh1: %v", err)
	}

	if oh1.IsEqualTo(nil) {
		t.Error("non-nil should not equal nil")
	}
}
