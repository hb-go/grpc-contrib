package consul

import (
	"testing"
)

func TestEncodingVersion(t *testing.T) {
	testData := []struct {
		decoded string
		encoded string
	}{
		{"1.0.0", "v-789c32d433d03300040000ffff02ce00ee"},
		{"latest", "v-789cca492c492d2e01040000ffff08cc028e"},
	}

	for _, data := range testData {
		e := encodeVersion(data.decoded)

		if e[0] != data.encoded {
			t.Fatalf("Expected %s got %s", data.encoded, e)
		}

		d, ok := decodeVersion(e)
		if !ok {
			t.Fatalf("Unexpected %t for %s", ok, data.encoded)
		}

		if d != data.decoded {
			t.Fatalf("Expected %s got %s", data.decoded, d)
		}

		d, ok = decodeVersion([]string{data.encoded})
		if !ok {
			t.Fatalf("Unexpected %t for %s", ok, data.encoded)
		}

		if d != data.decoded {
			t.Fatalf("Expected %s got %s", data.decoded, d)
		}
	}
}
