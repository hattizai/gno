//go:build cleveldb

package db

import (
	"bytes"
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func BenchmarkRandomReadsWrites2(b *testing.B) {
	b.StopTimer()

	numItems := int64(1000000)
	internal := map[int64]int64{}
	for i := 0; i < int(numItems); i++ {
		internal[int64(i)] = int64(0)
	}
	db, err := NewCLevelDB(fmt.Sprintf("test_%x", randStr(12)), "")
	if err != nil {
		b.Fatal(err.Error())
		return
	}

	fmt.Println("ok, starting")
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		// Write something
		{
			idx := (int64(rand.Int()) % numItems)
			internal[idx]++
			val := internal[idx]
			idxBytes := int642Bytes(int64(idx))
			valBytes := int642Bytes(int64(val))
			// fmt.Printf("Set %X -> %X\n", idxBytes, valBytes)
			db.Set(
				idxBytes,
				valBytes,
			)
		}
		// Read something
		{
			idx := (int64(rand.Int()) % numItems)
			val := internal[idx]
			idxBytes := int642Bytes(int64(idx))
			valBytes := db.Get(idxBytes)
			// fmt.Printf("Get %X -> %X\n", idxBytes, valBytes)
			if val == 0 {
				if !bytes.Equal(valBytes, nil) {
					b.Errorf("Expected %v for %v, got %X",
						nil, idx, valBytes)
					break
				}
			} else {
				if len(valBytes) != 8 {
					b.Errorf("Expected length 8 for %v, got %X",
						idx, valBytes)
					break
				}
				valGot := bytes2Int64(valBytes)
				if val != valGot {
					b.Errorf("Expected %v for %v, got %v",
						val, idx, valGot)
					break
				}
			}
		}
	}

	db.Close()
}

/*
func int642Bytes(i int64) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(i))
	return buf
}

func bytes2Int64(buf []byte) int64 {
	return int64(binary.BigEndian.Uint64(buf))
}
*/

func TestCLevelDBBackend(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("test_%x", randStr(12))
	// Can't use "" (current directory) or "./" here because levigo.Open returns:
	// "Error initializing DB: IO error: test_XXX.db: Invalid argument"
	db, err := NewDB(name, CLevelDBBackend, t.TempDir())
	require.NoError(t, err)

	_, ok := db.(*CLevelDB)
	assert.True(t, ok)
}

func TestCLevelDBStats(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("test_%x", randStr(12))
	db, err := NewDB(name, CLevelDBBackend, t.TempDir())
	require.NoError(t, err)

	assert.NotEmpty(t, db.Stats())
}
