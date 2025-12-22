// Package snowflake provides a very simple Twitter snowflake generator and parser.
package snowflake

import (
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"trpc.group/trpc-go/trpc-go/log"
)

// 一个节点数，对应一个模组。1023个节点数，可能不够用，采用减少序列号的方式，增加节点数量，其余代码不变
// 14位节点ID - 支持 16383 个节点
// 4位子节点ID - 支持 16 个子节点
// 4位序列号 - 支持 16 个序列号（一毫秒）
//
// 代码参考自 [bwmarrin/snowflake v0.3.0](https://github.com/bwmarrin/snowflake/releases/tag/v0.3.0)
// Twitter 雪花算法，由63位构成，存储在int64中
// 41位毫秒时间戳
// 10位节点ID - 支持 1023 个节点
// 12位序列号 - 支持 4095 个序列号（一毫秒）
// 节点和序列号的 22 位可以自由调整
// 如果需要更多节点，可以参考 [sony/sonyflake](https://github.com/sony/sonyflake)

var (
	// Epoch is set to the twitter snowflake epoch of Nov 04 2010 01:42:54 UTC in milliseconds
	// You may customize this to set a different epoch for your application.
	Epoch int64 = 1288834974657

	// NodeBits holds the number of bits to use for Node
	// Remember, you have a total 22 bits to share between Node/Subnode/Step
	NodeBits uint8 = 8

	// SubnodeBits 子节点 holds the number of bits to use for Subnode
	// Remember, you have a total 22 bits to share between Node/Subnode/Step
	SubnodeBits uint8 = 8

	// StepBits holds the number of bits to use for Step
	// Remember, you have a total 22 bits to share between Node/Subnode/Step
	StepBits uint8 = 6

	// DEPRECATED: the below four variables will be removed in a future release.
	// nodeMax holds the maximum Node value 16383
	nodeMax      int64 = -1 ^ (-1 << NodeBits)
	nodeMask           = nodeMax << (StepBits + SubnodeBits)
	subnodeMax   int64 = -1 ^ (-1 << SubnodeBits)
	subnodeMask        = subnodeMax << StepBits
	stepMask     int64 = -1 ^ (-1 << StepBits)
	timeShift          = NodeBits + SubnodeBits + StepBits
	nodeShift          = SubnodeBits + StepBits
	subnodeShift       = StepBits
)

const encodeBase32Map = "ybndrfg8ejkmcpqxot1uwisza345h769"

var decodeBase32Map [256]byte

const encodeBase58Map = "123456789abcdefghijkmnopqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ"

var decodeBase58Map [256]byte

// JSONSyntaxError JSONSyntaxError
// A JSONSyntaxError is returned from UnmarshalJSON if an invalid ID is provided.
type JSONSyntaxError struct{ original []byte }

// Error Error
func (j JSONSyntaxError) Error() string {
	return fmt.Sprintf("invalid snowflake ID %q", string(j.original))
}

// ErrInvalidBase58 is returned by ParseBase58 when given an invalid []byte
var ErrInvalidBase58 = errors.New("invalid base58")

// ErrInvalidBase32 is returned by ParseBase32 when given an invalid []byte
var ErrInvalidBase32 = errors.New("invalid base32")

// init init
// Create maps for decoding Base58/Base32.
// This speeds up the process tremendously.
func init() {
	for i := 0; i < len(encodeBase58Map); i++ {
		decodeBase58Map[i] = 0xFF
	}
	for i := 0; i < len(encodeBase58Map); i++ {
		decodeBase58Map[encodeBase58Map[i]] = byte(i)
	}
	for i := 0; i < len(encodeBase32Map); i++ {
		decodeBase32Map[i] = 0xFF
	}
	for i := 0; i < len(encodeBase32Map); i++ {
		decodeBase32Map[encodeBase32Map[i]] = byte(i)
	}
}

// Node Node
// A Node struct holds the basic information needed for a snowflake generator
// node
type Node struct {
	mu      sync.Mutex
	epoch   time.Time
	time    int64
	node    int64
	subnode int64
	step    int64

	nodeMax      int64
	nodeMask     int64
	subnodeMax   int64
	subnodeMask  int64
	stepMask     int64
	timeShift    uint8
	nodeShift    uint8
	subnodeShift uint8
}

// ID ID
// An ID is a custom type used for a snowflake ID.  This is used so we can
// attach methods onto the ID.
type ID int64

var (
	snowflakeNode *Node
	once          sync.Once
)

// GetNode NewNode returns a new snowflake node that can be used to generate snowflake
// IDs
func GetNode(node int64, subnode int64) (*Node, error) {
	once.Do(func() {
		n := Node{}
		n.node = node
		n.nodeMax = -1 ^ (-1 << NodeBits)
		n.nodeMask = n.nodeMax << (StepBits + SubnodeBits)
		n.subnode = subnode
		n.subnodeMax = -1 ^ (-1 << SubnodeBits)
		n.subnodeMask = n.subnodeMax << StepBits
		n.stepMask = -1 ^ (-1 << StepBits)
		n.timeShift = NodeBits + SubnodeBits + StepBits
		n.nodeShift = SubnodeBits + StepBits
		n.subnodeShift = StepBits
		if n.node < 0 || n.node > n.nodeMax {
			log.Errorf("Node number must be between 0 and " + strconv.FormatInt(n.nodeMax, 10))
			return
		}
		if n.subnode < 0 || n.subnode > n.subnodeMax {
			log.Errorf("Subnode number must be between 0 and " + strconv.FormatInt(n.subnodeMax, 10))
			return
		}
		var curTime = time.Now()
		// add time.Duration to curTime to make sure we use the monotonic clock if available
		n.epoch = curTime.Add(time.Unix(Epoch/1000, (Epoch%1000)*1000000).Sub(curTime))
		snowflakeNode = &n
	})
	if snowflakeNode == nil {
		return nil, errors.New("snowflake node is nil")
	}
	return snowflakeNode, nil
}

// Generate creates and returns a unique snowflake ID
// To help guarantee uniqueness
// - Make sure your system is keeping accurate system time
// - Make sure you never have multiple nodes running with the same node ID
func (n *Node) Generate() ID {
	n.mu.Lock()
	now := time.Since(n.epoch).Nanoseconds() / 1000000
	if now == n.time {
		n.step = (n.step + 1) & n.stepMask
		if n.step == 0 {
			for now <= n.time {
				now = time.Since(n.epoch).Nanoseconds() / 1000000
			}
		}
	} else {
		n.step = 0
	}
	n.time = now
	r := ID(
		now<<n.timeShift |
			(n.node << n.nodeShift) |
			(n.subnode << n.subnodeShift) |
			n.step,
	)
	n.mu.Unlock()
	return r
}

// Int64 returns an int64 of the snowflake ID
func (f ID) Int64() int64 {
	return int64(f)
}

// ParseInt64 converts an int64 into a snowflake ID
func ParseInt64(id int64) ID {
	return ID(id)
}

// String returns a string of the snowflake ID
func (f ID) String() string {
	return strconv.FormatInt(int64(f), 10)
}

// ParseString converts a string into a snowflake ID
func ParseString(id string) (ID, error) {
	i, err := strconv.ParseInt(id, 10, 64)
	return ID(i), err

}

// Base2 returns a string base2 of the snowflake ID
func (f ID) Base2() string {
	return strconv.FormatInt(int64(f), 2)
}

// ParseBase2 converts a Base2 string into a snowflake ID
func ParseBase2(id string) (ID, error) {
	i, err := strconv.ParseInt(id, 2, 64)
	return ID(i), err
}

// Base32 uses the z-base-32 character set but encodes and decodes similar
// to base58, allowing it to create an even smaller result string.
// NOTE: There are many different base32 implementations so becareful when
// doing any interoperation.
func (f ID) Base32() string {

	if f < 32 {
		return string(encodeBase32Map[f])
	}

	b := make([]byte, 0, 12)
	for f >= 32 {
		b = append(b, encodeBase32Map[f%32])
		f /= 32
	}
	b = append(b, encodeBase32Map[f])

	for x, y := 0, len(b)-1; x < y; x, y = x+1, y-1 {
		b[x], b[y] = b[y], b[x]
	}

	return string(b)
}

// ParseBase32 parses a base32 []byte into a snowflake ID
// NOTE: There are many different base32 implementations so becareful when
// doing any interoperation.
func ParseBase32(b []byte) (ID, error) {

	var id int64

	for i := range b {
		if decodeBase32Map[b[i]] == 0xFF {
			return -1, ErrInvalidBase32
		}
		id = id*32 + int64(decodeBase32Map[b[i]])
	}

	return ID(id), nil
}

// Base36 returns a base36 string of the snowflake ID
func (f ID) Base36() string {
	return strconv.FormatInt(int64(f), 36)
}

// ParseBase36 converts a Base36 string into a snowflake ID
func ParseBase36(id string) (ID, error) {
	i, err := strconv.ParseInt(id, 36, 64)
	return ID(i), err
}

// Base58 returns a base58 string of the snowflake ID
func (f ID) Base58() string {

	if f < 58 {
		return string(encodeBase58Map[f])
	}

	b := make([]byte, 0, 11)
	for f >= 58 {
		b = append(b, encodeBase58Map[f%58])
		f /= 58
	}
	b = append(b, encodeBase58Map[f])

	for x, y := 0, len(b)-1; x < y; x, y = x+1, y-1 {
		b[x], b[y] = b[y], b[x]
	}

	return string(b)
}

// ParseBase58 parses a base58 []byte into a snowflake ID
func ParseBase58(b []byte) (ID, error) {

	var id int64

	for i := range b {
		if decodeBase58Map[b[i]] == 0xFF {
			return -1, ErrInvalidBase58
		}
		id = id*58 + int64(decodeBase58Map[b[i]])
	}

	return ID(id), nil
}

// Base64 returns a base64 string of the snowflake ID
func (f ID) Base64() string {
	return base64.StdEncoding.EncodeToString(f.Bytes())
}

// ParseBase64 converts a base64 string into a snowflake ID
func ParseBase64(id string) (ID, error) {
	b, err := base64.StdEncoding.DecodeString(id)
	if err != nil {
		return -1, err
	}
	return ParseBytes(b)

}

// Bytes returns a byte slice of the snowflake ID
func (f ID) Bytes() []byte {
	return []byte(f.String())
}

// ParseBytes converts a byte slice into a snowflake ID
func ParseBytes(id []byte) (ID, error) {
	i, err := strconv.ParseInt(string(id), 10, 64)
	return ID(i), err
}

// IntBytes returns an array of bytes of the snowflake ID, encoded as a
// big endian integer.
func (f ID) IntBytes() [8]byte {
	var b [8]byte
	binary.BigEndian.PutUint64(b[:], uint64(f))
	return b
}

// ParseIntBytes converts an array of bytes encoded as big endian integer as
// a snowflake ID
func ParseIntBytes(id [8]byte) ID {
	return ID(int64(binary.BigEndian.Uint64(id[:])))
}

// Time returns an int64 unix timestamp in milliseconds of the snowflake ID time
// DEPRECATED: the below function will be removed in a future release.
func (f ID) Time() int64 {
	return (int64(f) >> timeShift) + Epoch
}

// Node returns an int64 of the snowflake ID node number
// DEPRECATED: the below function will be removed in a future release.
func (f ID) Node() int64 {
	return int64(f) & nodeMask >> nodeShift
}

// Subnode 子节点
func (f ID) Subnode() int64 {
	return int64(f) & subnodeMask >> subnodeShift
}

// Step returns an int64 of the snowflake step (or sequence) number
// DEPRECATED: the below function will be removed in a future release.
func (f ID) Step() int64 {
	return int64(f) & stepMask
}

// MarshalJSON returns a json byte array string of the snowflake ID.
func (f ID) MarshalJSON() ([]byte, error) {
	buff := make([]byte, 0, 22)
	buff = append(buff, '"')
	buff = strconv.AppendInt(buff, int64(f), 10)
	buff = append(buff, '"')
	return buff, nil
}

// UnmarshalJSON converts a json byte array of a snowflake ID into an ID type.
func (f *ID) UnmarshalJSON(b []byte) error {
	if len(b) < 3 || b[0] != '"' || b[len(b)-1] != '"' {
		return JSONSyntaxError{b}
	}

	i, err := strconv.ParseInt(string(b[1:len(b)-1]), 10, 64)
	if err != nil {
		return err
	}

	*f = ID(i)
	return nil
}
