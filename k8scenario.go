// +build ignore

package main

import (
    "fmt"
    "flag"
    "log"
    "io/ioutil"
    "io"
    "bufio"   // for reading stdin

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
    __DATE_VERSION__="2020-Sep-05_00h15m52"
    __K8SCENARIO_VERSION__="k8scenario.public"

    // Default url used to download scenarii
    __DEFAULT_PUBURL__="https://k8scenario.github.io/static/k8scenarii"

    // time to wait sleeping if running in cluster
    INCLUSTER_SLEEP_SECS=3600

    CHECK_FIXED_SLEEP_SECS=5

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

    USER= os.Getenv("USER")
    tempfile = fmt.Sprintf("/tmp/temp.%s", USER)

    // Globally accessible so we can easily append it to os/exec environment:
    kubeconfig = ""

    menu      = false
    version   = false
    incluster = false
    verbose   = false
    dbg       = false
    // If set to true:
    //   Once scenario is solved, don't delete namespace until enter has been pressed:
    wait_label= false

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

func silent_exec(command string) (string,int) {
    showOP := false
    assert := false
    return _exec(showOP, assert, command)
}

func myexec(command string) (string,int) {
    showOP := true
    assert := false
    return _exec(showOP, assert, command)
}

func assert_exec(command string) (string,int) {
    showOP := true
    assert := true
    return _exec(showOP, assert, command)
}

func silent_exec_pipe(pipeCmd string) (string,int) {
    showOP := false
    assert := false
    return _exec_pipe( showOP, assert, "/bin/sh", "-c", pipeCmd )
}
func assert_exec_pipe(pipeCmd string) (string,int) {
    showOP := true
    assert := true
    return _exec_pipe( showOP, assert, "/bin/sh", "-c", pipeCmd )
}
func exec_pipe(pipeCmd string) (string,int) {
    showOP := true
    assert := false
    return _exec_pipe( showOP, assert, "/bin/sh", "-c", pipeCmd )
}

func _exec_pipe(showOP bool, assert bool, pipeCmd ...string) (string, int) {
    debug(fmt.Sprintf("---- %s\n", strings.Join(pipeCmd[2:], " ")))

    head := pipeCmd[0]
    parts := pipeCmd[1:len(pipeCmd)]
    cmd := exec.Command(head,parts...)

    output, err := cmd.Output()

    if err != nil {
        msg := fmt.Sprintf("exec.Command(PIPE) returned error with %s\n", err.Error())
        if assert {
            log.Fatalf(msg)
        } else {
            debug(msg)
        }
    }

    return return_out_exit_status(showOP, err,  output)
}

func return_out_exit_status(showOP bool, err error,  output []byte) (string, int) {

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

    if showOP {
        //fmt.Printf("\n%s%s%s%s%s%s",

        fmt.Printf("%s%s%s", EXIT_COLOUR, exit_status, colour_me_normal)

    }
    return string(output) + exit_status, exit_status_int
}

func _exec(showOP bool, assert bool, command string) (string, int) {
    debug("---- " + command + "\n")

    parts := strings.Fields(command)
    head := parts[0]
    parts = parts[1:len(parts)]
    //cmd := exec.Command(head,parts...)
    output, err := exec.Command(head,parts...).Output()
    //output, err := exec.Command(pipeCmd[0], pipeCmd[1:]).Output()
    if err != nil {
        msg := fmt.Sprintf("exec.Command(PIPE) returned error with %s\n", err.Error())
        if assert {
            log.Fatalf(msg)
        } else {
            debug(msg)
        }
    }

    return return_out_exit_status(showOP, err,  output)
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

func install_scenario(scenario int) (string, string, string) {
    //debug(fmt.Sprintf("Downloading scenario %d ...\n", scenario))
    fmt.Sprintf("Downloading scenario %d ...\n", scenario)

    scenario_name := fmt.Sprintf("scenario%d", scenario)
    zipUrl        := fmt.Sprintf("%s/%s.zip", *serverUrl, scenario_name)
    zipFile       := "/tmp/a.zip"

    os.Setenv("SCENARIO", string(scenario))

    if strings.Contains(*serverUrl, "file:") {
        zipFile = zipUrl[ 8: ]
        debug(fmt.Sprintf("Using file %s for %s ... \n", zipFile, scenario_name))
    }

    if strings.Contains(*serverUrl, "http:") ||
       strings.Contains(*serverUrl, "https:") {
        debug(fmt.Sprintf("Downloading %s for %s ... \n", zipUrl, scenario_name))
        if err := DownloadFile(zipFile, zipUrl); err != nil {
            panic(err)
        }
    }

    if ! isFile(zipFile) {
        log.Fatalf("No such file: %s", zipFile)
    }

    return install_scenario_zip(zipFile, scenario)
}

func readFileFromReader(f *zip.File) (string) {
    //fmt.Printf("Contents of %s:\n", f.Name)
    rc, err := f.Open()
    if err != nil {
        fmt.Println("Open error")
        log.Fatal(err)
    }

    tmpfh, err := os.Create(tempfile)
    //_, _ = io.CopyN(tmpfh, rc, 2000)
    _, _ = io.CopyN(tmpfh, rc, 10000)
    //if err != nil {
        // Dont' print EOF - fmt.Println(err)
        //fmt.Println("CopyN error")
        //log.Fatal(err)
    //}
    tmpfh.Close()

    bytes, readf_err := ioutil.ReadFile(tempfile)

    if readf_err != nil {
        fmt.Println(f.Name + ": " + readf_err.Error())
    }

    str := string(bytes)
    debug(str)

    rc.Close()

    //return str, readf_err
    return str
}

func write_check_script(check_script string, extra string) (string) {

    //CHECK_SCRIPTbytes, readf_err := ioutil.ReadFile(tempfile)
    //if readf_err != nil {
        //fmt.Println("check_script: " + readf_err.Error())
    //}

    CHECK_SCRIPT := extra + string( check_script )
    debug(CHECK_SCRIPT)
    /*
    fmt.Printf("Len(extra)=%d\n", len(extra))
    fmt.Printf("Len(check_script)=%d\n", len(check_script))
    fmt.Println("======")
    fmt.Printf("Len(CHECK_SCRIPT)=%d\n", len(CHECK_SCRIPT))
    */

    // TODO: invoke SETUP_SCRIPT once with '-pre' argument - HERE
    return CHECK_SCRIPT
}

func apply_setup_script(scenario int, setup_script string, extra string, prepost_yaml string) {
    if setup_script == "" {
        return
    }

    SETUP_SCRIPT := extra + setup_script


    // TODO: invoke script once with '-pre' argument - HERE
    if prepost_yaml == "--pre-yaml" {
        set_args := "\nset -- --pre-yaml\n\n"
        _, _ = exec_pipe(set_args + SETUP_SCRIPT)
    } else if prepost_yaml == "--post-yaml" {
        set_args := "\nset -- --post-yaml\n\n"
	_, _ = exec_pipe(set_args + SETUP_SCRIPT)
    } else {
        _, _ = exec_pipe(SETUP_SCRIPT)
    }
}

func writeFile(filename string, content string) {
    f, err := os.Create(tempfile)
    if err != nil {
        fmt.Println(tempfile + ": " + err.Error())
    }
    f.WriteString(content)
    //defer f.Close()
    f.Sync()
    f.Close()
}

func install_scenario_zip(zipFile string, scenario int) (string, string, string) {
    scenario_name := fmt.Sprintf("scenario%d", scenario)
    fmt.Print()
    showVersion()
    fmt.Printf("\n---- %s[%s]%s Installing into namespace %s%s%s\n",
               colour_me_yellow, scenario_name, colour_me_normal,
               colour_me_green, *namespace, colour_me_normal)
    // Open a zip archive for reading.
    r, err := zip.OpenReader(zipFile)
    if err != nil {
        log.Fatal(err)
    }
    defer r.Close()

    // Iterate through the files in the archive,
    // printing some of their contents.
    //for _, f := range r.File {
        //strlen := len(f.Name)
        //if f.Name[strlen-1:] == "/" { continue; }
        //debug(fmt.Sprintf("File: %s:\n", f.Name))
    //}

    pipeCmd := "kubectl get namespace | awk '/^k8scenario/ { printf \"namespace/%s \", $1; }'"
    namespaces, _ := exec_pipe(pipeCmd)
    if namespaces != "" {
        fmt.Println( "Deleting existing: " + namespaces)
        silent_exec( fmt.Sprintf("kubectl label --overwrite namespace %s status=deleting scenario=%d loop-", *namespace, scenario) )
        silent_exec("kubectl delete " + namespaces)
    }

    fmt.Println( "(Re)creating namespace: " + *namespace)
    silent_exec("kubectl create namespace " + *namespace)
    exec_pipe( fmt.Sprintf("kubectl get namespace %s --no-headers", *namespace) )
    INSTRUCTIONS := ""
    CHALLENGE_TYPE := "task"
    SETUP_SCRIPT := ""
    CHECK_SCRIPT := ""
    FUNCTIONS_RC := ""
    EXPORT_NAMESPACE := "export NS=" + *namespace + "\n"
    EXPORT_NOMARK := "export NO_MARK_SCENARIO=DONT_MARK_NAMESPACE\n"

    // empty array of strings:
    YAML_FILES := []string{}
    YAML_FILE_NAMES := []string{}

    file_loop := 0
    for _, f := range r.File {
        file_loop = file_loop  + 1
        strlen := len(f.Name)

        if f.Name[strlen-1:] == "/" { continue; }

        // GATHER FUNCTIONS_RC:
        if strings.Contains(f.Name, ".functions.rc") { FUNCTIONS_RC = readFileFromReader(f) }

        // GATHER SETUP_SCRIPT:
        if strings.Contains(f.Name, "SETUP_SCENARIO.sh") { SETUP_SCRIPT = readFileFromReader(f) }

        // GATHER CHECK_SCRIPT:
        if strings.Contains(f.Name, "CHECK_SCENARIO.sh") { CHECK_SCRIPT = readFileFromReader(f) }

        // GATHER INSTRUCTIONS:
        if strings.Contains(f.Name, "INSTRUCTIONS.txt") { INSTRUCTIONS = readFileFromReader(f) }

        // GATHER CHALLENGE_TYPE:
        if strings.Contains(f.Name, "CHALLENGE_TYPE.txt") { CHALLENGE_TYPE = readFileFromReader(f) }

        if CaseInsensitiveContains(f.Name, "yaml") || CaseInsensitiveContains(f.Name, "yml")   {
            yaml_content   := readFileFromReader(f)
            YAML_FILE_NAMES = append(YAML_FILE_NAMES, f.Name)
            YAML_FILES      = append(YAML_FILES, yaml_content)
        }

        silent_exec( "rm " + tempfile )
    }

    CHALLENGE_TYPE = strings.TrimSuffix( CHALLENGE_TYPE, "\n")
    apply_setup_script(scenario, SETUP_SCRIPT, EXPORT_NAMESPACE + FUNCTIONS_RC + "\n", "--pre-yaml")
    CHECK_SCRIPT = write_check_script(CHECK_SCRIPT, EXPORT_NAMESPACE + EXPORT_NOMARK + FUNCTIONS_RC + "\n")

    for file_idx := 0 ; file_idx < len(YAML_FILES); file_idx++ {
        filename := YAML_FILE_NAMES[file_idx]
        content  := YAML_FILES[file_idx]

        writeFile(tempfile, content)

        applyCmd := fmt.Sprintf("kubectl apply -n %s -f %s", *namespace, tempfile)

        //fmt.Println( applyCmd )
        path_elems := strings.Split(filename, "/")
        filename   = path_elems[len(path_elems)-1]


        full_cmd := fmt.Sprintf("%s | grep -v ^$\n", applyCmd)
        //fmt.Printf("Command '%s'\n", full_cmd)
        exec_pipe( full_cmd )
    }
    apply_setup_script(scenario, SETUP_SCRIPT, EXPORT_NAMESPACE + FUNCTIONS_RC + "\n", "--post-yaml")

    debug("======== " + zipFile + "\n")
    silent_exec( fmt.Sprintf("kubectl label namespace %s status=readytofix scenario=%d", *namespace, scenario) )

    return CHECK_SCRIPT, INSTRUCTIONS, CHALLENGE_TYPE
}

func menu_loop() {
    listUrl := fmt.Sprintf("%s/index.list", *serverUrl)
    listFile := "/tmp/index.list"


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

    for true {
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
            check_script, instructions, challenge_type := install_scenario(scenario)

            loop_check(check_script, instructions, challenge_type, scenario)
        }
    }
}

func loop_check(check_script string, instructions string, challenge_type string, scenario int) {
    scenario_name := fmt.Sprintf("scenario%d", scenario)

    if check_script == "" {
        fmt.Printf("---- %s[%s]%s NO Cluster check script - choose new scenario when you want to move on\n",
                   colour_me_yellow, scenario_name, colour_me_normal)
        return
    }

    prompt := ""
    switch challenge_type {
        case "task": prompt = "task incomplete"
        case "fix": prompt = "fix incomplete"
        default: prompt = "**** incomplete"
    }

    //for_check_loop:
    loop_num := 0
    for true {
        loop_num = loop_num + 1

        // Only show instructions once very 10 loops ...
        if (loop_num % 10) == 1 {
            fmt.Printf("\n%s", instructions)
        }

        sleep(CHECK_FIXED_SLEEP_SECS, 
	      fmt.Sprintf("\n%s[%s]/%d%s - %s%s%s",
	          colour_me_yellow, scenario_name, loop_num, colour_me_normal,
	          colour_me_red, prompt, colour_me_normal) )

        _, err_code := exec_pipe(check_script)
        if err_code == 0 {
            well_done := fmt.Sprintf("\n---- %s[%s]%s %sWELL DONE !!!!%s - The scenario appears to be fixed !!!!\n",
                             colour_me_yellow, scenario_name, colour_me_normal,
                             colour_me_green, colour_me_normal)
            fmt.Println(well_done)
            return
        }
    }
}

func sleep(secs int, msg string) {
    if msg != "" {
        fmt.Printf("%s - Sleep %ds ...", msg, secs)
    } else {
        fmt.Printf("Sleep %ds ... ", secs)
    }

    delay := time.Duration(secs) * 1000 * time.Millisecond
    time.Sleep(delay)
    //fmt.Println("Done")
}

func debug(msg string) {
    if dbg {
        fmt.Printf("%sDEBUG:%s<<%s>>DEBUG\n", colour_me_cyan, msg, colour_me_normal)
    }
}

func showVersion() {
    fmt.Println("Version: " + __K8SCENARIO_VERSION__ + "/" + __DATE_VERSION__)
}

func press(prompt string) {
    if prompt != "" {
        fmt.Print(prompt)
    }
    reader := bufio.NewReader(os.Stdin)
    fmt.Print("Press <enter> to continue ...")
    _, _ = reader.ReadString('\n')
}

func wait_on_label() {
    op      := ""

    for true {
        op, _ = silent_exec( fmt.Sprintf("kubectl get namespace %s --no-headers --show-labels", *namespace) )
	fmt.Print("wait_on_label: " + op)
	if strings.Contains(op, "regtest=ok") {
	    return;
        }
        sleep(CHECK_FIXED_SLEEP_SECS, "wait_on_label")
    }
}

func main() {
    flag.BoolVar(&version, "version",  false, "Show version string (false)")
    flag.BoolVar(&menu,    "menu",     false, "Menu to select a scenario")
    flag.BoolVar(&verbose, "verbose ", false, "Verbose mode")
    flag.BoolVar(&dbg,     "debug",    false, "Debug mode")
    //flag.BoolVar(&wait_enter, "wait",  false, "Wait on enter after scenario OK")
    flag.BoolVar(&wait_label, "wait",  false, "Wait on namespace label regtest=ok after scenario OK")

    flag.Parse()

    if version {
        showVersion()
        os.Exit(0)
    }
    if dbg {
        fmt.Println("serverUrl: " + *serverUrl)
    }

    if len(os.Args) == 1 {
         menu=true
    }


    if *zipFile != "" {
        path_elems := strings.Split(*zipFile, "/scenario")
	zipFileNum := path_elems[ len(path_elems)-1 ]
	zipFileNum  = strings.TrimSuffix(zipFileNum, ".zip")

	var err error
	*scenario, err = strconv.Atoi(zipFileNum)
        if err != nil {
            panic(err.Error())
	}

        check_script, instructions, challenge_type := install_scenario_zip(*zipFile, *scenario)

        loop_check(check_script, instructions, challenge_type, *scenario)
        if wait_label {
            wait_on_label()
            //press("")
        }
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

    _, _, _ = install_scenario(*scenario)


    if incluster {
        sleep(INCLUSTER_SLEEP_SECS, "in cluster")
    }

    os.Exit(0)
}




