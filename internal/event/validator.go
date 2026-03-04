package event

import "errors"

var(
	ErrInvalidEventType = errors.New("event type is required")
	ErrInvalidEventData = errors.New("event data is required")
)


func Validate(req *CreateEventRequest) error {
	if req == nil{
		return ErrInvalidEventData
	}
	if req.Type == "" {
		return ErrInvalidEventType
	}
	if req.Data == nil{
		req.Data = make(map[string]interface{})
	}
	return nil
}
