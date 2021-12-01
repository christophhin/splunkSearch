package main

import (
  "fmt"
  "os"
  "io"
  "path/filepath"
  "gopkg.in/ini.v1"
  "net/http"
  "crypto/tls"
  "strings"
  "net/url"
)

type INI struct {
  host     string
  port     int
  user     string
  pswd     string
}

func readINI() INI {
  // --- find ini file ---
  file, _ := os.Readlink("/proc/self/exe")

  // --- read ini file ---
  cfg, err := ini.Load(filepath.Join(filepath.Dir(file), "splunkSearch.ini"))
  if err != nil {
    panic(err.Error())
  }

  port, err := cfg.Section("splunk").Key("port").Int() 
  ini := INI {
                cfg.Section("splunk").Key("host").String(),
                port,
                cfg.Section("splunk").Key("user").String(),
                cfg.Section("splunk").Key("pswd").String(),
             }
  return ini
}

func main() {
  // --- check command line arguments ---
  if len(os.Args) != 4 {
    fmt.Println("Usage:")
    fmt.Println("   splunkSearch search_string earliest latest");
    os.Exit(-1)	
  }
  search   := url.QueryEscape(os.Args[1])
  earliest := url.QueryEscape(os.Args[2])
  latest   := url.QueryEscape(os.Args[3])

  ini := readINI()

  tr := &http.Transport {
	  TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
  }
  client := &http.Client{Transport: tr}
  
  bodyStr := fmt.Sprintf("search=search %s&earliest_time=%s&latest_time=%s&output_mode=csv", search, earliest, latest)
  body := strings.NewReader(bodyStr)
  url  := "https://" + ini.host + ":" + fmt.Sprintf("%d", ini.port) + "/servicesNS/admin/search/search/jobs/export"

  req, err := http.NewRequest("POST", url, body)
  if err != nil {
    panic(err.Error())
  }
  req.SetBasicAuth(ini.user, ini.pswd)
  req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

  resp, err := client.Do(req)
  if err != nil {
    panic(err.Error())
  }
  defer resp.Body.Close()  


  _, err = io.Copy(os.Stdout, resp.Body)
  if err != nil {
    panic(err.Error())
  }
}
