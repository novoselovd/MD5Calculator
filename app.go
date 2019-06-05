package main

import (
    "crypto/md5"
    "encoding/hex"
    "encoding/json"
    "fmt"
    "github.com/kardianos/osext"
    "io"
    "log"
    "net/http"
    "net/url"
    "os"
    "strconv"
)

// Path for files download(optional)
var WORKING_DIR = ""

type hashedFile struct {
    id     string `json:"-"`
    Md5    string `json:"md5"`
    Status string `json:"status"`
    Url    string `json:"url"`
}

type tempStruct struct {
    Status  string `json:"status"`
}

var db []hashedFile

// Checking if str is url
func isUrl(str string) bool {
    u, err := url.Parse(str)
    return err == nil && u.Scheme != "" && u.Host != ""
}

// Getting md5 hash of file
func getHash(filePath string) (string, error) {
    var returnMD5String string

    // Open the passed argument and check for any error
    file, err := os.Open(filePath)
    if err != nil {
        return returnMD5String, err
    }

    defer file.Close()

    // Open a new hash interface to write to
    hash := md5.New()

    // Copy the file in the hash interface and check for any error
    if _, err := io.Copy(hash, file); err != nil {
        return returnMD5String, err
    }

    // Get the 16 bytes hash
    hashInBytes := hash.Sum(nil)[:16]

    // Convert the bytes to a string
    returnMD5String = hex.EncodeToString(hashInBytes)

    return returnMD5String, nil

}

// Downloading file
func downloadFile(filepath string, url string) (err error) {
    out, err := os.Create(filepath)
    if err != nil  {
        return err
    }
    defer out.Close()

    // Get the data
    resp, err := http.Get(url)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    // Check server response
    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("bad status: %s", resp.Status)
    }

    // Writer the body to file
    _, err = io.Copy(out, resp.Body)
    if err != nil  {
        return err
    }

    return nil
}

func checkRouterHandler(w http.ResponseWriter, r *http.Request) {
    dataFound := false
    switch r.Method {
    case "GET":
        w.Header().Set("Content-Type", "application/json")
        keys, ok := r.URL.Query()["id"]

        // Getting passed id
        if !ok || len(keys[0]) < 1 {
            log.Println("Url Param 'id' is missing")
            return
        }
        id := string(keys[0])

        for _, item := range db {
            if item.id == id {
                dataFound = true
                if item.Status == "running" {                 // In process case
                    runningTask := tempStruct{
                        Status: "running",
                    }

                    w.WriteHeader(http.StatusOK)
                    json.NewEncoder(w).Encode(runningTask)

                    break
                } else if item.Status == "error occurred" {                // Error case
                    errorTask := tempStruct{
                        Status: "error occurred",
                    }

                    w.WriteHeader(http.StatusOK)
                    json.NewEncoder(w).Encode(errorTask)
                }

                w.WriteHeader(http.StatusOK)
                json.NewEncoder(w).Encode(item)

                break
            }
        }

        if !dataFound {                                  // If task doesn't exist
            notFoundTask := tempStruct{
                Status: "identifier doesn't exist",
            }

            w.WriteHeader(http.StatusNotFound)
            json.NewEncoder(w).Encode(notFoundTask)
        }
    default:
        fmt.Fprint(w, "usage: curl -X GET http://127.0.0.1:5000/check?id=c4ca4238a0b923820dcc509a6f75849b")
    }
}

func submitRouterHandler(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case "POST":
        if err := r.ParseForm(); err != nil {
            fmt.Fprintf(w, "ParseForm() err: %v", err)
            return
        }
        uri := r.FormValue("url")
        if isUrl(uri) {
            temp := []byte(strconv.Itoa(len(db)+1))
            id := fmt.Sprintf("%x", md5.Sum(temp))
            fmt.Fprintf(w,"{\"id\":\"%s\"}\n", id)

            newHashedFile := hashedFile{
                id:     id,
                Md5:    "",
                Status: "running",
                Url:    uri,
            }
            db = append(db, newHashedFile)

            // Downloading a file and getting it's hash in background
            go func() {
                if err := downloadFile(WORKING_DIR + id + ".md5", uri); err == nil {
                   if hash, errHash := getHash(WORKING_DIR + id + ".md5"); errHash == nil {
                       for i, item := range db {
                           if item.id == id {
                               (&db[i]).Status = "done"
                               (&db[i]).Md5 = hash
                               break
                           }
                       }
                   }
                } else {
                    for i, item := range db {
                        if item.id == id {
                            (&db[i]).Status = "error occurred"
                            break
                        }
                    }
                }
            }()
        }
    default:
        fmt.Fprint(w, "usage: curl -X POST -d \"url=http://site.com/file.txt\" http://127.0.0.1:5000/submit")
    }
}

func main() {
    if WORKING_DIR == "" {
        WORKING_DIR, _ = osext.ExecutableFolder()
    }
    http.HandleFunc("/check", checkRouterHandler)
    http.HandleFunc("/submit", submitRouterHandler)
    err := http.ListenAndServe(":5000", nil)
    if err != nil {
        log.Fatal("ListenAndServe: ", err)
    }
}

