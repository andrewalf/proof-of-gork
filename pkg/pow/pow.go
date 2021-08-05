package pow

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

const defaultDifficulty = 5
const defaultNonceLength = 20

// because we must have possibility to use any algorithms
// package provides only default implementations of these types
type NonceStrategy func(uint) []byte
type HashStrategy func([]byte) []byte

func init() {
	rand.Seed(time.Now().UnixNano())
}

func defaultNonceStrategy(d uint) []byte {
	chars := []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ123456789")
	res := make([]byte, d)
	for i := range res {
		res[i] = chars[rand.Intn(len(chars))]
	}
	return res
}

func DefaultHashStrategy(b []byte) []byte {
	h := sha256.Sum256(b)
	return h[:]
}

// server generates challenges and validates solutions
type Server struct {
	Difficulty  uint
	NonceLength uint
	NonceFunc   NonceStrategy
	HashFunc    HashStrategy
	Secret      []byte
}

func NewServer(c *Server) *Server {
	if c.NonceLength == 0 {
		c.NonceLength = defaultNonceLength
	}
	if c.Difficulty == 0 {
		c.Difficulty = defaultDifficulty
	}
	if c.NonceFunc == nil {
		c.NonceFunc = defaultNonceStrategy
	}
	if c.HashFunc == nil {
		c.HashFunc = DefaultHashStrategy
	}
	return c
}

func (s *Server) Generate() (string, error) {
	nonce := s.NonceFunc(s.NonceLength)
	mac, e := calculateMac(nonce, s.Secret, s.HashFunc)
	if e != nil {
		return "", e
	}
	c := &challenge{
		Nonce:      nonce,
		Mac:        mac,
		Difficulty: s.Difficulty,
	}
	return fmt.Sprintf("%s-%d-%s", c.Nonce, c.Difficulty, c.Mac), nil
}

func (s *Server) Validate(raw string) (bool, error) {
	sl, e := newSolution(raw)
	if e != nil {
		return false, e
	}
	nMac, e := calculateMac(sl.nonce, s.Secret, s.HashFunc)
	if e != nil {
		return false, e
	}
	if nMac != sl.mac {
		return false, nil
	}
	dataInt, e := strconv.Atoi(sl.data)
	if e != nil {
		return false, nil
	}
	hash := calculateHash(sl.nonce, uint32(dataInt), s.HashFunc)
	return isHashCorrect(hash, s.Difficulty), nil
}

// this is servers side entity. client resolves challenge and
// send raw string to server, where that string parsed to this entity
type solution struct {
	nonce []byte
	data  string
	mac   string
}

func newSolution(raw string) (*solution, error) {
	p := strings.SplitN(raw, "-", 3)
	if len(p) != 3 {
		return nil, errors.New("invalid solution format, it has less then 2 delimiters")
	}
	return &solution{
		nonce: []byte(p[0]),
		data:  p[1],
		mac:   p[2],
	}, nil
}

// this is clint side entity. server sends challenge as a string
// and client parses string to this entity
type challenge struct {
	Nonce      []byte
	Mac        string
	Difficulty uint
}

func NewChallenge(raw string) (*challenge, error) {
	p := strings.SplitN(raw, "-", 3)
	if len(p) != 3 {
		return nil, errors.New("invalid challenge format, it has less then 2 delimiters")
	}
	d, e := strconv.Atoi(p[1])
	if e != nil {
		return nil, e
	}
	return &challenge{
		Nonce:      []byte(p[0]),
		Difficulty: uint(d),
		Mac:        p[2],
	}, nil
}

// this can be parallelized, but this requires a little bit more time to implement properly
// hash returned just for tests, that's not necessary for the client side
func Solve(c *challenge, h HashStrategy) (string, []byte, error) {
	for i := 0; i <= math.MaxUint32; i++ {
		h := calculateHash(c.Nonce, uint32(i), h)
		if isHashCorrect(h, c.Difficulty) {
			return fmt.Sprintf("%s-%d-%s", c.Nonce, i, c.Mac), h, nil
		}
	}

	return "", nil, errors.New("hash not found in range of uint32")
}

func calculateHash(nonce []byte, data uint32, h HashStrategy) []byte {
	return h(append(nonce, intToBytes(data)...))
}

func calculateMac(nonce, secret []byte, h HashStrategy) (string, error) {
	if len(nonce) == 0 {
		return "", errors.New("empty data is impossible for mac generation")
	}
	return hex.EncodeToString(h(append(nonce, secret...))), nil
}

func isHashCorrect(hash []byte, difficulty uint) bool {
	hs := string(hash)
	if strings.HasPrefix(hs, fmt.Sprintf("%0*d", difficulty, 0)) {
		return true
	}
	return false
}

func intToBytes(i uint32) []byte {
	s := make([]byte, 4)
	binary.BigEndian.PutUint32(s, i)
	return s
}
