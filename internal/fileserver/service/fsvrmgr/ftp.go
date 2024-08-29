package fsvrmgr

import (
	"dse/internal/fileserver/common"
	"dse/internal/fileserver/service/fsvrmgr/ftp/zsftp"
	"dse/internal/fileserver/service/ticketmgr"
	"fmt"
	"log"
)

func MainTest() {
	//var err error
	zsftpClient := zsftp.Zsftp{"192.168.11.80", 21, "ftpuser3", "zx123456", "/"}
	rtn, fileInfoList := zsftpClient.GetFileList(false)
	fmt.Println(rtn, fileInfoList)
	remoteFile := "/w/1.txt"
	//_, filename := filepath.Split(remoteFile)
	//localDir := "D:\\test"
	//localpath := filepath.Join(localDir, filename)
	localpath1 := "D:\\test\\1.txt"
	localpath2 := "D:\\test\\2.txt"
	rtn = zsftpClient.UploadFile(remoteFile, localpath1)
	rtn = zsftpClient.UploadResumeFile(remoteFile, localpath2)
	fmt.Println(rtn)
}


type ZsftpSvr struct {
	Priv interface{}
	Var  interface{}
}

func (s *ZsftpSvr) Conn() (bool, error) {
	v := s.Priv.(FileSvrVerify)
	z := zsftp.Zsftp{v.Ip, v.Port, v.User, v.Passwd, ""}
	return z.Login()
}

func (s *ZsftpSvr) Close() (bool, error) {
	return true, nil
}

func (s *ZsftpSvr) Mount() (bool, error) {
	return true, nil
}

func (s *ZsftpSvr) UnMount() (bool, error) {
	return true, nil
}

func (s *ZsftpSvr) Dirents() (bool, error) {
	v := s.Priv.(*Dirents)
	//isDir := v.IsFileOnly > 0
	//isSubDir := v.IsRecur > 0


	dir := common.AddPrefix(v.RootPath) + common.AddPrefix(v.RelPath)
	z := zsftp.Zsftp{v.Ip, v.Port, v.User, v.Passwd, dir}
	z.GetFileList(true)

	return false, nil
}

func (s *ZsftpSvr) Read() (bool, error) {
	return true, nil
}

func (s *ZsftpSvr) Write() (bool, error) {
	return true, nil
}

func (s *ZsftpSvr) FileFerry() (bool, error) {
	log.Printf("%+v", s)

	v := s.Priv.(*Ferry)
	//src := v.Src
	//dst := v.Dst

	t := ticketmgr.Ticket{Id: v.Tid}

	//mask := syscall.Umask(0)
	//defer syscall.Umask(mask)

	tFiles, err := t.GetFerryFileList()
	if err != nil {
		return false, err
	}

	for _, tf := range tFiles {
		log.Print(tf)
	}

	return true, nil
}

func (s *ZsftpSvr) Reload() (bool, error) {
	return true, nil
}

func (s *ZsftpSvr) GetVar() interface{} {
	return s.Var
}

func (s *ZsftpSvr) SetVar(v interface{}) {
	s.Var = v
}
