package pow_test

import (
	"bytes"
	"proof-of-gork/pkg/pow"
	"testing"
)

func TestServer_FullFlowSuccess(t *testing.T) {
	s := pow.NewServer(&pow.Server{
		Difficulty:  2,
		NonceLength: 20,
		Secret:      []byte("qwerty"),
	})
	for i := 0; i < 3; i++ {
		rawCh, e := s.Generate()
		if e != nil {
			t.Error("error is not expected", e)
		}
		ch, e := pow.NewChallenge(rawCh)
		if e != nil {
			t.Error("error is not expected", e)
		}
		slRaw, hash, e := pow.Solve(ch, pow.DefaultHashStrategy)
		if e != nil {
			t.Error("error is not expected", e)
		}

		t.Log("REQUEST: ", rawCh)
		t.Log("HASH: ", string(hash))
		t.Log("SOLUTION: ", slRaw)

		ok, e := s.Validate(slRaw)
		if e != nil {
			t.Error("error is not expected", e)
		}
		if !ok {
			t.Error("server validation says: YOU SHALL NOT PASS")
		}
	}
}

func TestServer_FullFlowFail(t *testing.T) {
	s := pow.NewServer(&pow.Server{
		Difficulty:  2,
		NonceLength: 20,
		Secret:      []byte("qwerty"),
	})
	for i := 0; i < 3; i++ {
		rawCh, e := s.Generate()
		if e != nil {
			t.Error("error is not expected", e)
		}
		ch, e := pow.NewChallenge(rawCh)
		if e != nil {
			t.Error("error is not expected", e)
		}

		// original data mutation
		switch i {
		case 0:
			ch.Difficulty--
		case 1:
			ch.Nonce = append(ch.Nonce, byte('a'))
		case 2:
			tmp := []byte(ch.Mac)
			tmp[0] = byte('!')
			ch.Mac = string(tmp)
		}

		slRaw, _, e := pow.Solve(ch, pow.DefaultHashStrategy)
		if e != nil {
			t.Errorf("error is not expected in case %d: %s", i, e.Error())
		}

		ok, e := s.Validate(slRaw)
		if e != nil {
			t.Errorf("error is not expected in case %d", i)
		}
		if ok {
			t.Errorf("server validation says it's ok but SERVER LIES, case %d", i)
		}
	}
}

func TestNewChallengeValidReqSuccess(t *testing.T) {
	nonce, dif, mac := "11111", "22222", "33333"

	c, e := pow.NewChallenge(nonce + "-" + dif + "-" + mac)
	if e != nil {
		t.Error("error is not expected", e)
	}

	if !bytes.Equal(c.Nonce, []byte(nonce)) {
		t.Error("nonce does not match")
	}
	if c.Mac != mac {
		t.Error("mac does not match")
	}
	// a little bit of cheating with dif test value
	if c.Difficulty != uint(22222) {
		t.Error("difficulty does not match")
	}
}

func TestNewChallengeInvalidReqFail(t *testing.T) {
	nonce, dif := "11111", "22222"

	_, e := pow.NewChallenge(nonce + "-" + dif)
	if e == nil {
		t.Error("error is expected but err is nil")
	}
}

func benchmarkServerSolve(d, l uint, b *testing.B) {
	s := pow.NewServer(&pow.Server{
		Difficulty:  d,
		NonceLength: l,
		Secret:      []byte("qwerty"),
	})
	rawCh, _ := s.Generate()
	ch, _ := pow.NewChallenge(rawCh)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, e := pow.Solve(ch, pow.DefaultHashStrategy)
		if e != nil {
			b.Logf("error received: %s", e.Error())
		}
	}
}

func BenchmarkServerSolve_1_5(b *testing.B) {
	benchmarkServerSolve(1, 5, b)
}

func BenchmarkServerSolve_2_10(b *testing.B) {
	benchmarkServerSolve(2, 10, b)
}

func BenchmarkServerSolve_2_20(b *testing.B) {
	benchmarkServerSolve(2, 20, b)
}

func BenchmarkServerSolve_2_50(b *testing.B) {
	benchmarkServerSolve(2, 50, b)
}

func benchmarkServerValidate(d, l uint, b *testing.B) {
	s := pow.NewServer(&pow.Server{
		Difficulty:  d,
		NonceLength: l,
		Secret:      []byte("qwerty"),
	})
	rawCh, _ := s.Generate()
	ch, _ := pow.NewChallenge(rawCh)
	res, _, e := pow.Solve(ch, pow.DefaultHashStrategy)
	if e != nil {
		b.Logf("error received: %s", e.Error())
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ok, e := s.Validate(res)
		if e != nil {
			b.Logf("error received: %s", e.Error())
		}
		if !ok {
			b.Logf("invalid result from server")
		}
	}
}

func BenchmarkServerValidate_1_5(b *testing.B) {
	benchmarkServerValidate(1, 5, b)
}

func BenchmarkServerValidate_2_10(b *testing.B) {
	benchmarkServerValidate(2, 10, b)
}

func BenchmarkServerValidate_2_20(b *testing.B) {
	benchmarkServerValidate(2, 20, b)
}

func BenchmarkServerValidate_2_50(b *testing.B) {
	benchmarkServerValidate(2, 50, b)
}

func benchmarkServerGenerate(d, l uint, b *testing.B) {
	s := pow.NewServer(&pow.Server{
		Difficulty:  d,
		NonceLength: l,
		Secret:      []byte("qwerty"),
	})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = s.Generate()
	}
}

func BenchmarkServerGenerate_2_10(b *testing.B) {
	benchmarkServerGenerate(2, 10, b)
}

func BenchmarkServerGenerate_2_20(b *testing.B) {
	benchmarkServerGenerate(2, 20, b)
}

func BenchmarkServerGenerate_10_30(b *testing.B) {
	benchmarkServerGenerate(10, 30, b)
}

func BenchmarkServerGenerate_30_60(b *testing.B) {
	benchmarkServerGenerate(30, 60, b)
}
