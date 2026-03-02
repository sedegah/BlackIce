package features

import "blackice/core/fastpath"

func ShouldEscalate(f fastpath.WindowFeatures) bool {
	return f.SuspiciousScore >= 0.45
}
