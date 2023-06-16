package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"

	"github.com/vishu42/terrasome/pkg/targz"
)

func main() {
	m := http.NewServeMux()

	m.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) {
		// get binary version
		cmd := exec.Command("terraform", "version")
		stdoutStderr, err := cmd.CombinedOutput()
		if err != nil {
			log.Fatal(err)
		}

		w.Write(stdoutStderr)
	})

	m.HandleFunc("/api/v1/upload", func(w http.ResponseWriter, r *http.Request) {
		// only accept POST requests
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		// get the file
		log.Println("getting file from form")
		formFile, _, err := r.FormFile("file")
		if err != nil {
			log.Fatal(err)
		}
		defer formFile.Close()

		// create temp directory to store the file
		td, err := os.MkdirTemp("", "terrasome")
		if err != nil {
			log.Fatal(err)
		}

		// log the temp directory
		log.Println("created temp directory: " + td)

		// cleaup
		defer func() {
			// remove the temp directory
			log.Println("cleaning up...")
			log.Println("removing temp directory: " + td)
			err := os.RemoveAll(td)
			if err != nil {
				log.Fatal(err)
			}
		}()

		// create a file in the temp directory and write the form file to it
		log.Println("creating temp file: " + td + "/plan.tar.gz")
		tarFile, err := os.Create(td + "/plan.tar.gz")
		if err != nil {
			log.Fatal(err)
		}
		defer tarFile.Close()

		log.Println("copying form file to temp file")
		// copy the file to the temp file
		_, err = io.Copy(tarFile, formFile)
		if err != nil {
			log.Fatal(err)
		}

		log.Println("creating temp directory: " + td + "/plan")
		// create a new directory under the temp dir and untar the file
		err = os.Mkdir(td+"/plan", 0o755)
		if err != nil {
			log.Fatal(err)
		}

		log.Println("untaring file to temp directory: " + td + "/plan")
		// untar the file to the new directory under the temp dir called plan
		err = targz.UntarTar(td+"/plan", td+"/plan.tar.gz")
		if err != nil {
			log.Println("error untaring file")
			log.Fatal(err)
		}

		log.Println("running terraform plan")
		// run terraform plan
		cmd := exec.Command("terraform", "plan")
		cmd.Dir = td + "/plan"
		stdoutStderr, err := cmd.CombinedOutput()
		if err != nil {
			log.Fatal(err)
		}

		w.Write(stdoutStderr)
	})

	s := &http.Server{
		Addr:    ":80",
		Handler: m,
	}
	err := s.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
