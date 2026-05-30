package models

import (
	"github.com/SurveyController/SurveyCore/internal/execution"
	runstate "github.com/SurveyController/SurveyCore/internal/runtime"
)

type ExecutionConfig = execution.ExecutionConfig
type ThreadProgressState = runstate.ThreadProgressState
type ExecutionState = runstate.ExecutionState

func NewExecutionState() *ExecutionState {
	return runstate.NewExecutionState()
}
