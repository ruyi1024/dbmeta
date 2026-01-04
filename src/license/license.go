package license

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"dbmcloud/src/database"
	"dbmcloud/src/model"
	"dbmcloud/src/utils"
)

func getMacAddr() string {
	interfaces, err := net.Interfaces()
	if err != nil {
		log.Panic("Get loacl Mac failed")
	}
	for _, inter := range interfaces {
		mac := inter.HardwareAddr
		if mac.String() != "" {
			return mac.String()
		}
	}
	return ""
}

func encryptFile(filename string, key []byte) error {
	inFile, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	fileData := string(inFile)

	ciphertext, err := utils.AesEncrypt([]byte(fileData), key)
	if err != nil {
		return err
	}
	outFile, err := os.Create(filename + ".ENC")
	if err != nil {
		return err
	}
	defer outFile.Close()

	if _, err = outFile.Write(ciphertext); err != nil {
		return err
	}

	return nil
}

func decryptFile(filename string, key []byte) ([]byte, error) {
	inFile, err := os.ReadFile(filename)
	if err != nil {
		return []byte(""), err
	}

	fileData := string(inFile)

	plaintext, err := utils.AesDecrypt([]byte(fileData), key)
	if err != nil {
		return []byte(""), err
	}

	return plaintext, nil
}

type License struct {
	ExpireDate    time.Time `json:"expire_date"`    // 过期日期
	MachineID     string    `json:"machine_id"`     // 机器码
	MaxDatasource int       `json:"max_datasource"` // 数据源
}

func Check() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered from panic:", r)
			os.Exit(0)

		}
	}()
	// 读取本地license文件
	/*
		licenseContent, err := os.ReadFile("LICENSE")
		if err != nil {
			log.Panic("Failed to read license file:", err)
			return
		}
	*/
	key := []byte("Ht9@Km5#LwQ2&Np0")
	licenseContent, err := decryptFile("LICENSE.ENC", key)
	if err != nil {
		log.Panic("Failed to read license file:", err)
		return
	}

	var license License
	err = json.Unmarshal(licenseContent, &license)
	if err != nil {
		log.Panic("Failed to parse license file:", err)
		return
	}
	// fmt.Println(license.MachineID)
	// fmt.Println(license.ExpireDate)
	// fmt.Println(license.MaxDatasource)

	// 验证机器码
	machineIdList := strings.Split(license.MachineID, ";")
	currentMachineID := getMacAddr()
	validMachine := false
	for _, machineID := range machineIdList {
		if machineID == currentMachineID {
			validMachine = true
		}
	}
	if !validMachine {
		log.Panic("License check failid: Invalid machine ID.")
		return
	}

	// 验证过期日期
	if time.Now().After(license.ExpireDate) {
		log.Panic("License check failid: license has expired.")
		return
	}

	var count int64
	database.DB.Model(&model.Datasource{}).Count(&count)
	// 验证数据源数量
	if count > int64(license.MaxDatasource) {
		log.Panic("License check failid: datasource limited.")
		return
	}

	fmt.Println("License is valid.")
}

func Display() {
	currentMachineID := getMacAddr()
	fmt.Println("Machine Id: ", currentMachineID)
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered from panic:", r)
			os.Exit(0)

		}
	}()
	// 读取本地license文件
	/*
		licenseContent, err := os.ReadFile("LICENSE")
		if err != nil {
			log.Panic("Failed to read license file:", err)
			return
		}
	*/
	key := []byte("Ht9@Km5#LwQ2&Np0")
	licenseContent, err := decryptFile("LICENSE.ENC", key)
	if err != nil {
		log.Panic("Failed to read license file:", err)
		return
	}

	var license License
	err = json.Unmarshal(licenseContent, &license)
	if err != nil {
		log.Panic("Failed to parse license file:", err)
		return
	}
	fmt.Printf("LicenseI Info: machine id:%s, expire date:%s, max datasource: %d", license.MachineID, license.ExpireDate, license.MaxDatasource)

}
