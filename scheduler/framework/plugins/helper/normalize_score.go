package helper

import (
	framework "github.com/turtacn/cloud-prophet/scheduler/framework/base"
)

func DefaultNormalizeScore(maxPriority int64, reverse bool, scores framework.NodeScoreList) *framework.Status {
	var maxCount int64
	for i := range scores {
		if scores[i].Score > maxCount {
			maxCount = scores[i].Score
		}
	}

	if maxCount == 0 {
		if reverse {
			for i := range scores {
				scores[i].Score = maxPriority
			}
		}
		return nil
	}

	for i := range scores {
		score := scores[i].Score

		score = maxPriority * score / maxCount
		if reverse {
			score = maxPriority - score
		}

		scores[i].Score = score
	}
	return nil
}
