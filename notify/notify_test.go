package notify

import (
	"testing"
)

func TestDoNotify(t *testing.T) {
	DefaultNotify.Do(`.. DD测试`)
}


