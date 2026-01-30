package state

import "github.com/charmbracelet/bubbles/viewport"

// https://www.youtube.com/watch?v=Gl31diSVP8M

type StateInterface interface {
	KeyPressTab()

	KeyPressUp()
	KeyPressDown()
	KeyPressLeft()
	KeyPressRight()

	KeyPressCtrlD()
	KeyPressCtrlP()
}

type ResutlValue struct {
	ID       string
	Title    string
	Body     string
	Expanded string
}

type InputMethodInterface interface {
	GetMatching(string) []string
	ToggleInputs()
	Focus()
	Submit() *[]ResutlValue
}

type ResultInterface interface {
	FocusX(int)
	// Goes Up if nothing to go up then return false
	GoUp() bool
	GoDown()
}

type BaseResult struct {
	viewportArea     viewport.Model
	resultValues     []ResutlValue
	resutlValueIndex int
	ResultInterface
}

func NewBaseResult(result *[]ResutlValue, width int, hight int) *BaseResult {
	return &BaseResult{
		resultValues:     *result,
		resutlValueIndex: -1,
		viewportArea:     viewport.New(width, hight),
	}
}

func (self *BaseResult) FocusX(index int) {
	if -1 < index && index < len(self.resultValues) {
		self.resutlValueIndex = index

	}
}

type StateRegion int

const (
	InputRegion StateRegion = iota
	ResultRegion
)

type BaseState struct {
	focusIndex StateRegion
	results    BaseResult
	input      InputMethodInterface
}

func (self *BaseState) KeyPressCtrlP() {
	self.input.Submit()
}

func (self *BaseState) KeyPressTab() {
	self.input.ToggleInputs()
}

// type

type FeatureState struct {
	BaseState
	StateInterface
}

func NewFeatureState()
