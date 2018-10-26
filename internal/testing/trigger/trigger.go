package trigger

import (
	"context"
	"fmt"

	"github.com/project-flogo/core/data/metadata"
	"github.com/project-flogo/core/trigger"
)

var triggerMetadata = trigger.NewMetadata(&Settings{}, &HandlerSettings{}, &Output{}, &Reply{})

func init() {
	trigger.Register(&Trigger{}, &TriggerFactory{})
}

// Handler is a trigger handler
type Handler struct {
	handler  trigger.Handler
	settings string
}

// Trigger is a dummy trigger for testing
type Trigger struct {
	settings *Settings
	id       string
}

var handlers []Handler

// TriggerFactory creates dummy triggers for testing
type TriggerFactory struct {
}

// New creates a new dummy trigger
func (*TriggerFactory) New(config *trigger.Config) (trigger.Trigger, error) {
	s := &Settings{}
	err := metadata.MapToStruct(config.Settings, s, true)
	if err != nil {
		return nil, err
	}

	return &Trigger{id: config.Id, settings: s}, nil
}

func (f *TriggerFactory) Metadata() *trigger.Metadata {
	return triggerMetadata
}

// Metadata implements trigger.Trigger.Metadata
func (t *Trigger) Metadata() *trigger.Metadata {
	return triggerMetadata
}

func (t *Trigger) Initialize(ctx trigger.InitContext) error {
	for _, handler := range ctx.GetHandlers() {
		s := &HandlerSettings{}
		err := metadata.MapToStruct(handler.Settings(), s, true)
		if err != nil {
			return err
		}

		handlers = append(handlers, Handler{
			settings: s.ASetting,
			handler:  handler,
		})
	}

	return nil
}

// Start starts the dummy trigger
func (t *Trigger) Start() error {
	return nil
}

// Stop stops the dummy trigger
func (t *Trigger) Stop() error {
	return nil
}

// Fire is a test function for firing one of the trigger handlers with given content
func Fire(h int, content interface{}) (map[string]interface{}, error) {
	if h >= len(handlers) {
		return nil, fmt.Errorf("invalid handler %v", h)
	}
	output := &Output{Content: content}
	return handlers[h].handler.Handle(context.Background(), output.ToMap())
}

// Reset resets the trigger for another test
func Reset() {
	handlers = nil
}
