package models

import "github.com/SurveyController/SurveyCore/internal/domain"

const (
	ReverseFillFormatAuto        = domain.ReverseFillFormatAuto
	ReverseFillFormatWJXSequence = domain.ReverseFillFormatWJXSequence
	ReverseFillFormatWJXScore    = domain.ReverseFillFormatWJXScore
	ReverseFillFormatWJXText     = domain.ReverseFillFormatWJXText

	ReverseFillStatusReverse  = domain.ReverseFillStatusReverse
	ReverseFillStatusFallback = domain.ReverseFillStatusFallback
	ReverseFillStatusBlocked  = domain.ReverseFillStatusBlocked

	ReverseFillKindChoice    = domain.ReverseFillKindChoice
	ReverseFillKindText      = domain.ReverseFillKindText
	ReverseFillKindMultiText = domain.ReverseFillKindMultiText
	ReverseFillKindMatrix    = domain.ReverseFillKindMatrix
)

type ReverseFillAnswer = domain.ReverseFillAnswer
type ReverseFillSampleRow = domain.ReverseFillSampleRow
type ReverseFillIssue = domain.ReverseFillIssue
type ReverseFillQuestionPlan = domain.ReverseFillQuestionPlan
type ReverseFillSpec = domain.ReverseFillSpec
type ReverseFillRuntimeState = domain.ReverseFillRuntimeState
type ReverseFillAcquireResult = domain.ReverseFillAcquireResult
