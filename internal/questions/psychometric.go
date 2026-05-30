package questions

import (
	"fmt"
	"math"
	"math/rand"
)

// PsychometricItem describes one item in a psychometric plan.
type PsychometricItem struct {
	Kind          string // "single", "matrix"
	QuestionIndex int
	RowIndex      *int
	OptionCount   int
	Bias          string // "left", "center", "right"
	IsReversed    bool
	ScoreByChoice []float64 // score mapping for each choice
	TargetProb    []float64
}

// PsychometricPlan holds pre-generated answers for a single dimension.
type PsychometricPlan struct {
	Items   []PsychometricItem
	Theta   float64
	SigmaE  float64
	Choices map[string]int // key -> choice index
}

// GetChoice returns the pre-generated choice for a question.
func (p *PsychometricPlan) GetChoice(questionIndex int, rowIndex *int) *int {
	key := choiceKey(questionIndex, rowIndex)
	if choice, ok := p.Choices[key]; ok {
		return &choice
	}
	return nil
}

// DimensionPsychometricPlan holds plans per dimension.
type DimensionPsychometricPlan struct {
	Plans map[string]*PsychometricPlan
}

// GetChoice delegates to the correct dimension's plan.
func (d *DimensionPsychometricPlan) GetChoice(questionIndex int, rowIndex *int) *int {
	for _, plan := range d.Plans {
		if choice := plan.GetChoice(questionIndex, rowIndex); choice != nil {
			return choice
		}
	}
	return nil
}

// BuildPsychometricPlan creates a psychometric plan for a set of items.
func BuildPsychometricPlan(items []PsychometricItem, targetAlpha float64) *PsychometricPlan {
	if len(items) < 2 {
		return nil
	}
	if targetAlpha <= 0 {
		targetAlpha = 0.85
	}

	k := len(items)
	rho := computeRhoFromAlpha(targetAlpha, k)
	sigmaE := computeSigmaEFromRho(rho)
	theta := rand.NormFloat64()

	choices := make(map[string]int)
	orientation := inferDimensionOrientation(items)
	for _, item := range items {
		key := choiceKey(item.QuestionIndex, item.RowIndex)
		itemDirection := orientation.ItemDirections[key]
		if itemDirection == "" {
			itemDirection = item.Bias
		}
		score := generatePsychoAnswer(theta, item.OptionCount, itemDirection, sigmaE, orientation.ReversedKeys[key] || item.IsReversed)
		// Apply score_by_choice mapping if available
		choice := mapScoreToChoice(score, item)
		choices[key] = choice
	}

	return &PsychometricPlan{
		Items:   items,
		Theta:   theta,
		SigmaE:  sigmaE,
		Choices: choices,
	}
}

// BuildDimensionPsychometricPlan creates per-dimension plans.
func BuildDimensionPsychometricPlan(groupedItems map[string][]PsychometricItem, targetAlpha float64) *DimensionPsychometricPlan {
	plans := make(map[string]*PsychometricPlan)
	for dimension, items := range groupedItems {
		if len(items) >= 2 {
			plan := BuildPsychometricPlan(items, targetAlpha)
			if plan != nil {
				plans[dimension] = plan
			}
		}
	}
	if len(plans) == 0 {
		return nil
	}
	return &DimensionPsychometricPlan{Plans: plans}
}

type dimensionOrientation struct {
	ItemDirections map[string]string
	ReversedKeys   map[string]bool
}

func inferDimensionOrientation(items []PsychometricItem) dimensionOrientation {
	result := dimensionOrientation{
		ItemDirections: make(map[string]string),
		ReversedKeys:   make(map[string]bool),
	}
	leftStrength := 0.0
	rightStrength := 0.0
	strengths := make(map[string]float64)
	for _, item := range items {
		key := choiceKey(item.QuestionIndex, item.RowIndex)
		direction, strength := itemDirection(item)
		result.ItemDirections[key] = direction
		strengths[key] = strength
		switch direction {
		case "left":
			leftStrength += strength
		case "right":
			rightStrength += strength
		}
	}

	anchor := "center"
	anchorStrength := leftStrength
	weakerStrength := rightStrength
	if rightStrength > leftStrength {
		anchor = "right"
		anchorStrength = rightStrength
		weakerStrength = leftStrength
	} else if leftStrength > rightStrength {
		anchor = "left"
	}
	ambiguous := anchor == "center" || anchorStrength < 0.2 || anchorStrength <= weakerStrength*1.15
	if ambiguous {
		return result
	}
	for key, direction := range result.ItemDirections {
		if strengths[key] <= 0 {
			continue
		}
		if (direction == "left" || direction == "right") && direction != anchor {
			result.ReversedKeys[key] = true
		}
	}
	return result
}

func itemDirection(item PsychometricItem) (string, float64) {
	probs := normalizePsychometricProbabilities(item.TargetProb, item.OptionCount)
	if len(probs) == 0 {
		probs = buildBiasTargetProbabilities(item.OptionCount, item.Bias)
	}
	denom := float64(maxInt(item.OptionCount-1, 1))
	mean := 0.5
	if denom > 0 {
		weighted := 0.0
		for idx, value := range probs {
			weighted += float64(idx) * value
		}
		mean = math.Max(0, math.Min(1, weighted/denom))
	}
	if mean <= 0.4 {
		return "left", math.Abs(mean - 0.5)
	}
	if mean >= 0.6 {
		return "right", math.Abs(mean - 0.5)
	}
	return "center", math.Abs(mean - 0.5)
}

func normalizePsychometricProbabilities(values []float64, optionCount int) []float64 {
	if optionCount <= 0 {
		return nil
	}
	result := make([]float64, optionCount)
	total := 0.0
	for i := 0; i < optionCount && i < len(values); i++ {
		value := math.Max(0, values[i])
		if math.IsNaN(value) || math.IsInf(value, 0) {
			value = 0
		}
		result[i] = value
		total += value
	}
	if total <= 0 {
		return nil
	}
	for i := range result {
		result[i] /= total
	}
	return result
}

func buildBiasTargetProbabilities(optionCount int, bias string) []float64 {
	if optionCount <= 1 {
		optionCount = 2
	}
	if optionCount == 2 {
		switch bias {
		case "left":
			return []float64{0.75, 0.25}
		case "right":
			return []float64{0.25, 0.75}
		default:
			return []float64{0.5, 0.5}
		}
	}
	raw := make([]float64, optionCount)
	center := float64(optionCount-1) / 2
	power := 8.0
	if bias == "center" {
		power = 3
	}
	for i := range raw {
		var linear float64
		switch bias {
		case "left":
			linear = 1.0 - float64(i)/float64(optionCount-1)
		case "right":
			linear = float64(i) / float64(optionCount-1)
		default:
			linear = 1.0 - math.Abs(float64(i)-center)/math.Max(center, 1)
		}
		raw[i] = math.Pow(math.Max(linear, 0), power)
	}
	return normalizePsychometricProbabilities(raw, optionCount)
}

func computeRhoFromAlpha(alpha float64, k int) float64 {
	// rho = alpha / (k - alpha*(k-1))
	denom := float64(k) - alpha*(float64(k)-1)
	if denom <= 0 {
		return 0.5
	}
	return alpha / denom
}

func computeSigmaEFromRho(rho float64) float64 {
	if rho <= 0 {
		return 1.0
	}
	// sigma_e = sqrt(1/rho - 1)
	return math.Sqrt(1.0/rho - 1.0)
}

func generatePsychoAnswer(theta float64, optionCount int, bias string, sigmaE float64, isReversed bool) int {
	biasShift := 0.0
	switch bias {
	case "left":
		biasShift = -0.5
	case "right":
		biasShift = 0.5
	}

	effectiveTheta := theta
	if isReversed {
		effectiveTheta = -theta
	}

	z := effectiveTheta + biasShift + sigmaE*rand.NormFloat64()
	return zToCategory(z, optionCount)
}

// mapScoreToChoice maps a raw score to a choice index using ScoreByChoice if available.
func mapScoreToChoice(score int, item PsychometricItem) int {
	if len(item.ScoreByChoice) == 0 {
		return score
	}
	// ScoreByChoice maps choice index -> score value
	// Find the choice index whose score matches
	for choiceIdx, targetScore := range item.ScoreByChoice {
		if int(targetScore) == score {
			return choiceIdx
		}
	}
	// Fallback: find closest score
	bestIdx := 0
	bestDiff := 999
	for choiceIdx, targetScore := range item.ScoreByChoice {
		diff := abs(int(targetScore) - score)
		if diff < bestDiff {
			bestDiff = diff
			bestIdx = choiceIdx
		}
	}
	return bestIdx
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func zToCategory(z float64, optionCount int) int {
	if optionCount <= 1 {
		return 0
	}
	// Use inverse normal quantile thresholds (matching Python algorithm)
	// Divide [0,1] into optionCount equal bins, use inverse-normal thresholds
	m := float64(optionCount)
	phi := normalCDF(z)
	// Find which bin z falls into
	idx := int(math.Floor(phi * m))
	if idx >= optionCount {
		idx = optionCount - 1
	}
	if idx < 0 {
		idx = 0
	}
	return idx
}

// normalCDF computes the cumulative distribution function of the standard normal.
func normalCDF(x float64) float64 {
	return 0.5 * (1 + math.Erf(x/math.Sqrt2))
}

func choiceKey(questionIndex int, rowIndex *int) string {
	if rowIndex != nil {
		return fmt.Sprintf("q:%d:%d", questionIndex, *rowIndex)
	}
	return fmt.Sprintf("q:%d", questionIndex)
}
