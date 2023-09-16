package main

import (
	"crypto/rand"
	"fmt"
	"github.com/xuri/excelize/v2"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

func generateRandomIPv6() string {
	ip := make([]byte, 16)

	// Đặt các giá trị cố định để đảm bảo định dạng đúng của IPv6
	ip[0] = 0x20
	ip[1] = 0x01

	// Tạo số ngẫu nhiên cho 12 byte cuối
	_, err := rand.Read(ip[8:])
	if err != nil {
		panic(err)
	}

	return fmt.Sprintf("2001:19f0:7001:321a:%02x:%02x:%02x:%02x", ip[8], ip[9], ip[10], ip[11])
}
func main() {
	for {
		fileExcel, _ := excelize.OpenFile("./data2.xlsx")
		//filePath := "/etc/squid/squid.conf"
		filePath := "./config24.conf"
		file, err := os.Open(filePath)
		if err != nil {
			fmt.Println("Lỗi khi mở file:", err)
			return
		}

		// Đọc nội dung file vào một biến
		content, err := ioutil.ReadFile(filePath)
		if err != nil {
			fmt.Println("Lỗi khi đọc file:", err)
			return
		}
		fileString := string(content)
		ipv6Remove := []string{}
		mutex := &sync.Mutex{}
		wg := &sync.WaitGroup{}
		for i := 1; i <= 300; i++ {
			wg.Add(1)
			go func(i int) {
				ipv6, _ := fileExcel.GetCellValue("Sheet1", fmt.Sprintf("E%d", i))
				ipv6Gen := generateRandomIPv6()
				mutex.Lock()
				ipv6Remove = append(ipv6Remove, ipv6)
				fileString = strings.ReplaceAll(fileString, ipv6, ipv6Gen)
				fileExcel.SetCellValue("Sheet1", fmt.Sprintf("E%d", i), ipv6Gen)
				mutex.Unlock()
				cmd := exec.Command("sudo", "ip", "-6", "address", "add", fmt.Sprintf("%s/64", ipv6Gen), "dev", "enp1s0")
				// Chạy lệnh và xem kết quả
				_, err := cmd.CombinedOutput()
				if err != nil {
					fmt.Println("Lỗi:", err)
					return
				}
				fmt.Printf("Index %d: Add ip %s successful\n", i, ipv6Gen)
				wg.Done()
			}(i)
		}
		wg.Wait()

		fileExcel.Save()
		cmd := exec.Command("sudo", "service", "squid", "restart")

		output, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Println("Lỗi:", err)
			return
		}

		fmt.Println("Restart:", string(output))
		for index, ip := range ipv6Remove {
			wg.Add(1)
			go func(ip string, index int) {
				cmd := exec.Command("sudo", "ip", "-6", "address", "del", fmt.Sprintf("%s/64", ip), "dev", "enp1s0")

				// Chạy lệnh và xem kết quả
				_, err := cmd.CombinedOutput()
				if err != nil {
					fmt.Println("Lỗi:", err)
					return
				}
				fmt.Printf("Index %d: Delete %s successful\n", index, ip)
			}(ip, index)
		}
		file.Close()
		time.Sleep(time.Minute * 5)
		fmt.Println("Done")
	}

}
