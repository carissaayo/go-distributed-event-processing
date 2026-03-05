package event

import "time"

func Enrich(evt *Event) *Event {
	if evt.Timestamp.IsZero() {
		evt.Timestamp = time.Now().UTC()
	}

	if evt.Data == nil {
		evt.Data = make(map[string]interface{})
	}

	evt.Data["_enriched_at"] = time.Now().UTC().Format(time.RFC3339)

	return evt
}
