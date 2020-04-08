// Copyright 2019 dfuse Platform Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package merger

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/dfuse-io/dstore"
	"go.uber.org/zap"
)

// findNextBaseBlock will return an error if there is a gap found ...
func (m *Merger) FindNextBaseBlock() (uint64, error) {
	prefix := highestFilePrefix(m.destStore, m.minimalBlockNum, m.chunkSize)
	zlog.Debug("find_next_base looking with prefix", zap.String("prefix", prefix))
	var lastNumber uint64
	foundAny := false
	err := m.destStore.Walk(prefix, ".tmp", func(filename string) error {
		fileNumberVal, err := strconv.ParseUint(filename, 10, 32)
		if err != nil {
			zlog.Warn("findNextBaseBlock skipping unknown file", zap.String("filename", filename))
			return nil
		}
		fileNumber := fileNumberVal
		if fileNumber < m.minimalBlockNum {
			return nil
		}
		foundAny = true

		if lastNumber == 0 {
			lastNumber = fileNumber
		} else {
			if fileNumber != lastNumber+m.chunkSize {
				return fmt.Errorf("hole was found between %d and %d", lastNumber, fileNumber)
			}
			lastNumber = fileNumber
		}
		return nil
	})
	if err != nil {
		zlog.Error("find_next_base_block found hole", zap.Error(err))
	}
	if !foundAny {
		return m.minimalBlockNum, err
	}

	return lastNumber + m.chunkSize, err
}

func getLeadingZeroes(blockNum uint64) (leadingZeros int) {
	zlog.Debug("looking for filename", zap.String("filename", fileNameForBlocksBundle(int64(blockNum))))
	for i, digit := range fileNameForBlocksBundle(int64(blockNum)) {
		if digit == '0' && leadingZeros == 0 {
			continue
		}
		if leadingZeros == 0 {
			leadingZeros = i
			return
		}
	}
	return
}

func scanForHighestPrefix(store dstore.Store, chunckSize, blockNum uint64, lastPrefix string, level int) string {
	if level == -1 {
		return lastPrefix
	}

	inc := chunckSize * uint64(math.Pow10(level))
	fmt.Println("inc:", inc)
	for {
		b := blockNum + inc
		leadingZeroes := strings.Repeat("0", getLeadingZeroes(b))
		prefix := leadingZeroes + strconv.Itoa(int(b))
		if !fileExistWithPrefix(prefix, store) {
			break
		}
		blockNum = b
		lastPrefix = prefix
	}
	return scanForHighestPrefix(store, chunckSize, blockNum, lastPrefix, level-1)
}

// findMinimalLastBaseBlocksBundle tries to minimize the number of network calls
// to storage, by trying incremental first digits, one at a time..
func highestFilePrefix(store dstore.Store, minimalBlockNum uint64, chuckSize uint64) (filePrefix string) {
	leadingZeroes := strings.Repeat("0", getLeadingZeroes(minimalBlockNum))
	blockNumStr := strconv.Itoa(int(minimalBlockNum))
	filePrefix = leadingZeroes + blockNumStr
	if !fileExistWithPrefix(filePrefix, store) {
		return
	}

	filePrefix = scanForHighestPrefix(store, chuckSize, minimalBlockNum, filePrefix, 4)
	fmt.Println("highestFilePrefix:", filePrefix)
	return
}

func fileExistWithPrefix(filePrefix string, s dstore.Store) bool {
	needZeros := 10 - len(filePrefix)
	resultFileName := filePrefix + strings.Repeat("0", needZeros)
	exists, err := s.FileExists(resultFileName)
	if err != nil {
		zlog.Error("looking for file existence on archive store", zap.Error(err))
		return false
	}
	if exists {
		return true
	}
	return false
}

func fileNameForBlocksBundle(blockNum int64) string {
	return fmt.Sprintf("%010d", blockNum)
}
