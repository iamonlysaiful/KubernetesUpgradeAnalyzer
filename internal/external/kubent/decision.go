package kubent

type DecisionStatus string

const (
	DecisionGo   DecisionStatus = "GO"
	DecisionNoGo DecisionStatus = "NO_GO"
)

type Decision struct {
	Status      DecisionStatus
	Summary     string
	Limitations []Limitation
}

func DecideKubentMVP(results []CoverageResult) Decision {
	var limitations []Limitation
	for _, result := range results {
		if result.Status != CoverageVerified {
			limitations = append(limitations, Limitation{
				Code:    "TARGET_COVERAGE_NOT_VERIFIED",
				Summary: "kubent coverage is not verified for " + result.TargetVersion,
			})
		}
	}
	if len(limitations) > 0 {
		return Decision{
			Status:      DecisionNoGo,
			Summary:     "kubent coverage is not sufficient for MVP recommendation claims",
			Limitations: limitations,
		}
	}
	return Decision{
		Status:  DecisionGo,
		Summary: "kubent coverage is verified for MVP target minors",
	}
}
