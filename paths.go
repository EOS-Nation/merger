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
	"strconv"
	"strings"
	"time"
)

func blockNumToStr(blockNum uint64) (blockStr string) {
	return fmt.Sprintf("%010d", blockNum)
}

// parseFilename parses file names formatted like:
// * 0000000100-20170701T122141.0-24a07267-e5914b39
// * 0000000101-20170701T122141.5-dbda3f44-09f6d693
func parseFilename(filename string) (blockNum uint64, blockTime time.Time, blockIDSuffix string, previousBlockIDSuffix string, err error) {
	parts := strings.Split(filename, "-")
	if len(parts) != 4 {
		err = fmt.Errorf("wrong filename format: %q", filename)
		return
	}
	blockNumVal, err := strconv.ParseUint(parts[0], 10, 32)
	if err != nil {
		err = fmt.Errorf("failed parsing %q: %s", parts[0], err)
		return
	}
	blockNum = uint64(blockNumVal)

	blockTime, err = time.Parse("20060102T150405.999999", parts[1])
	if err != nil {
		err = fmt.Errorf("failed parsing %q: %s", parts[1], err)
		return
	}

	blockIDSuffix = parts[2]
	previousBlockIDSuffix = parts[3]
	return
}
