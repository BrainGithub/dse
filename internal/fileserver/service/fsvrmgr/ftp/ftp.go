package ftp

import (
	"fmt"
)

func MainTest() {
	//var err error
	zsftpClient := Zsftp{"192.168.11.80", 21, "ftpuser3", "zx123456", "/"}
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