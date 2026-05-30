package models

import (
	"github.com/SurveyController/SurveyConsole/internal/execution"
	runstate "github.com/SurveyController/SurveyConsole/internal/runtime"
)

type ExecutionConfig = execution.ExecutionConfig
type ThreadProgressState = runstate.ThreadProgressState
type ExecutionState = runstate.ExecutionState

func NewExecutionState() *ExecutionState {
	return runstate.NewExecutionState()
}
