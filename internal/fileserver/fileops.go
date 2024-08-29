package fileserver

import (
    "dse/internal/fileserver/service/fsvrmgr"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "log"
    "net/http"
)

func verify(rsp http.ResponseWriter, r *http.Request) *appError {
    body, _ := ioutil.ReadAll(r.Body)

    v := new(fsvrmgr.FileSvrVerify)
    json.Unmarshal(body, v)

    res := appError{
        Error:  nil,
        Status: false,
    }

    fSvr := fsvrmgr.GetFSIns(v.Proto, v)

    flag, err := fSvr.Conn()
    if err != nil || !flag {
        log.Printf("verify %+v: %v, %v", v, flag, err)
        res.Error = err
        res.Message = err.Error()
        return &res
    }

    res.Status = true
    resp, _ := json.Marshal(res)
    rsp.Write(resp)

    return nil
}

func close(rsp http.ResponseWriter, r *http.Request) *appError {
    body, _ := ioutil.ReadAll(r.Body)

    v := new(fsvrmgr.FileSvrVerify)
    json.Unmarshal(body, v)

    res := appError{
        Error:  nil,
        Status: false,
    }

    fSvr := fsvrmgr.GetFSIns(v.Proto, v)

    flag, err := fSvr.Close()
    if err != nil || !flag {
        log.Printf("verify %+v: %v, %v", v, flag, err)
        res.Error = err
        res.Message = err.Error()
        return &res
    }

    res.Status = true
    resp, _ := json.Marshal(res)
    rsp.Write(resp)

    return nil
}

func dirents(rsp http.ResponseWriter, r *http.Request) *appError {
    body, _ := ioutil.ReadAll(r.Body)

    appErr := appError{
        Message: "",
        Code:    0,
        Status:  false,
    }

    v := new(fsvrmgr.Dirents)
    json.Unmarshal(body, v)

    fSvr := fsvrmgr.GetFSIns(v.Proto, v)

    ok, err := fSvr.Dirents()
    if !ok || err != nil {
        appErr.Error = err
        appErr.Message = err.Error()
        return &appErr
    }

    fInfo := fSvr.GetVar()
    res := map[string]interface{}{}
    res["fileList"] = fInfo
    res["status"] = true
    resp, _ := json.Marshal(res)
    rsp.Write(resp)

    return nil
}

func check(rsp http.ResponseWriter, r *http.Request) *appError {
    appErr := appError{
        Message: "Un-implemented error",
        Code:    0,
        Status:  false,
    }
    resp, _ := json.Marshal(appErr)
    rsp.Write(resp)
    return nil
}

func ferry(rsp http.ResponseWriter, r *http.Request) (appErr *appError) {
    body, _ := ioutil.ReadAll(r.Body)

    log.Printf("body:%s", body)

    appErr = &appError{
        Message: "",
        Code:    0,
        Status:  true,
    }

    v := new(fsvrmgr.Ferry)
    json.Unmarshal(body, v)

    if !addShedJob(fsvrmgr.GetFSIns(v.Src.Proto, v)) {
        appErr.Status = false
        appErr.Message = "ticket job queue full, too much ticket job now"
    }

    resp, _ := json.Marshal(appErr)
    rsp.Write(resp)
    return
}

func reload(rsp http.ResponseWriter, r *http.Request) *appError {
    appErr := appError{
        Message: "Un-implemented error",
        Code:    0,
        Status:  false,
    }
    resp, _ := json.Marshal(appErr)
    rsp.Write(resp)

    return nil
}
