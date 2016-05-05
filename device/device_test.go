package device

import "testing"

func TestGetTtyNumber(t *testing.T) {
	sample := []struct {
		src string
		num int
	}{
		{"/dev/USB0", 0},
		{"/dev/USB1", 1},
	}
	for _, v := range sample {
		n, err := getttyNum(v.src)
		if err != nil {
			t.Fatal(err)
		}
		if n != v.num {
			t.Errorf("expected %d got 5d", v.num, n)
		}
	}
}
