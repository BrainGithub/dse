package fileserver

import (
    "database/sql"
    "encoding/json"
    "flag"
    "fmt"
    "io"
    "log"
    "net/http"
    "os"
    "path/filepath"

    _ "github.com/go-sql-driver/mysql"
    "github.com/gorilla/mux"

    "dse/internal/fileserver/common"
    "dse/internal/fileserver/service/ticketmgr"
)

var dataDir, absDataDir string
var centralDir string
var logFile, absLogFile string
var rpcPipePath string

var sboxDB *sql.DB

type appError common.AppError

type fileServerOptions struct {
    host               string
    port               uint32
    maxUploadSize      uint64
    maxDownloadDirSize uint64
    // Block size for indexing uploaded files
    fixedBlockSize uint64
    // Maximum number of goroutines to index uploaded files
    maxIndexingThreads uint32
    webTokenExpireTime uint32
    // File mode for temp files
    clusterSharedTempFileMode uint32
    windowsEncoding           string
    // Timeout for fs-id-list requests.
    fsIDListRequestTimeout uint32
}

var options fileServerOptions

func init() {
    flag.StringVar(&centralDir, "F", "", "central config directory")
    flag.StringVar(&dataDir, "d", "", "seafile data directory")
    flag.StringVar(&logFile, "l", "", "log file path")
    flag.StringVar(&rpcPipePath, "p", "", "rpc pipe path")
}

func loadSboxDB() {
    host := "127.0.0.1"
    user := "sboxweb"
    password := "Sbox123456xZ"
    dbName := "sbox_db"
    port := 3306

    unixSocket := ""
    useTLS := false

    var dsn string
    if unixSocket == "" {
        dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?tls=%t", user, password, host, port, dbName, useTLS)
    } else {
        dsn = fmt.Sprintf("%s:%s@unix(%s)/%s", user, password, unixSocket, dbName)
    }

    var err error
    sboxDB, err = sql.Open("mysql", dsn)
    if err != nil {
        log.Fatalf("Failed to open database: %v", err)
    }

}

func loadFileServerOptions() {
    initDefaultOptions()
}

func initDefaultOptions() {
    options.host = "0.0.0.0"
    options.port = 20921
    options.maxDownloadDirSize = 100 * (1 << 20)
    options.fixedBlockSize = 1 << 23
    options.maxIndexingThreads = 1
    options.webTokenExpireTime = 7200
    options.clusterSharedTempFileMode = 0600
}

func StartApp() {
    flag.Parse()

    loadSboxDB()
    ticketmgr.InitDB(sboxDB)

    loadFileServerOptions()

    if logFile == "" {
        absLogFile = filepath.Join(absDataDir, "dse_file_server.log")
        fp, err := os.OpenFile(absLogFile, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
        if err != nil {
            log.Fatalf("Failed to open or create log file: %v", err)
        }
        log.SetOutput(fp)
    }

    log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

    sizeSchedulerInit()

    router := newHTTPRouter()
    addr := fmt.Sprintf("%s:%d", options.host, options.port)

    log.Printf("dse file server: %v started.", addr)

    err := http.ListenAndServe(addr, router)
    if err != nil {
        log.Printf("File server exiting: %v", err)
    }
}

func newHTTPRouter() *mux.Router {
    r := mux.NewRouter()
    r.HandleFunc("/", handleProtocolVersion)
    r.HandleFunc("/version", handleProtocolVersion)
    r.Handle("/dse/verify", appHandler(verifySvrCB)).Methods(http.MethodPost)
    r.Handle("/dse/unmount", appHandler(unmountSvrCB)).Methods(http.MethodPost)
    r.Handle("/dse/file/get", appHandler(fileGetCB)).Methods(http.MethodPost)
    r.Handle("/dse/file/check", appHandler(fileCheckCB)).Methods(http.MethodPost)
    r.Handle("/dse/file/ferry", appHandler(fileFerryCB)).Methods(http.MethodPost)
    r.Handle("/dse/file/reload", appHandler(fileReloadCB)).Methods(http.MethodPost)
    return r
}

func handleProtocolVersion(rsp http.ResponseWriter, r *http.Request) {
    io.WriteString(rsp, "{\"version\": \"5.3\"}")
}

func verifySvrCB(rsp http.ResponseWriter, r *http.Request) *appError {
    log.Print("verifySvrCB")
    return verify(rsp, r)
}

func unmountSvrCB(rsp http.ResponseWriter, r *http.Request) *appError {
    log.Print("unmountSvrCB")
    return close(rsp, r)
}

func fileGetCB(rsp http.ResponseWriter, r *http.Request) *appError {
    return dirents(rsp, r)
}

func fileCheckCB(rsp http.ResponseWriter, r *http.Request) *appError {
    return check(rsp, r)
}

func fileFerryCB(rsp http.ResponseWriter, r *http.Request) *appError {
    return ferry(rsp, r)
}

func fileReloadCB(rsp http.ResponseWriter, r *http.Request) *appError {
    return reload(rsp, r)
}

type appHandler func(http.ResponseWriter, *http.Request) *appError

func (fn appHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

    if e := fn(w, r); e != nil {
        if e.Error != nil {
            e.Error = nil
            resp, _ := json.Marshal(e)
            w.Write(resp)
            return
        }

        if e.Error != nil && e.Code == http.StatusInternalServerError {
            log.Printf("path %s internal server error: %v\n", r.URL.Path, e.Error)
        }
        if e.Code == 0 {
            e.Code = http.StatusOK
        }
        http.Error(w, e.Message, e.Code)
    }
}
