package blockmgr

import (
    "os"
)

const (
    BLK_SIZE_8M  = 8 * 1024 * 1024
    BLK_SIZE_10M = 10 * 1024 * 1024
)

type Chunk struct {
    Key   string
    Start int64
    Len   int
    Buf   []byte
    Src   string
    Dst   string
    Mode  os.FileMode
}


