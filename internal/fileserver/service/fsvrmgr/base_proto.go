package fsvrmgr

type FileSvrVerify struct {
    Ip     string `json:"ip"`
    Port   int    `json:"port"`
    Proto  string `json:"protocol"`
    User   string `json:"userName"`
    Passwd string `json:"passwd"`
    Src    string `json:"devDir"`
    Dst    string `json:"mountDir"`
}

type Dirents struct {
    Ip         string `json:"ip"`
    Port       int    `json:"port"`
    Proto      string `json:"protocol"`
    User       string `json:"userName"`
    Passwd     string `json:"passwd"`
    RelPath    string `json:"relPath"`
    RootPath   string `json:"rootPath"`
    IsDirOnly  int    `json:"dirFlag"`
    IsRecur    int    `json:"isRec"`
}

type SrcSvrInfo struct {
    Typ      string   `json:"serverType"`
    Ip       string   `json:"ip"`
    Port     int      `json:"port"`
    Proto    string   `json:"protocol"`
    User     string   `json:"userName"`
    Passwd   string   `json:"passwd"`
    RelPath  []string `json:"relPathList"`
    RootPath string   `json:"rootPath"`
}

type DstSvrInfo struct {
    Typ      string `json:"serverType"`
    Ip       string `json:"ip"`
    Port     int    `json:"port"`
    Proto    string `json:"protocol"`
    User     string `json:"userName"`
    Passwd   string `json:"passwd"`
    RelPath  string `json:"relPath"`
    RootPath string `json:"rootPath"`
}

type Ferry struct {
    Tid int     `json:"ticket_id"`
    Src SrcSvrInfo `json:"src_server_info"`
    Dst DstSvrInfo `json:"dst_server_info"`
}

var ProtoSupport = map[string]bool {
    "nfs": true,
    "samba": true,
}

type FileSvr interface {
    Conn() (bool, error)
    Close() (bool, error)
    Mount() (bool, error)
    UnMount() (bool, error)
    Dirents() (bool, error)
    Read() (bool, error)
    Write() (bool, error)
    FileFerry() (bool, error)
    Reload() (bool, error)
    SetVar(interface{})
    GetVar() interface{}
}

func GetFSIns(t string, v interface{}) FileSvr {
    var s FileSvr
    switch t {
    case "smaba":
        fallthrough
    case "nfs":
        s = &NfsSvr{Priv: v, Var: nil}
    case "ftp":
    case "scp":
    case "rsync":
    }
    return s
}
