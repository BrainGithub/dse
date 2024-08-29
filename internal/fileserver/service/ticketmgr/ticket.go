package ticketmgr

import (
    "database/sql"
    "dse/internal/fileserver/common"
    "fmt"
    "log"
    "path/filepath"
    "strings"
)

type Ticket struct {
    Id int
    Path string
    Name string
    Size int64
    TransferSize int64
}

var db *sql.DB

func InitDB(d *sql.DB) {
    db = d
}

func (t Ticket)OnProcess() (bool, error) {
    sql := fmt.Sprintf("update tab_ticket_files set transfer_size=%d where ticket_id=%d and path='%s'",
        t.TransferSize, t.Id, t.Path)
    res, err := db.Exec(sql)
    log.Printf("err:%+v, sql:%s, res:%+v", err, sql, res)
    if err != nil {
        return false, nil
    }
    return true, nil
}

func (t Ticket)OnFinished() (bool, error) {
    sql := fmt.Sprintf("update tab_ticket_files set state=if(transfer_size=size, 0, 2) where ticket_id=%d and path='%s'",
        t.Id, t.Path)
    res, err := db.Exec(sql)
    log.Printf("err:%+v, sql:%s, res:%+v", err, sql, res)
    if err != nil {
        return false, nil
    }
    return true, nil
}

func (t Ticket)GetFerryFileList() ([]Ticket, error) {
    var tList []Ticket
    sql := fmt.Sprintf("select path, transfer_size from tab_ticket_files "+
        "where ticket_id=%d and state!=0", t.Id)
    rows, err := db.Query(sql)
    log.Printf("err:%+v, sql:%s, res:%+v", err, sql, rows)
    if err != nil {
        return tList, err
    }

    defer rows.Close()

    for rows.Next() {
        t := Ticket{Id: t.Id}
        if err := rows.Scan(&t.Path, &t.TransferSize); err != nil {
            log.Printf("err:%+v", err)
            return tList, err
        }

        tList = append(tList, t)
    }

    return tList, nil
}

func (t Ticket)GetFilePath(sfRoot, dfRoot, sfRel, dfRel, sTyp, dTye string) (sf, df string) {
    path := t.Path
    pathArr := strings.SplitN(path, "/", 5)
    rawFile := pathArr[4]

    switch dTye {
    case "server":
        df = filepath.Join(dfRoot, dfRel, rawFile)
    case "local":
        df = path
        common.MkSureDir(filepath.Dir(df))
    }

    switch sTyp {
    case "server":
        sf = filepath.Join(sfRoot, sfRel, rawFile)
    case "local":
        sf = path
    }
    return
}
