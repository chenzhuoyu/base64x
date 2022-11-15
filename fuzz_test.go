package base64x

import (
    `encoding/base64`
    `encoding/json`
    `testing`
    `github.com/stretchr/testify/require`
    `github.com/davecgh/go-spew/spew`
)

func FuzzBase64(f *testing.F) {
    var corpus = []string {
        "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/",
        "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_",
        "=",
        `\/`,
        "\r\n",
        `\r\n`,
        `\u0036`, `\u0039`, `\u003d`,
        `"\u0036"`, `"\u003d\u003d"`,
    }
    for _, c := range(corpus) {
        f.Add([]byte(c))
    }
    f.Fuzz(fuzzBase64Impl)
}

func fuzzBase64Impl(t *testing.T, data []byte) {
    fuzzBase64CommonImpl(t, data)
    fuzzBase64JsonImpl(t, data)
}

type EncodeFuzzPairs struct {
    ours      Encoding
    stdlib    *base64.Encoding
}

var fuzzPairs = []EncodeFuzzPairs {
    {StdEncoding, base64.StdEncoding},
    {URLEncoding, base64.URLEncoding},
    {RawStdEncoding, base64.RawStdEncoding},
    {RawURLEncoding, base64.RawURLEncoding},
}

func fuzzBase64CommonImpl(t *testing.T, data []byte) {
    for _, fp := range(fuzzPairs) {
        // fuzz encode
        encoded0 := fp.ours.EncodeToString(data)
        encoded1 := fp.stdlib.EncodeToString(data)
        require.Equalf(t, encoded0, encoded1, "encode from %s", spew.Sdump(data))
        // fuzz decode
        encoded := encoded1
        dbuf0 := make([]byte, fp.ours.DecodedLen(len(encoded)))
        dbuf1 := make([]byte, fp.stdlib.DecodedLen(len(encoded)))
        count0, err0 := fp.ours.Decode(dbuf0, []byte(encoded))
        count1, err1 := fp.stdlib.Decode(dbuf1, []byte(encoded))
        require.Equalf(t, dbuf0, dbuf1, "decode from %s", spew.Sdump(encoded))
        require.Equalf(t, err0 != nil, err1 != nil, "decode from %s", spew.Sdump(encoded))
        require.Equalf(t, count0, count1, "decode from %s", spew.Sdump(encoded))
    }
}

func fuzzBase64JsonImpl(t *testing.T, data []byte) {
    // fuzz valid JSON-encoded base64
    jencoded, _ := json.Marshal(data)
    var dbuf0, dbuf1 []byte
    dbuf0 = make([]byte,  JSONStdEncoding.DecodedLen(len(jencoded)))
    count0, err0 := JSONStdEncoding.Decode(dbuf0, jencoded[1:len(jencoded) - 1])
    err1 := json.Unmarshal(jencoded, &dbuf1)
    require.Equalf(t, dbuf0[:count0], dbuf1,  "decode json from %s", spew.Sdump(jencoded))
    require.Equalf(t, err0 != nil, err1 != nil, "decode json from %s", spew.Sdump(jencoded))
}