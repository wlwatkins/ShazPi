package structs

import "time"

type CommChannels struct {
	PlayChannel   chan bool
	RecordChannel chan time.Duration
	FetchAPI      chan bool
}
