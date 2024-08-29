package ftp

import (
	ftp4go2 "dse/internal/fileserver/service/fsvrmgr/ftp/lib/ftp4go"
	"fmt"
	"path"
	"strings"
)

// FTP file info index
type InfoIndex int
const (
	FILE_TYPE_INDEX    InfoIndex = 0
	FILE_SIZE_INDEX    InfoIndex = 4
	FILE_NAME_INDEX    InfoIndex = 8
)

//Return file info field
const (
	FILE_INFO_NAME string = "name"
	FILE_INFO_SIZE string = "size"
	FILE_INFO_TYPE string = "type"
	FILE_INFO_REALPATH string = "realPath"
	FILE_INFO_TIME string = "createTime"
)

const (
	FILE_TYPE_DIR string = "dir"
	FILE_TYPE_FILE string = "file"
)

type Zsftp struct {
	Ip string
	Port int
	Username string
	Password string
	Dir string
}

func getFileType(typeString string) string {
	fileType := ""
	dirFlag := strings.HasPrefix(typeString, "d")
	if dirFlag {
		fileType = FILE_TYPE_DIR
	} else {
		fileType = FILE_TYPE_FILE
	}

	return fileType
}

func (z Zsftp) GetFileList(includeFile bool) (bool, []map[string]string) {
	var err error
	var rtn bool = false
	var fileInfoList []map[string]string

	ftpClient := ftp4go2.NewFTP(0)
	//connect
	_, err = ftpClient.Connect(z.Ip, z.Port, "")
	if err != nil {
		return rtn, fileInfoList
	}

	defer ftpClient.Quit()

	_, err = ftpClient.Login(z.Username, z.Password, "")
	if err != nil {
		return rtn, fileInfoList
	}

	fileList, err := ftpClient.Dir(z.Dir)
	if err != nil {
		return rtn, fileInfoList
	}

	for i := 0; i < len(fileList); i++ {
		infos := strings.Fields(strings.TrimSpace(fileList[i]))
		fileType := getFileType(infos[FILE_TYPE_INDEX])

		if FILE_TYPE_FILE == fileType && !includeFile {
			continue
		}

		fileInfo := map[string]string{
			FILE_INFO_NAME: infos[FILE_NAME_INDEX],
			FILE_INFO_SIZE: infos[FILE_SIZE_INDEX],
			FILE_INFO_TYPE: fileType,
			FILE_INFO_REALPATH: path.Join(z.Dir, infos[FILE_NAME_INDEX]),
			FILE_INFO_TIME: "",
		}
		fileInfoList = append(fileInfoList, fileInfo)
	}
	rtn = true

	return rtn, fileInfoList
}

func (z Zsftp) DownloadFile(remoteFile, localPath string) bool {
	var err error
	var rtn bool = false

	ftpClient := ftp4go2.NewFTP(0)
	//connect
	_, err = ftpClient.Connect(z.Ip, z.Port, "")
	if err != nil {
		return rtn
	}

	defer ftpClient.Quit()

	_, err = ftpClient.Login(z.Username, z.Password, "")
	if err != nil {
		return rtn
	}

	err = ftpClient.DownloadFile(remoteFile, localPath, false)
	if err == nil {
		rtn = true
	}

	return rtn
}

func (z Zsftp) DownloadResumeFile(remoteFile, localPath string) bool {
	var err error
	var rtn bool = false

	ftpClient := ftp4go2.NewFTP(0)
	//connect
	_, err = ftpClient.Connect(z.Ip, z.Port, "")
	if err != nil {
		return rtn
	}

	defer ftpClient.Quit()

	_, err = ftpClient.Login(z.Username, z.Password, "")
	if err != nil {
		return rtn
	}

	// get the remote file size
	_, err = ftpClient.Size(remoteFile)
	if err != nil {
		return rtn
	}

	err = ftpClient.DownloadResumeFile(remoteFile, localPath, false)
	if err == nil {
		rtn = true
	}

	return rtn
}

func updateProgress(info *ftp4go2.CallbackInfo) {
	fmt.Println(info.Filename, info.Resourcename, info.BytesTransmitted, info.Eof)
}

func (z Zsftp) UploadFile(remoteFile, localPath string) bool {
	var err error
	var rtn = false

	ftpClient := ftp4go2.NewFTP(0)
	//connect
	_, err = ftpClient.Connect(z.Ip, z.Port, "")
	if err != nil {
		return rtn
	}

	defer ftpClient.Quit()

	_, err = ftpClient.Login(z.Username, z.Password, "")
	if err != nil {
		return rtn
	}

	err = ftpClient.UploadFile(remoteFile, localPath, false, updateProgress)
	if err == nil {
		rtn = true
	}

	return rtn
}

func (z Zsftp) UploadResumeFile(remoteFile, localPath string) bool {
	var err error
	var rtn = false

	ftpClient := ftp4go2.NewFTP(0)
	//connect
	_, err = ftpClient.Connect(z.Ip, z.Port, "")
	if err != nil {
		return rtn
	}

	defer ftpClient.Quit()

	_, err = ftpClient.Login(z.Username, z.Password, "")
	if err != nil {
		return rtn
	}

	err = ftpClient.UploadResumeFile(remoteFile, localPath, false, updateProgress)
	if err == nil {
		rtn = true
	}

	return rtn
}