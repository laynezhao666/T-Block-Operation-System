package snmp

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gosnmp/gosnmp"
)

func randomOID(k int) string {
	nums := []string{"1", "3", "6"}
	for i := 0; i < k; i++ {
		nums = append(nums, strconv.Itoa(rand.Intn(9)+1))
	}
	nums = append(nums, "0")
	return strings.Join(nums, ".")
}

func randomOIDs(n int) []string {
	oids := make([]string, n)
	for i := range oids {
		oids[i] = randomOID(10)
	}
	return oids
}

// TestGetValues test snmp get values
func TestGetValues(t *testing.T) {
	rand.Seed(time.Now().Unix())
	target := gosnmp.Default
	target.Port = 16161
	target.Target = "127.0.0.1"
	if err := target.Connect(); err != nil {
		t.Error(err)
	}
	oids := randomOIDs(target.MaxOids)
	r, err := target.Get(oids)
	if err != nil {
		t.Error(err)
	}
	if len(r.Variables) != len(oids) {
		t.Errorf("oid num: %v, result variable num: %v", len(oids), len(r.Variables))
	}
	for _, v := range r.Variables {
		fmt.Printf("%v: \t%v\n", v.Name, v.Value)
	}
}
