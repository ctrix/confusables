package confusables

import (
	"testing"
)

func TestSkeleton(t *testing.T) {
	s := "ρ⍺у𝓅𝒂ן"
	expected := "paypal"
	skeleton := Skeleton(s)

	if skeleton != expected {
		t.Errorf("skeleton(%s) should result in %s", s, expected)
	}
}

func TestCompareEqual(t *testing.T) {
	vectors := [][]string{
		[]string{"ρ⍺у𝓅𝒂ן", "𝔭𝒶ỿ𝕡𝕒ℓ"},
		[]string{"𝖶", "W"},
		[]string{"so̷s", "søs"},
		[]string{"paypal", "paypal"},
		[]string{"scope", "scope"},
		[]string{"ø", "o̷"},
		[]string{"O", "0"},
		[]string{"ν", "v"},
		[]string{"Ι", "l"},
	}

	for _, v := range vectors {
		s1, s2 := v[0], v[1]
		if Skeleton(s1) != Skeleton(s2) {
			t.Errorf("Skeleton strings %+q and %+q were expected to be equal", s1, s2)
		}
	}
}

func TestCompareDifferent(t *testing.T) {
	s1 := "Paypal"
	s2 := "paypal"

	if Skeleton(s1) == Skeleton(s2) {
		t.Errorf("Skeleton strings %+q and %+q were expected to be different", s1, s2)
	}
}

func BenchmarkSkeletonNoop(b *testing.B) {
	s := "skeleton"

	for i := 0; i < b.N; i++ {
		Skeleton(s)
	}
}

func BenchmarkSkeleton(b *testing.B) {
	s := "ѕ𝗄℮|е𝗍ο𝔫"

	for i := 0; i < b.N; i++ {
		Skeleton(s)
	}
}
