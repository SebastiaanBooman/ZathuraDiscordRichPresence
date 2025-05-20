package main

import (
	"github.com/godbus/dbus/v5"

	"ZathuraDiscordRichPresence/discord_rpc"
	"ZathuraDiscordRichPresence/logging"

	"encoding/json"
	"errors"
	"flag"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type DocumentInfo struct {
	Filename      string   `json:"filename"`
	NumberOfPages int      `json:"numberofpages"`
	Index         *[]Index `json:"index"`
}

type Index struct {
	Title    string   `json:"title"`
	Page     int      `json:"page"`
	Subindex *[]Index `json:"subindex"`
}

func getZathuraProcessId() (isZathuraProcessRunning bool, firstZathuraProcessId string, error error) {
	const processNameToSearch = "zathura"

	cmd := exec.Command("pgrep", "--list-name", processNameToSearch)

	output, err := cmd.Output()
	if err != nil {
		logging.ErrorWithContext("Can not get Zathura process id due to error: ", err)
		return false, "", err
	}

	zathura_process_id_names := strings.Split(strings.TrimSpace(string(output)), "\n")

	if len(zathura_process_id_names) == 0 {
		logging.ErrorWithContext("Can not get Zathura process id due to error: ", err)
		return false, "", nil
	}

	if len(zathura_process_id_names) > 1 {
		logging.WarnWithContext("Multiple processes found containing \""+processNameToSearch+"\" in their name. Attempting to use the first one which matches \"zathura\" exactly", zathura_process_id_names)
	}

	for _, process_id_name := range zathura_process_id_names {
		split_process_id_name := strings.Split(string(process_id_name), " ")
		if len(split_process_id_name) != 2 {
			logging.ErrorWithContext("Length of split_process_id_name was not 2. Actual length:", len(split_process_id_name))
			continue
		}

		if split_process_id_name[1] == "zathura" {
			return true, split_process_id_name[0], nil
		}
	}

	logging.LogInfo("No processes found containing exact name: \"" + processNameToSearch + "\"")

	return false, "", nil
}

func getZathuraDocumentInfo(
	dbusConnection *dbus.Conn,
	zathuraProcessId string) (
	fileName string, pageNumber int, numberOfPages int, documentInfo string, error error) {
	process_id := "org.pwmt.zathura.PID-" + zathuraProcessId

	objectPath := dbus.ObjectPath("/org/pwmt/zathura")

	const ZATHURA_INTERFACE_NAME = "org.pwmt.zathura"
	const DBUS_PROPERTIES_GET_METHOD_NAME = "org.freedesktop.DBus.Properties.Get"

	const PROPERTY_NAME_FILE_NAME = "filename"
	const PROPERTY_NAME_PAGE_NUMBER = "pagenumber"
	const PROPERTY_NAME_NUMBER_OF_PAGES = "numberofpages"
	const PROPERTY_NAME_DOCUMENT_INFO = "documentinfo"

	obj := dbusConnection.Object(process_id, objectPath)
	var fileNamePropertyValue string
	var pageNumberPropertyValue int
	var numberOfPagesPropertyValue int
	var documentInfoPagesPropertyValue string

	err := obj.Call(DBUS_PROPERTIES_GET_METHOD_NAME, 0, ZATHURA_INTERFACE_NAME, PROPERTY_NAME_FILE_NAME).Store(&fileNamePropertyValue)

	//NOTE: Not logging if an error occurs here as it could mean that no file is opened (which would spam the log file. Not sure if there is a better way to check for this condition...
	if err != nil {
		return "", 0, 0, "", err
	}

	err = obj.Call(DBUS_PROPERTIES_GET_METHOD_NAME, 0, ZATHURA_INTERFACE_NAME, PROPERTY_NAME_PAGE_NUMBER).Store(&pageNumberPropertyValue)

	if err != nil {
		return "", 0, 0, "", err
	}

	err = obj.Call(DBUS_PROPERTIES_GET_METHOD_NAME, 0, ZATHURA_INTERFACE_NAME, PROPERTY_NAME_NUMBER_OF_PAGES).Store(&numberOfPagesPropertyValue)

	if err != nil {
		return "", 0, 0, "", err
	}

	err = obj.Call(DBUS_PROPERTIES_GET_METHOD_NAME, 0, ZATHURA_INTERFACE_NAME, PROPERTY_NAME_DOCUMENT_INFO).Store(&documentInfoPagesPropertyValue)

	if err != nil {
		return "", 0, 0, "", err
	}

	//NOTE: pagenumber is 0-based
	pageNumberPropertyValue += 1

	return fileNamePropertyValue, pageNumberPropertyValue, numberOfPagesPropertyValue, documentInfoPagesPropertyValue, nil
}

func deserializeDocumentInfo(documentInfoString string) (DocumentInfo, error) {
	var documentInfo DocumentInfo
	err := json.Unmarshal([]byte(documentInfoString), &documentInfo)

	if err != nil {
		logging.ErrorWithContext("Could not deserialize JSON: ", err)
		return documentInfo, err
	}

	return documentInfo, nil
}

func getChapterNameBasedOnPageNumber(documentInfo DocumentInfo, pageNumber int) (chapterName string, error error) {
	documentIndex := *documentInfo.Index

	if documentIndex == nil {
		logging.Warn("No index found on documentInfo with Filename: " + documentInfo.Filename)
		return "", errors.New("No index found on documentInfo with Filename: " + documentInfo.Filename)
	}

	indexName := ""
	latestFoundIndex := -1

	//NOTE: Unsure if the index array is always sorted (would make sense if it was)
	for _, indexEntry := range documentIndex {
		if indexEntry.Page > latestFoundIndex && indexEntry.Page <= pageNumber {
			indexName = indexEntry.Title
			latestFoundIndex = indexEntry.Page
		}
	}

	if indexName == "" {
		logging.Warn("No index found on documentInfo with Filename (indexName == \"\"): " + documentInfo.Filename)
		return "", errors.New("No index found on documentInfo with Filename (indexName == \"\"): " + documentInfo.Filename)
	}

	return indexName, nil
}

// Determines which icon to use based on file extension
// TODO: Add more icons (?)
func getDocumentIcon(fileNameWithExtension string) string {
	extensionSplit := strings.Split(fileNameWithExtension, ".")
	extension := extensionSplit[len(extensionSplit)-1]

	switch {
	case strings.EqualFold(extension, "pdf"):
		return "pdf-icon"
	case strings.EqualFold(extension, "epub"):
		return "epub-icon"
	default:
		return "blank-icon"
	}
}

func main() {
	shouldShowChapters := flag.Bool("show-chapters", false, "Whether to show document chapter information in the rich presence")
	flag.Parse()

	logFile, err := logging.SetupLogger("app.log")
	if err != nil {
		panic("Failed to setup log file: " + err.Error())
	}
	defer logFile.Close()

	var isDiscordRpcConnected bool = false
	var timeStartedReading time.Time
	var lastOpenedFileName string
	var dbusConnection *dbus.Conn

	logging.LogInfo("Application started! Starting main loop... ")

	for {
		if dbusConnection == nil || !dbusConnection.Connected() {
			dbusConnection, err = dbus.ConnectSessionBus()

			if err != nil {
				logging.Error("Failed to connect to session bus: " + err.Error())
				time.Sleep(5 * time.Second)
				continue
			}
		}

		isZathuraProcessRunning, zathuraProcessId, err := getZathuraProcessId()
		if err != nil || !isZathuraProcessRunning {
			if isDiscordRpcConnected {
				discordrpc.Logout()
				isDiscordRpcConnected = false
			}
			continue
		}

		if !isDiscordRpcConnected {
			err := discordrpc.Login()
			if err != nil {
				logging.ErrorWithContext("Error during discordrpc.Login", err)
				continue
			}
			isDiscordRpcConnected = true
		}

		var isDocumentOpen bool = false

		fileName, pageNumber, numberOfPages, documentInfo, err := getZathuraDocumentInfo(
			dbusConnection, zathuraProcessId)

		if err == nil {
			//TODO: Add return value for tracking whether document is open (?)
			isDocumentOpen = true
		}

		var largeText string = "Zathura - a document viewer"

		if isDocumentOpen {
			fileNameWithoutPath := filepath.Base(fileName)

			if fileName != lastOpenedFileName {
				timeStartedReading = time.Now()
				lastOpenedFileName = fileName
			}

			documentIcon := getDocumentIcon(fileNameWithoutPath)

			if *shouldShowChapters {

				deserializedDocumentInfo, err := deserializeDocumentInfo(documentInfo)

				if err != nil {
					continue
				}
				chapterName, err := getChapterNameBasedOnPageNumber(deserializedDocumentInfo, pageNumber)

				if err == nil {
					largeText = "Chapter: " + chapterName
				}
			}

			err = discordrpc.SetActivity("Page: "+strconv.Itoa(pageNumber)+"/"+strconv.Itoa(numberOfPages), fileNameWithoutPath, documentIcon, largeText, timeStartedReading)

			if err != nil {
				logging.ErrorWithContext("Error during discordrpc.SetActiviy", err)
			}
		} else {
			if timeStartedReading.IsZero() {
				timeStartedReading = time.Now()
			}
			lastOpenedFileName = ""
			err = discordrpc.SetActivity("No file opened", "", "blank-icon", largeText, timeStartedReading)
		}

		time.Sleep(1 * time.Second)
	}
}
