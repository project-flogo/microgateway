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

type Handler struct {
	handler  trigger.Handler
	settings string
}

type Trigger struct {
	settings *Settings
	id       string
}

var handlers []Handler

type TriggerFactory struct {
}

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

func (t *Trigger) Start() error {
	return nil
}

func (t *Trigger) Stop() error {
	return nil
}

func Fire(h int, content interface{}) (map[string]interface{}, error) {
	if h >= len(handlers) {
		return nil, fmt.Errorf("invalid handler %v", h)
	}
	output := &Output{Content: content}
	return handlers[h].handler.Handle(context.Background(), output.ToMap())
}
