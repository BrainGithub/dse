package common

import (
    "log"
    "os"
    "strings"
)

func AddSuffix(p string) string {
    if !strings.HasSuffix(p, "/") {
        return p + "/"
    }
    return p
}

func AddPrefix(p string) string {
    if !strings.HasPrefix(p, "/") {
        return "/" + p
    }
    return p
}

func MkSureDir(p string) {
    if _, err := os.Stat(p); err != nil {
        if err := os.MkdirAll(p, 0777); err != nil {
            log.Print(err)
        }
    }
}

func MkSureDirPerm(p string, perm os.FileMode) {
    if _, err := os.Stat(p); err != nil {
        if err := os.MkdirAll(p, perm); err != nil {
            log.Print(err)
        }
    }
}
