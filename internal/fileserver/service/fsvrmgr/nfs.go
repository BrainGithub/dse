package fsvrmgr

import (
    "dse/internal/fileserver/common"
    "dse/internal/fileserver/service/blockmgr"
    "dse/internal/fileserver/service/ticketmgr"
    "fmt"
    "io"
    "io/ioutil"
    "log"
    "os"
    "os/exec"
    "path/filepath"
    //"syscall"
)

const (
    Mount   = "timeout 15 mount -t %s -o rw,soft,timeo=21,retry=3 %s %s"
    UnMount = "timeout 15 umount -f %s"
    MountLs = "timeout 15 mount | grep \"%s on %s type nfs\""
)

type NfsSvr struct {
    Priv interface{}
    Var  interface{}
}

func (s *NfsSvr) Conn() (bool, error) {
    return s.Mount()
}

func (s *NfsSvr) Close() (bool, error) {
    return s.UnMount()
}

func (s *NfsSvr) Mount() (bool, error) {
    v := s.Priv.(*FileSvrVerify)

    if _, err := os.Stat(v.Dst); err != nil {
        if err := os.MkdirAll(v.Dst, os.ModePerm); err != nil {
            log.Printf("MkdirAll:%s, error:%+v", v.Dst, err)
            return false, err
        }
    }

    mountls := fmt.Sprintf(MountLs, v.Src, v.Dst)
    cmd := exec.Command("/bin/bash", "-c", mountls)
    output, err := cmd.Output()
    log.Printf("%s, error:%+v, output:%s.", mountls, err, output)

    if string(output) != "" {
        return true, nil
    }

    mount := fmt.Sprintf(Mount, v.Proto, v.Ip+":"+v.Src, v.Dst)
    cmd = exec.Command("/bin/bash", "-c", mount)
    output, err = cmd.Output()
    log.Printf("%s, error:%+v, output:%s.", mount, err, output)

    if err != nil {
        return false, err
    }

    return true, nil
}

func (s *NfsSvr) UnMount() (bool, error) {
    v := s.Priv.(*FileSvrVerify)

    mount := fmt.Sprintf(UnMount, v.Dst)
    cmd := exec.Command("/bin/bash", "-c", mount)
    output, err := cmd.Output()
    log.Printf("%s, error:%+v, output:%s", mount, err, output)

    if err != nil {
        return false, err
    }

    return true, nil
}

func (s *NfsSvr) Dirents() (bool, error) {
    v := s.Priv.(*Dirents)

    absPath := filepath.Join(v.RootPath, v.RelPath)
    isDirOnly := v.IsDirOnly > 0
    isRecur := v.IsRecur > 0

    var fileList []map[string]interface{}
    var relDirs = []string{v.RelPath}

    for {
        if len(relDirs) <= 0 {
            break
        }

        dirs := relDirs
        relDirs = []string{}
        for _, p := range dirs {
            infs, err := ioutil.ReadDir(filepath.Join(absPath, p))
            if err != nil {
                log.Printf("read dir error:%+v", err)
                return false, err
            }

            for _, i := range infs {
                f := map[string]interface{}{}
                f["name"] = i.Name()
                f["size"] = i.Size()
                //f["createTime"] = time.Unix(i.Sys().(*syscall.Stat_t).Ctim.Sec, 0)
                relPath := filepath.Join(p, i.Name())
                f["relPath"] = relPath

                if i.IsDir() {
                    f["type"] = "dir"
                    relDirs = append(relDirs, relPath)
                } else {
                    if !isDirOnly {
                        f["type"] = "file"
                    }
                }

                fileList = append(fileList, f)
            }
        }

        if !isRecur {
            break
        }
    }

    log.Printf("dirents:%+v", fileList)
    s.Var = fileList

    return true, nil
}

func (s *NfsSvr) Read() (bool, error) {
    return true, nil
}

func (s *NfsSvr) Write() (bool, error) {
    return true, nil
}

func read(c *blockmgr.Chunk) (bool, error) {
    f, err := os.Open(c.Src)
    if err != nil {
        return false, err
    }
    defer f.Close()

    l, err := f.ReadAt(c.Buf[:], c.Start)
    if err != nil && err != io.EOF {
        return false, nil
    }

    c.Len = l

    return true, nil
}

func write(c *blockmgr.Chunk) (bool, error) {
    f, err := os.OpenFile(c.Dst, os.O_WRONLY|os.O_CREATE, c.Mode)
    if err != nil {
        return false, err
    }
    defer f.Close()

    buf := c.Buf[:c.Len]
    l, err := f.WriteAt(buf[:c.Len], c.Start)
    if err != nil {
        return false, err
    }

    if l != c.Len {
        log.Printf("to write:%d, wrote:%d", c.Len, l)
        return false, err
    }

    return true, nil
}

func (s *NfsSvr) FileFerry() (bool, error) {
    log.Printf("%+v", s)

    v := s.Priv.(*Ferry)
    src := v.Src
    dst := v.Dst

    t := ticketmgr.Ticket{Id: v.Tid}
    cap := blockmgr.BLK_SIZE_8M
    c := blockmgr.Chunk{Buf: make([]byte, blockmgr.BLK_SIZE_8M)}

    //mask := syscall.Umask(0)
    //defer syscall.Umask(mask)

    tFiles, err := t.GetFerryFileList()
    if err!= nil {
        return false, err
    }

    for _, tf := range tFiles {
        path := tf.Path
        t.Path = path

        sf, df := t.GetFilePath(src.RootPath, dst.RootPath, "", dst.RelPath, src.Typ, dst.Typ)

        log.Printf("ferry from:%s, to:%s", sf, df)

        sfInfo, err := os.Stat(sf)
        if err != nil {
            log.Printf("src file:%s does not exist", sf)
            continue
        } else {
            if sfInfo.IsDir() {
                common.MkSureDirPerm(df, sfInfo.Mode())
                os.Chmod(df, sfInfo.Mode())
                continue
            } else {
                sfDirInfo, _ := os.Stat(filepath.Dir(sf))
                common.MkSureDirPerm(filepath.Dir(df), sfDirInfo.Mode())
            }
        }

        c.Src, c.Dst, c.Mode, c.Start = sf, df, sfInfo.Mode(), tf.TransferSize

        for {
            if ok, err := read(&c); !ok || err != nil {
                return false, err
            }

            if c.Len < cap {
                if ok, err := write(&c); !ok || err != nil {
                    return false, err
                }

                c.Start += int64(c.Len)
                c.Len = 0

                t.TransferSize = c.Start

                if _, err := t.OnFinished(); err != nil {
                    return false, err
                }

                os.Chmod(df, c.Mode)
                break
            }

            if ok, err := write(&c); !ok || err != nil {
                return false, err
            }

            c.Start += int64(c.Len)
            c.Len = 0

            t.TransferSize = c.Start
            if _, err := t.OnProcess(); err != nil {
                return false, err
            }
        }
    }

    return true, nil
}

func (s *NfsSvr) Reload() (bool, error) {
    return true, nil
}

func (s *NfsSvr) GetVar() interface{} {
    return s.Var
}

func (s *NfsSvr) SetVar(v interface{}) {
    s.Var = v
}
