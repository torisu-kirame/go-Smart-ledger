package miniledgerclient

import "testing"

func TestParseStatus_heightVariants(t *testing.T) {
	cases := []struct {
		raw    string
		height uint64
	}{
		{`{"height":42,"uptime":"1h","role":"leader"}`, 42},
		{`{"height":"99","uptime":"1h"}`, 99},
		{`{"height":3.0}`, 3},
	}
	for _, tc := range cases {
		st, err := parseStatus([]byte(tc.raw))
		if err != nil {
			t.Fatalf("%s: %v", tc.raw, err)
		}
		if st.Height != tc.height {
			t.Fatalf("%s: got height %d want %d", tc.raw, st.Height, tc.height)
		}
	}
}
