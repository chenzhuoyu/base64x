package base64x

import (
    `crypto/rand`
    `encoding/base64`
    `io`
    `strings`
    `testing`
)

type TestPair struct {
    decoded string
    encoded string
}

type EncodingTest struct {
    enc  Encoding            // Encoding to test
    conv func(string) string // Reference string converter
}

var pairs = []TestPair{
    // RFC 3548 examples
    {"\x14\xfb\x9c\x03\xd9\x7e", "FPucA9l+"},
    {"\x14\xfb\x9c\x03\xd9", "FPucA9k="},
    {"\x14\xfb\x9c\x03", "FPucAw=="},

    // RFC 4648 examples
    {"", ""},
    {"f", "Zg=="},
    {"fo", "Zm8="},
    {"foo", "Zm9v"},
    {"foob", "Zm9vYg=="},
    {"fooba", "Zm9vYmE="},
    {"foobar", "Zm9vYmFy"},

    // Wikipedia examples
    {"sure.", "c3VyZS4="},
    {"sure", "c3VyZQ=="},
    {"sur", "c3Vy"},
    {"su", "c3U="},
    {"leasure.", "bGVhc3VyZS4="},
    {"easure.", "ZWFzdXJlLg=="},
    {"asure.", "YXN1cmUu"},
    {"sure.", "c3VyZS4="},

    // Relatively long strings
    {
        "Twas brillig, and the slithy toves",
        "VHdhcyBicmlsbGlnLCBhbmQgdGhlIHNsaXRoeSB0b3Zlcw==",
    }, {
        "\x9dyH\xd2Y\x9e^e\x9e\xb1\x9a\x9a\x12\xfe\x8a\x07\xc7\x07\xcc\xe8l\x81" +
        "\xf2\xd9\xe3\x89\xb5\x98\xee\xbd\x8etQ`2>\\t:_\xd7w\xe6\xb5\x96\xc7\xff\x9c",
        "nXlI0lmeXmWesZqaEv6KB8cHzOhsgfLZ44m1mO69jnRRYDI+XHQ6X9d35rWWx/+c",
    },
}

// Do nothing to a reference base64 string (leave in standard format)
func stdRef(ref string) string {
    return ref
}

// Convert a reference string to URL-encoding
func urlRef(ref string) string {
    ref = strings.ReplaceAll(ref, "+", "-")
    ref = strings.ReplaceAll(ref, "/", "_")
    return ref
}

// Convert a reference string to raw, unpadded format
func rawRef(ref string) string {
    return strings.TrimRight(ref, "=")
}

// Both URL and unpadding conversions
func rawURLRef(ref string) string {
    return rawRef(urlRef(ref))
}

var encodingTests = []EncodingTest{
    {StdEncoding, stdRef},
    {URLEncoding, urlRef},
    {RawStdEncoding, rawRef},
    {RawURLEncoding, rawURLRef},
}

func testEqual(t *testing.T, msg string, args ...interface{}) bool {
    t.Helper()
    if args[len(args) - 2] != args[len(args) - 1] {
        t.Errorf(msg, args...)
        return false
    }
    return true
}

func TestEncoder(t *testing.T) {
    for _, p := range pairs {
        for _, tt := range encodingTests {
            got := tt.enc.EncodeToString([]byte(p.decoded))
            testEqual(t, "Encode(%q) = %q, want %q", p.decoded, got, tt.conv(p.encoded))
        }
    }
}

func benchmarkStdlibWithSize(b *testing.B, nb int) {
    buf := make([]byte, nb)
    dst := make([]byte, base64.StdEncoding.EncodedLen(nb))
    _, _ = io.ReadFull(rand.Reader, buf)
    b.SetBytes(int64(nb))
    b.ResetTimer()
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            base64.StdEncoding.Encode(dst, buf)
        }
    })
}

func benchmarkBase64xWithSize(b *testing.B, nb int) {
    buf := make([]byte, nb)
    dst := make([]byte, StdEncoding.EncodedLen(nb))
    _, _ = io.ReadFull(rand.Reader, buf)
    b.SetBytes(int64(nb))
    b.ResetTimer()
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            StdEncoding.Encode(dst, buf)
        }
    })
}

func BenchmarkEncoderStdlib_16B    (b *testing.B) { benchmarkStdlibWithSize(b, 16) }
func BenchmarkEncoderStdlib_56B    (b *testing.B) { benchmarkStdlibWithSize(b, 56) }
func BenchmarkEncoderStdlib_128B   (b *testing.B) { benchmarkStdlibWithSize(b, 128) }
func BenchmarkEncoderStdlib_4kB    (b *testing.B) { benchmarkStdlibWithSize(b, 4 * 1024) }
func BenchmarkEncoderStdlib_256kB  (b *testing.B) { benchmarkStdlibWithSize(b, 256 * 1024) }
func BenchmarkEncoderStdlib_1MB    (b *testing.B) { benchmarkStdlibWithSize(b, 1024 * 1024) }

func BenchmarkEncoderBase64x_16B   (b *testing.B) { benchmarkBase64xWithSize(b, 16) }
func BenchmarkEncoderBase64x_56B   (b *testing.B) { benchmarkBase64xWithSize(b, 56) }
func BenchmarkEncoderBase64x_128B  (b *testing.B) { benchmarkBase64xWithSize(b, 128) }
func BenchmarkEncoderBase64x_4kB   (b *testing.B) { benchmarkBase64xWithSize(b, 4 * 1024) }
func BenchmarkEncoderBase64x_256kB (b *testing.B) { benchmarkBase64xWithSize(b, 256 * 1024) }
func BenchmarkEncoderBase64x_1MB   (b *testing.B) { benchmarkBase64xWithSize(b, 1024 * 1024) }

func TestDecoder(t *testing.T) {
    for _, p := range pairs {
        for _, tt := range encodingTests {
            encoded := tt.conv(p.encoded)
            dbuf := make([]byte, tt.enc.DecodedLen(len(encoded)))
            count, err := tt.enc.Decode(dbuf, []byte(encoded))
            testEqual(t, "Decode(%q) = error %v, want %v", encoded, err, error(nil))
            testEqual(t, "Decode(%q) = length %v, want %v", encoded, count, len(p.decoded))
            testEqual(t, "Decode(%q) = %q, want %q", encoded, string(dbuf[0:count]), p.decoded)

            dbuf, err = tt.enc.DecodeString(encoded)
            testEqual(t, "DecodeString(%q) = error %v, want %v", encoded, err, error(nil))
            testEqual(t, "DecodeString(%q) = %q, want %q", encoded, string(dbuf), p.decoded)
        }
    }
}

func TestDecoderError(t *testing.T) {
    _, err := StdEncoding.DecodeString("!aGVsbG8sIHdvcmxk")
    if err != base64.CorruptInputError(0) {
        panic(err)
    }
    _, err = StdEncoding.DecodeString("aGVsbG8!sIHdvcmxk")
    if err != base64.CorruptInputError(7) {
        panic(err)
    }
}

func benchmarkStdlibDecoder(b *testing.B, v string) {
    src := []byte(v)
    dst := make([]byte, base64.StdEncoding.DecodedLen(len(v)))
    b.SetBytes(int64(len(v)))
    b.ResetTimer()
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            _, _ = base64.StdEncoding.Decode(dst, src)
        }
    })
}

func benchmarkBase64xDecoder(b *testing.B, v string) {
    src := []byte(v)
    dst := make([]byte, StdEncoding.DecodedLen(len(v)))
    b.SetBytes(int64(len(v)))
    b.ResetTimer()
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            _, _ = StdEncoding.Decode(dst, src)
        }
    })
}

var data = `////////////////////////////////////////////////////////////////`
func BenchmarkDecoderStdLib  (b *testing.B) { benchmarkStdlibDecoder(b, data) }
func BenchmarkDecoderBase64x (b *testing.B) { benchmarkBase64xDecoder(b, data) }
