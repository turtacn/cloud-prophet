package parallelize

import (
	"context"
	"k8s.io/client-go/util/workqueue"
	"math"
)

const parallelism = 16

func chunkSizeFor(n int) workqueue.Options {
	s := int(math.Sqrt(float64(n)))
	if r := n/parallelism + 1; s > r {
		s = r
	} else if s < 1 {
		s = 1
	}
	return workqueue.WithChunkSize(s)
}

func Until(ctx context.Context, pieces int, doWorkPiece workqueue.DoWorkPieceFunc) {
	workqueue.ParallelizeUntil(ctx, parallelism, pieces, doWorkPiece, chunkSizeFor(pieces))
}
