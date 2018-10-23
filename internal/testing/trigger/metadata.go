package trigger

import "github.com/project-flogo/core/data/coerce"

type Settings struct {
	ASetting int `md:"aSetting,required"`
}

type HandlerSettings struct {
	ASetting string `md:aSetting,required"`
}

type Output struct {
	Content interface{} `md:"content"`
}

func (o *Output) FromMap(values map[string]interface{}) error {
	var err error
	o.Content, err = coerce.ToAny(values["content"])
	if err != nil {
		return err
	}

	return nil
}

func (o *Output) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"content": o.Content,
	}
}

type Reply struct {
	AReply interface{} `md:"aReply"`
}

func (r *Reply) FromMap(values map[string]interface{}) error {
	r.AReply = values["aReply"]
	return nil
}

func (r *Reply) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"aReply": r.AReply,
	}
}
