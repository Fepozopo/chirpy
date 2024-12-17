package api

import (
	"sync/atomic"
)

type ApiConfig struct {
	fileserverHits atomic.Int32
}
