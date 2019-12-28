// +build ignore

package main

import (
    "fmt"
    "flag"
    "log"
    "io/ioutil"
    "io"
    "bufio"

    "os"
    "os/exec"
    "syscall" // for checking exit status

    "strconv"
    "strings" // for split
    "archive/zip"

    // for type introspection:
    // "reflect"

    // for sleep:
    "time"

    "net/http"
)

const (
    __DATE_VERSION__="2019-Dec-28_20h33m29"

    // Default url used to download scenarii
    __DEFAULT_PUBURL__="https://k8scenario.github.io/static/k8scenarii"

    // time to wait sleeping if running in cluster ????
    INCLUSTER_SLEEP_SECS=3600

    CHECK_FIXED_SLEEP_SECS=10

    escape             = "\x1b"
    colour_me_black    = escape + "[0;30m"
    colour_me_red      = escape + "[0;31m"
    colour_me_green    = escape + "[0;32m"
    colour_me_blue     = escape + "[0;34m"
    colour_me_cyan     = escape + "[0;36m"
    colour_me_yellow   = escape + "[1;33m"
    colour_me_normal   = escape + "[0;0m"
    // p1 := fmt.Sprintf("Served from %s %s%s%s", hostType, colour_me_yellow, hostName, colour_me_normal)
)



var (
    pubUrl = __DEFAULT_PUBURL__

    // Globally accessible so we can easily append it to os/exec environment:
    kubeconfig = ""

    menu      = false
    version   = false
    incluster = false
    verbose   = false
    dbg       = false

    namespace = flag.String("namespace", "k8scenario", "The namespace to use (all)")
    scenario  = flag.Int("scenario", 1, "k8s scenario to run (default: 1)")

    serverUrl = flag.String("server", pubUrl, "Get scenarii from specified server")
    localDir  = flag.String("dir", "", "Get scenarii from local dir")
    zipFile   = flag.String("zip", "", "Get scenarii from local zip file")
    localServer = flag.Bool("localServer", false, "Get scenarii from local server")
)

func CaseInsensitiveContains(s, substr string) bool {
    s, substr = strings.ToUpper(s), strings.ToUpper(substr)
    return strings.Contains(s, substr)
}

// DownloadFile will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory.
func DownloadFile(filepath string, url string) error {
    // Get the data
    head, err := http.Head(url)
    if err != nil {
        fmt.Printf("Failed to download HEADER <%s>\n", url)
        panic(err.Error())
        return err
    }
    defer head.Body.Close()
    if head.StatusCode != 200 {
        fmt.Printf("CODE: %d\n", head.StatusCode)
        os.Exit(1)
    }

    resp, err := http.Get(url)
    if err != nil {
        fmt.Printf("Failed to download <%s>\n", url)
        panic(err.Error())
        return err
    }
    defer resp.Body.Close()

    // Create the file
    out, err := os.Create(filepath)
    if err != nil {
        panic(err.Error())
        return err
    }
    defer out.Close()

    // Write the body to file
    _, err = io.Copy(out, resp.Body)
    if err != nil {
        panic(err.Error())
    }
    return err
}

func myexec(command string) string {
    return _myexec(false, command)
}

func assert_myexec(command string) string {
    return _myexec(true, command)
}

func assert_myexec_pipe(pipeCmd string) string {
    //return _myexec_pipe( true, "/bin/sh", "-c", "'" + pipeCmd + "'" )
    return _myexec_pipe( true, "/bin/sh", "-c", pipeCmd )
}
func myexec_pipe(pipeCmd string) string {
    //return _myexec_pipe( false, "/bin/sh", "-c", "'" + pipeCmd + "'" )
    return _myexec_pipe( false, "/bin/sh", "-c", pipeCmd )
}

func _myexec_pipe(assert bool, pipeCmd ...string) string {
    debug(fmt.Sprintf("---- %s\n", strings.Join(pipeCmd[2:], " ")))
    //fmt.Println("---- " + strings.Join(pipeCmd[2:]))
    //fmt.Printf("---- %s\n", pipeCmd)

    head := pipeCmd[0]
    parts := pipeCmd[1:len(pipeCmd)]
    //out, err := exec.Command(head,parts...).Output()
    //out, err := exec.Command(pipeCmd[0], pipeCmd[1:]).Output()
    //cmd := exec.Command(head,parts...)
    out, err := exec.Command(head,parts...).Output()
    //out, err := exec.Command(pipeCmd[0], pipeCmd[1:]).Output()
    if err != nil {
        msg := fmt.Sprintf("exec.Command(PIPE) returned error with %s\n", err.Error())
        if assert {
            log.Fatalf(msg)
        } else {
            debug(msg)
        }
    }

    /*
    // if kubeconfig != "" { cmd.Env = append(os.Environ(), "KUBECONFIG=" + kubeconfig) }

    out, err := cmd.Output()
    if err != nil {
        msg := fmt.Sprintf("exec.Command(PIPE) returned error with %s\n", err.Error())
        if assert {
            log.Fatalf(msg)
        } else {
            debug(msg)
        }
    }
    */

    return return_out_exit_status(err,  out)
}

func return_out_exit_status(err error,  out []byte) string {
    /*
    if msg, ok := err.(*exec.ExitError); ok { // there is error code
        os.Exit(msg.Sys().(syscall.WaitStatus).ExitStatus())
    } else {
        os.Exit(0)
    }
    */

    exit_status := ""
    exit_status_int := 0
    // From: https://stackoverflow.com/questions/10385551/get-exit-code-go/10385867#10385867
    /*if err := cmd.Wait(); err != nil {}*/
    if err != nil {
        if exiterr, ok := err.(*exec.ExitError); ok {
            // The program has exited with an exit code != 0 (Unix & Windows)
            if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
                exit_status_int = status.ExitStatus()
                exit_status = fmt.Sprintf("exit status: %d", exit_status_int)
		//log.Println(exit_status)
                //exit_status = "\n" + exit_status
            }
        } else {
            log.Fatalf("cmd.Wait: %v", err)
        }
    }

    //if msg, ok := err.(*exec.ExitError); ok { // there is error code
    //}

    EXIT_COLOUR:=colour_me_normal
    if exit_status_int != 0 { EXIT_COLOUR=colour_me_red }
    if exit_status != ""    { exit_status = " - " + exit_status }
    fmt.Printf("%s%s%s%s%s%s\n",
        colour_me_yellow, string(out), colour_me_normal,
        EXIT_COLOUR, exit_status, colour_me_normal)
    return string(out) + exit_status
}

func _myexec(assert bool, command string) string {
    debug("---- " + command + "\n")

    parts := strings.Fields(command)
    head := parts[0]
    parts = parts[1:len(parts)]
    //cmd := exec.Command(head,parts...)
    out, err := exec.Command(head,parts...).Output()
    //out, err := exec.Command(pipeCmd[0], pipeCmd[1:]).Output()
    if err != nil {
        msg := fmt.Sprintf("exec.Command(PIPE) returned error with %s\n", err.Error())
        if assert {
            log.Fatalf(msg)
        } else {
            debug(msg)
        }
    }

    /*
    //if kubeconfig != "" { cmd.Env = append(os.Environ(), "KUBECONFIG=" + kubeconfig) }
    out, err := cmd.Output()
    if err != nil {
        msg := fmt.Sprintf("exec.Command() returned error with %s\n", err.Error())
        if assert {
            log.Fatalf(msg)
        } else {
            debug(msg)
        }
    }*/

    return return_out_exit_status(err,  out)
}

func assertNoErr(err error) {
    if err != nil {
        panic(err)
    }
}

func checkErr(err error) {
    if err != nil {
        fmt.Print(err.Error())
    }
}

func isFile(filename string) bool {
    info, err := os.Stat(filename)
    if os.IsNotExist(err) {
        return false
    }
    return !info.IsDir()
}

func install_scenario(scenario int) (string, string) {
    //debug(fmt.Sprintf("Downloading scenario %d ...\n", scenario))
    fmt.Sprintf("Downloading scenario %d ...\n", scenario)

    scenarioName := fmt.Sprintf("scenario%d", scenario)
    zipUrl      := fmt.Sprintf("%s/%s.zip", *serverUrl, scenarioName)
    zipFile := "/tmp/a.zip"

    //*namespace = scenarioName
    *namespace = "k8scenario"

    if strings.Contains(*serverUrl, "file:") {
        zipFile = zipUrl[ 8: ]
        debug(fmt.Sprintf("Using file %s for %s ... \n", zipFile, scenarioName))
    }

    if strings.Contains(*serverUrl, "http:") ||
       strings.Contains(*serverUrl, "https:") {
        debug(fmt.Sprintf("Downloading %s for %s ... \n", zipUrl, scenarioName))
        if err := DownloadFile(zipFile, zipUrl); err != nil {
            panic(err)
        }
    }

    if ! isFile(zipFile) {
        log.Fatalf("No such file: %s", zipFile)
    }

    return install_scenario_zip(zipFile)
}

func install_scenario_zip(zipFile string) (string, string) {
    fmt.Println("Installing scenario ... into namespace " + *namespace)
    // Open a zip archive for reading.
    r, err := zip.OpenReader(zipFile)
    if err != nil {
        log.Fatal(err)
    }
    defer r.Close()

    // Iterate through the files in the archive,
    // printing some of their contents.
    for _, f := range r.File {
        strlen := len(f.Name)
        if f.Name[strlen-1:] == "/" { continue; }
        debug(fmt.Sprintf("File: %s:\n", f.Name))
    }

    pipeCmd := "kubectl get namespace | awk '/^k8scenario/ { printf \"namespace/%s \", $1; }'"
    namespaces := myexec_pipe(pipeCmd)
    if namespaces != "" {
        fmt.Println( "Deleting existing: " + namespaces)
        fmt.Println( myexec("kubectl delete " + namespaces) )
    }
    //for index, element := range namespaces { }

    fmt.Println( myexec(fmt.Sprintf("kubectl create namespace %s", *namespace)) )
    fmt.Println( myexec_pipe("kubectl get namespace | grep ^k8scenario") )

    INSTRUCTIONS := ""
    CHECK_SCRIPT := ""

    for _, f := range r.File {
        strlen := len(f.Name)
        if f.Name[strlen-1:] == "/" { continue; }
        //fmt.Printf("Contents of %s:\n", f.Name)
        rc, err := f.Open()
        if err != nil {
            fmt.Print("Open error")
            log.Fatal(err)
        }

        tempfile := "/tmp/temp"
        tmpfh, err := os.Create(tempfile)

        //_, err = io.CopyN(os.Stdout, rc, 100)
        _, err = io.CopyN(tmpfh, rc, 2000)
        if err != nil {
            // Dont' print EOF - fmt.Println(err)
            fmt.Print()
            //fmt.Println("CopyN error")
            //log.Fatal(err)
        }
        tmpfh.Close()

        // GATHER CHECK_SCRIPT:
        if strings.Contains(f.Name, "check.sh") {
            CHECK_SCRIPTbytes, errchk := ioutil.ReadFile(tempfile)
            if errchk != nil {
                fmt.Println(errchk)
            }

            CHECK_SCRIPT = string( CHECK_SCRIPTbytes )
	    debug(CHECK_SCRIPT)
        }

        // GATHER INSTRUCTIONS:
        //fmt.Println( myexec("ls -al " + tempfile) )
        if strings.Contains(f.Name, "INSTRUCTIONS.txt") {
            INSTRUCTIONSbytes, errins := ioutil.ReadFile(tempfile)
            if errins != nil {
                fmt.Println(errins)
            }

            INSTRUCTIONS = string( INSTRUCTIONSbytes )
        }

        // APPLY YAML:
        if CaseInsensitiveContains(f.Name, "yaml") ||
           CaseInsensitiveContains(f.Name, "yml")   {
            fmt.Println( myexec_pipe( fmt.Sprintf("kubectl apply -n %s -f %s | grep -v ^$\n", *namespace, tempfile)))
        }

        // TODO: here also - to show scenario status ??

        rc.Close()
        fmt.Println()
    }

    debug("======== " + zipFile + "\n")
    fmt.Println(INSTRUCTIONS)

    return CHECK_SCRIPT, INSTRUCTIONS
}

func menu_loop() {
    listUrl := fmt.Sprintf("%s/index.list", *serverUrl)
    listFile := "/tmp/index.list"

    if verbose {
        debug(fmt.Sprintf("Downloading %s\n", listUrl))
    } else {
        fmt.Printf("Downloading index.list\n")
    }

    if strings.Contains(*serverUrl, "file:") {
        listFile = listUrl[ 8: ]
        debug(fmt.Sprintf("Using file %s for index.list ... \n", listFile))
    }

    if strings.Contains(*serverUrl, "http:") ||
       strings.Contains(*serverUrl, "https:") {
        debug(fmt.Sprintf("Downloading %s for index.list ... \n", listUrl))
        if err := DownloadFile(listFile, listUrl); err != nil {
            panic(err)
        }
    }

    if ! isFile(listFile) {
        log.Fatalf("No such file: %s", listFile)
    }

    fileH, err := os.Open(listFile)
    if err != nil {
        log.Fatalf("failed opening file: %s", err)
    }

    scanner := bufio.NewScanner(fileH)
    scanner.Split(bufio.ScanLines)
    var txtlines []string
    var scenarii []string

    for scanner.Scan() {
        txt := scanner.Text()
        txtlines = append(txtlines, txt)
        scenarii = append(scenarii, txt)
    }
    fileH.Close()

    reader := bufio.NewReader(os.Stdin)

    //for_menu_loop:
    loop_num := 0
    for true {
        loop_num = loop_num + 1
        //fmt.Println()
        fmt.Print("Available scenarii: " )
        for _, eachline := range txtlines {
            fmt.Print(eachline + "  ")
        }

        fmt.Printf("\n%sselect scenario>>>%s ", colour_me_yellow, colour_me_normal)

        line, _ := reader.ReadString('\n')
        //fmt.Println("'" + line + "'")
        line = strings.TrimSpace(line)
        //fmt.Println("'" + line + "'")
        if CaseInsensitiveContains(line, "q") ||
           CaseInsensitiveContains(line, "quit") ||
           CaseInsensitiveContains(line, "exit") {
            os.Exit(0)
        }
        scenario, err := strconv.Atoi(line)
        if err != nil {
            fmt.Printf("Bad number <%s> select from: ", line)
            scenario = -1
            //break for_menu_loop;
        } else {
            check_script, instructions := install_scenario(scenario)

            // Only show instructions once very 10 loops ...
            if loop_num % 10 != 0 {
                instructions = ""
            }
            loop_check(check_script, instructions, scenario)
        }
    }
}

func loop_check(check_script string, instructions string, scenario int) {
    scenarioName := fmt.Sprintf("scenario%d", scenario)

    if check_script == "" {
        fmt.Printf("---- %s[%s]%s NO Cluster check script - choose new scenario when you want to move on\n",
                   colour_me_yellow, scenarioName, colour_me_normal)
        return
    }

    for true {
        //op := myexec(check_script)
        op := myexec_pipe(check_script)
        // fmt.Println(op)
        //op = myexec_pipe(check_script + " 2>&1 | tee /tmp/check_script." + string(scenario) + ".txt")
        if ! strings.Contains(op, "exit status") {
            well_done := fmt.Sprintf("---- %s[%s]%s %sWELL DONE !!!!%s - The scenario appears to be fixed !!!!\n",
	                             colour_me_yellow, scenarioName, colour_me_normal,
	                             colour_me_green, colour_me_normal)
            fmt.Println(well_done)
            return
        }

        //fmt.Println()
        //fmt.Println("---- Cluster appears broken:")
        //fmt.Println(instructions)
	fmt.Printf("%s%s%s", colour_me_yellow, instructions, colour_me_normal)
        sleep(CHECK_FIXED_SLEEP_SECS, 
	      fmt.Sprintf("%s[%s]%s - %scluster broken%s",
	                  colour_me_yellow, scenarioName, colour_me_normal,
	                  colour_me_red, colour_me_normal) )
    }
}

func sleep(secs int, msg string) {
    if msg != "" {
        fmt.Printf("%s - Sleeping <%d> secs ...", msg, secs)
    } else {
        fmt.Printf("Sleeping <%d> secs ... ", secs)
    }

    delay := time.Duration(secs) * 1000 * time.Millisecond
    time.Sleep(delay)
    //fmt.Println("Done")
}

func debug(msg string) {
    if dbg {
        fmt.Printf("%sDEBUG:%s%s", colour_me_cyan, msg, colour_me_normal)
    }
}

func main() {
    flag.BoolVar(&version, "version",  false, "Show version string (false)")
    flag.BoolVar(&menu,    "menu",     false, "Menu to select a scenario")
    flag.BoolVar(&verbose, "verbose ", false, "Verbose mode")
    flag.BoolVar(&dbg,     "debug",    false, "Debug mode")

    flag.Parse()

    fmt.Println("Version: " + __DATE_VERSION__)
    if version {
        os.Exit(0)
    }
    fmt.Println("serverUrl: " + *serverUrl)

    if len(os.Args) == 1 {
         menu=true
    }

    /*if len(os.Args) > 1 && os.Args[1] == "--version" {
        fmt.Println("gotcha !!")
        os.Exit(0)
    }*/

    if *zipFile != "" {
        _, _ = install_scenario_zip(*zipFile)
        os.Exit(0)
    }

    if *localDir != "" {
        *serverUrl = "file:///" + *localDir
    }

    if *localServer {
        *serverUrl = "http://127.0.0.1:9000"
    }

    kubeconfig = os.Getenv("KUBECONFIG")
    debug("INFO env[kubeconfig]=" + kubeconfig + "\n")
    debug("INFO env[HOME]=" + os.Getenv("HOME") + "\n")

    if kubeconfig != "" {
        if isFile(kubeconfig) {
            debug("[exists] kubeconfig=" + kubeconfig + "\n")
        } else {
            debug("[no such file] kubeconfig=" + kubeconfig + "\n")
            kubeconfig = ""
        }
    }

    if kubeconfig == "" {
        kubeconfigHOME := os.Getenv("HOME")+"/.kube/config"
        if isFile(kubeconfigHOME) {
            debug("[exists] kubeconfigHOME=" + kubeconfigHOME + "\n")
            kubeconfig = kubeconfigHOME
        } else {
            debug("[no such file] kubeconfigHOME=" + kubeconfigHOME + "\n")
        }
    }

    if menu {
        menu_loop()
    }

    _, _ = install_scenario(*scenario)

    /* type introspection:
    fmt.Println(reflect.TypeOf(rc))
    typerc := reflect.TypeOf(rc)
    for i:=0; i<typerc.NumMethod(); i++ {
        fmt.Print(typerc.Method(i))
    }*/

    if incluster {
        sleep(INCLUSTER_SLEEP_SECS, "in cluster")
    }

    os.Exit(0)
}



