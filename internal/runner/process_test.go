package runner

import "testing"

func TestLimitedBufferTruncates(t *testing.T) {
	t.Parallel()
	buffer := newLimitedBuffer(4)
	written, err := buffer.Write([]byte("abcdefgh"))
	if err != nil {
		t.Fatal(err)
	}
	if written != 8 {
		t.Fatalf("reported write=%d want=8", written)
	}
	if got := string(buffer.Bytes()); got != "abcd\n[output truncated by PlanParity]\n" {
		t.Fatalf("unexpected buffer: %q", got)
	}
}
