package dna

import (
	"testing"
)

func TestCompress_BenchmarkReductionRatio(t *testing.T) {
	verboseDoc := `
	Please kindly note that I would like to say that maybe and perhaps the architecture of the system might be typically designed with several layers.
	However, furthermore, and moreover, additionally, nevertheless, nonetheless, consequently, thereby, henceforth the controller communicates with the service layer.
	I think that a repository is responsible for the data persistence and the database queries.
	Thank you very much for your time and please kindly let me know if you have any questions.
	`

	res := Compress(verboseDoc, Aggressive)

	if res.Ratio < 0.40 {
		t.Errorf("expected aggressive compression ratio >= 0.40 (40%%), got %.2f (%.1f%%)", res.Ratio, res.Ratio*100)
	}

	t.Logf("Original size: %d, Compressed size: %d, Reduction: %.1f%%",
		res.OriginalSize, res.CompressedSize, res.Ratio*100)
}

func BenchmarkCavemanCompress(b *testing.B) {
	sample := `
	Please kindly note that I would like to mention that maybe the service layer shouldn't call the controller layer directly.
	However, furthermore, it is typically recommended to adhere to SOLID principles.
	`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Compress(sample, Normal)
	}
}
