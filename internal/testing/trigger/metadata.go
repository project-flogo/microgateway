package trigger

import "github.com/project-flogo/core/data/coerce"

// Settings are the settings for the dummy trigger
type Settings struct {
	ASetting int `md:"aSetting,required"`
}

// HandlerSettings are the settings for the dummy trigger handlers
type HandlerSettings struct {
	ASetting string `md:aSetting,required"`
}

// Output is the output of the dummy trigger
type Output struct {
	Content interface{} `md:"content"`
}

// FromMap sets Output from a map
func (o *Output) FromMap(values map[string]interface{}) error {
	var err error
	o.Content, err = coerce.ToAny(values["content"])
	if err != nil {
		return err
	}

	return nil
}

// ToMap converts Output to a map
func (o *Output) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"content": o.Content,
	}
}

// Reply is the reply the trigger gets
type Reply struct {
	AReply interface{} `md:"aReply"`
}

// FromMap sets Reply from a map
func (r *Reply) FromMap(values map[string]interface{}) error {
	r.AReply = values["aReply"]
	return nil
}

// ToMap converts Reply to a map
func (r *Reply) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"aReply": r.AReply,
	}
}
