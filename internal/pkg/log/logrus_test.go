package log

import "testing"

func TestLogrusAdaptor(t *testing.T) {
	items := []struct {
		level  int
		format string
	}{
		{0, "text"},
		{1, "text"},
		{2, "text"},
		{3, "text"},
		{4, "text"},
		{5, "text"},
		{6, "text"},
		{0, "json"},
		{1, "json"},
		{2, "json"},
		{3, "json"},
		{4, "json"},
		{5, "json"},
		{6, "json"},
	}
	for _, v := range items {
		_, err := NewLogrusAdaptor(v.level, v.format)
		if err != nil {
			t.Errorf("NewLogrusAdaptor: %s", err)
			break
		}
	}
}
