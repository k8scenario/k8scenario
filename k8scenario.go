// +build ignore

package main

import (
    "fmt"
    "sort"
    //"flag"
    //"log"
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

var (
    // Build time constants:
    BUILD_TIME    string = ""
    BUILD_VERSION string = ""
    VERSION       string = ""
    DEFAULT_URL   string = "" // Default url used to download scenarii
)

const (
    __DEFAULT_NAMESPACE__ = "k8scenario"

    // time to wait sleeping if running in cluster
    INCLUSTER_SLEEP_SECS=3600

    CHECK_FIXED_SLEEP_SECS=5
    EARLY_CHECK_FIXED_SLEEP_SECS=2

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
    pubUrl = DEFAULT_URL
    defaultNamespace = __DEFAULT_NAMESPACE__

    // Globally accessible so we can easily append it to os/exec environment:
    kubeconfig = ""

    menu      = false
    version   = false
    incluster = false
    verbose   = false
    help      = false
    // If set to true:
    //   Once scenario is solved, don't delete namespace until enter has been pressed:
    wait_on_regtest_ok_label = false
    localServer  = false
    // anIntVal  = 1; anInt  = &anIntVal

    /*namespace = FLAG.String("namespace", "k8scenario", "The namespace to use (all)")
    scenario  = FLAG.Int("scenario", 1, "k8s scenario to run (default: 1)")
    serverUrl = FLAG.String("server", pubUrl, "Get scenarii from specified server")
    localDir  = FLAG.String("dir", "", "Get scenarii from local dir")
    zipFile   = FLAG.String("zip", "", "Get scenarii from local zip file")
    localServer = FLAG.Bool("localServer", false, "Get scenarii from local server")
    */
    namespace = &defaultNamespace
    serverUrl = &pubUrl
    scenarioVal  = -1; scenario  = &scenarioVal
    localDirStr = ""; localDir  = &localDirStr
    zipFileStr = "";  zipFile   = &zipFileStr

    USER = os.Getenv("USER")
    tempfile = fmt.Sprintf("/tmp/tempfile.%s", USER)
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

func silent_assert_exec(command string) (string,int) {
    showOP := false
    assert := true
    return _exec(showOP, assert, command)
}

func silent_exec_pipe(pipeCmd string) (string,int) {
    showOP := false
    assert := false
    return _exec_pipe( showOP, assert, "/bin/bash", "-c", pipeCmd )
}
func assert_exec_pipe(pipeCmd string) (string,int) {
    showOP := true
    assert := true
    return _exec_pipe( showOP, assert, "/bin/bash", "-c", pipeCmd )
}
func exec_pipe(pipeCmd string) (string,int) {
    showOP := true
    assert := false
    return _exec_pipe( showOP, assert, "/bin/bash", "-c", pipeCmd )
}

func _exec_pipe(showOP bool, assert bool, pipeCmd ...string) (string, int) {
    head := pipeCmd[0]
    parts := pipeCmd[1:len(pipeCmd)]

    output, err := exec.Command(head,parts...).CombinedOutput()
    if err != nil {
        msg := fmt.Sprintf("_exec_pipe: exec.Command(PIPE) returned error with %s\n", err.Error())
        if assert {
            FatalError(msg)
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
            //FatalError("exec.ExitError: %v", err)
            showOP=true
            fmt.Printf("FATAL: exec.ExitError: %v", err)
            exit_status_int = -1 // artificially set to error to avoid false success detection
        }
    }

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
    if len(command) > 80 {
        debug("_exec([part] " + command[:80] + " ... )\n")
    } else {
        debug("_exec( " + command + " )\n")
    }

    parts := strings.Fields(command)
    head := parts[0]
    parts = parts[1:len(parts)]

    output, err := exec.Command(head,parts...).CombinedOutput()
    if err != nil {
        msg := fmt.Sprintf("_exec: exec.Command(PIPE) returned error with %s\n", err.Error())
        if assert {
            FatalError(msg)
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

func install_scenario(scenario int) (string, string, string, string) {
    //debug(fmt.Sprintf("Downloading scenario %d ...\n", scenario))
    //fmt.Sprintf("Downloading scenario %d ...\n", scenario)

    scenarioName := fmt.Sprintf("scenario%d", scenario)
    zipUrl      := fmt.Sprintf("%s/%s.zip", *serverUrl, scenarioName)
    zipFile := "/tmp/a.zip"

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
        FatalError("No such file: " + zipFile)
    }

    return install_scenario_zip(zipFile, scenario)
}

// FatalError: serious enough to stop everything (inc. regression tests)
func FatalError(msg string) {
    fmt.Println("FATAL: " + msg)
    os.Exit(99)
}

func readFileFromReader(f *zip.File) (string) {
    //fmt.Printf("Contents of %s:\n", f.Name)
    rc, err := f.Open()
    if err != nil {
        FatalError("readFileFromReader: Open error")
    }

    tempfile := "/tmp/tempfile.zzz"
    tmpfh, err := os.Create(tempfile)
    //_, _ = io.CopyN(tmpfh, rc, 2000)
    //_, _ = io.CopyN(tmpfh, rc, 10000)
    _, _ = io.CopyN(tmpfh, rc, 100000) // Now 100k bytes !!
    //if err != nil {
        // Dont' print EOF - fmt.Println(err)
        //fmt.Println("CopyN error")
        //FatalError(err)
    //}
    tmpfh.Close()

    bytes, readf_err := ioutil.ReadFile(tempfile)

    if readf_err != nil {
        fmt.Println("readFileFromReader: error on file - " + f.Name + ": error - " + readf_err.Error())
    }

    str := string(bytes)
    // debug("readFileFromReader(zip.File): " +str)

    rc.Close()

    return str
}

func write_check_script(check_script string, extra string) (string) {

    CHECK_SCRIPT := "#!/bin/bash\n" + "##__START__\n" +  extra + "#__START_check_script__\n" + string( check_script ) + "##__END__\n"
    //debug(CHECK_SCRIPT)

    // TODO: invoke SETUP_SCRIPT once with '-pre' argument - HERE
    return CHECK_SCRIPT
}

func apply_setup_script(scenario int, setup_script string, extra string, prepost_yaml string) {
    if setup_script == "" {
        return
    }

    SETUP_SCRIPT := "#!/bin/bash\n" + "##__START__\n" +  extra + "#__START_setup_script__\n" + string( setup_script ) + "##__END__\n"

    content := ""

    if prepost_yaml == "--pre-yaml" {
        set_args := "\nset -- --pre-yaml\n\n"
	content = set_args + SETUP_SCRIPT
        _, _ = exec_pipe(content)
    } else if prepost_yaml == "--post-yaml" {
        set_args := "\nset -- --post-yaml\n\n"
	content = set_args + SETUP_SCRIPT

	_, _ = exec_pipe(content)
    } else {
        content = SETUP_SCRIPT
        _, _ = exec_pipe(content)
    }
}

func writeFile(filename string, content string) {
    f, err := os.Create(filename)
    if err != nil {
        fmt.Println(filename + ": " + err.Error())
    }
    f.WriteString(content)
    //defer f.Close()
    f.Sync()
    f.Close()
}

func install_scenario_zip(zipFile string, scenario int) (string, string, string, string) {
    scenarioName := fmt.Sprintf("scenario%d", scenario)
    fmt.Print()
    showVersion()
    fmt.Printf("\n---- %s[%s]%s Installing into namespace %s%s%s\n",
               colour_me_yellow, scenarioName, colour_me_normal,
               colour_me_green, *namespace, colour_me_normal)
    // Open a zip archive for reading.
    r, err := zip.OpenReader(zipFile)
    if err != nil {
        FatalError("zip: " +  err.Error())
    }
    defer r.Close()


    pipeCmd := "kubectl get namespaces --no-headers " + *namespace + " | awk '{ print $1; }' "
    namespace_match, _ := exec_pipe(pipeCmd)
    if !strings.Contains(namespace_match, "NotFound")      {
        //fmt.Println( "namespace_match=<" + namespace_match +">" )
        fmt.Println( "Deleting existing: " + *namespace)
        silent_exec( fmt.Sprintf("kubectl label --overwrite namespace %s status=deleting scenario=%d loop-", *namespace, scenario) )
        assert_exec("kubectl delete namespace " + *namespace)
    }

    fmt.Println( "(Re)creating namespace: " + *namespace)
    assert_exec("kubectl create namespace " + *namespace)
    exec_pipe( fmt.Sprintf("kubectl get namespace %s --no-headers", *namespace) )
    INSTRUCTIONS := ""
    CHALLENGE_TYPE := "task"
    TAGS :=  ""
    SETUP_SCRIPT := ""
    CHECK_SCRIPT := ""
    FUNCTIONS_RC := ""
    EXPORT_NAMESPACE := "export NS="       + *namespace       + "\n"
    //EXPORT_SCENARIO := "export SCENARIO=" + string(scenario) + " # " + comment + "\n"
    EXPORT_SCENARIO  := fmt.Sprintf("export SCENARIO=%d\n", scenario)
    EXPORT_NOMARK    := "export NO_MARK_SCENARIO=DONT_MARK_NAMESPACE\n"

    // empty array of strings:
    YAML_FILES := []string{}
    YAML_FILE_NAMES := []string{}
    YAML_NO := 1

    file_loop := 0
    for _, f := range r.File {
        file_loop = file_loop  + 1
        strlen := len(f.Name)

	// Skip directories:
        if f.Name[strlen-1:] == "/" { continue; }

	i:=0
	for i=0; i<len(f.Name); i++ {
            if f.Name[i] == '/' { break; }
	}
	fileName := f.Name[i+1:]
	//fmt.Println(fileName)
        //if strings.Contains(f.Name, ".functions.rc")      { FUNCTIONS_RC = readFileFromReader(f) }
        //if strings.HasPrefix(f.Name, ".functions.rc")      { FUNCTIONS_RC = readFileFromReader(f) }

        // GATHER FUNCTIONS_RC:
        if fileName == ".functions.rc"      { FUNCTIONS_RC = readFileFromReader(f) }

        // GATHER SETUP_SCRIPT:
        if fileName == "SETUP_SCENARIO.sh"  { SETUP_SCRIPT = readFileFromReader(f) }

        // GATHER CHECK_SCRIPT:
        if fileName == "CHECK_SCENARIO.sh"  {
            CHECK_SCRIPT = readFileFromReader(f)
        }

        // GATHER INSTRUCTIONS:
        if fileName == "INSTRUCTIONS.txt"   { INSTRUCTIONS = readFileFromReader(f) }

        // GATHER CHALLENGE_TYPE:
        if fileName == "CHALLENGE_TYPE.txt" { CHALLENGE_TYPE = readFileFromReader(f) }

        // GATHER TAGS:
        if fileName == "TAGS.txt"           { TAGS = readFileFromReader(f) }

        if CaseInsensitiveContains(fileName, "yaml") || CaseInsensitiveContains(fileName, "yml")   {
            yaml_content   := readFileFromReader(f)
            YAML_FILE_NAMES = append(YAML_FILE_NAMES, f.Name)
            YAML_FILES      = append(YAML_FILES, yaml_content)
            YAML_NO = YAML_NO + 1
	        writeFile("/tmp/xx", yaml_content)
        }
    }

    CHALLENGE_TYPE = strings.TrimSuffix( CHALLENGE_TYPE, "\n")
    apply_setup_script(scenario,      SETUP_SCRIPT, EXPORT_NAMESPACE + EXPORT_SCENARIO + FUNCTIONS_RC + "\n", "--pre-yaml")
    CHECK_SCRIPT = write_check_script(CHECK_SCRIPT, EXPORT_NAMESPACE + EXPORT_SCENARIO + EXPORT_NOMARK + FUNCTIONS_RC + "\n")

    for file_idx := 0 ; file_idx < len(YAML_FILES); file_idx++ {
        filename := YAML_FILE_NAMES[file_idx]
        content  := YAML_FILES[file_idx]

        writeFile(tempfile, content)

        applyCmd := fmt.Sprintf("kubectl apply -n %s -f %s", *namespace, tempfile)
	fmt.Println(applyCmd)

        //fmt.Println( applyCmd )
        path_elems := strings.Split(filename, "/")
        filename   = path_elems[len(path_elems)-1]


        full_cmd := fmt.Sprintf("%s | grep -v ^$\n", applyCmd)
        //fmt.Printf("Command '%s'\n", full_cmd)
        exec_pipe( full_cmd )
        silent_exec( "rm " + tempfile )
    }
    apply_setup_script(scenario, SETUP_SCRIPT, EXPORT_NAMESPACE + FUNCTIONS_RC + "\n", "--post-yaml")

    debug("======== " + zipFile + "\n")
    silent_exec( fmt.Sprintf("kubectl label --overwrite namespace %s status=readytofix scenario=%d", *namespace, scenario) )

    return CHECK_SCRIPT, INSTRUCTIONS, CHALLENGE_TYPE, TAGS
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
        FatalError("No such file: " + listFile)
    }

    fileH, err := os.Open(listFile)
    if err != nil {
        FatalError("Failed opening file: " +  err.Error())
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

    sort.Slice(txtlines, func(a, b int) bool {
        valA, _ := strconv.Atoi(txtlines[a])
        valB, _ := strconv.Atoi(txtlines[b])
        return valA < valB
    })

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
            check_script, instructions, challenge_type, tags := install_scenario(scenario)

            _ = loop_check(check_script, instructions, challenge_type, scenario, tags)
        }
    }
}

func checkNamespaceForLabel(label_key string, label_value string) bool {
    pipeCmd := "kubectl get ns " + *namespace + " -o custom-columns=labels:.metadata.labels." + label_key + " --no-headers"
    namespace_match, _ := exec_pipe(pipeCmd)
    //fmt.Println("namespace_match=" + namespace_match)
    namespace_match = strings.TrimSuffix( namespace_match , "\n")
    //fmt.Println("namespace_match=" + namespace_match)
    if namespace_match == label_value {
        fmt.Println("Matched " + label_key + "=" + label_value)
        return true
    }
    return false
}

func loop_check(check_script string, instructions string, challenge_type string, scenario int, tags string) bool {
    scenarioName := fmt.Sprintf("scenario%d", scenario)
    abandon := false

    if check_script == "" {
        fmt.Printf("---- %s[%s]%s NO Cluster check script - choose new scenario when you want to move on\n",
                   colour_me_yellow, scenarioName, colour_me_normal)
        return abandon
    }

    prompt := ""
    switch challenge_type {
        case "task": prompt = "task incomplete"
        case "fix": prompt = "fix incomplete"
        default: prompt = "**** incomplete"
    }

    // Loop until check_script produces and error
    early_err_code := 1
    for early_err_code == 0 {
        //_, early_err_code = exec_pipe(check_script)
        _, early_err_code = silent_exec(check_script)
        if early_err_code == 0 {
            sleep(EARLY_CHECK_FIXED_SLEEP_SECS, "Setting up error condition ...\n")
        }
    }

    //for_check_loop:
    loop_num := 0
    for true {
        loop_num = loop_num + 1

	// Abandon test if namespace has label abandon=true
	//     used by Regression Testing to signal to (give up on this test) move on ...
        if checkNamespaceForLabel("abandon", "true") {
            abandon=true
            return abandon
        }

        // Only show instructions once very 10 loops ...
        if (loop_num % 10) == 1 {
            fmt.Printf("\n%s", instructions)
        }

        sleep(CHECK_FIXED_SLEEP_SECS, 
              fmt.Sprintf("\n%s[%s]/%d%s - %s%s%s",
                  colour_me_yellow, scenarioName, loop_num, colour_me_normal,
                  colour_me_red, prompt, colour_me_normal) )

        _, err_code := exec_pipe(check_script)
        writeFile("/tmp/k8scenario.go.check."+scenarioName+".sh", check_script);

        if err_code == 0 {
            well_done := fmt.Sprintf("\n---- %s[%s]%s %sWELL DONE !!!!%s - The scenario appears to be fixed !!!!\n",
                             colour_me_yellow, scenarioName, colour_me_normal,
                             colour_me_green, colour_me_normal)
            fmt.Println(well_done)
            return abandon
        }
    }
    return abandon
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
}

func showVersion() {
    versionStr := fmt.Sprintf("k8scenario: %s [Built at %s]\n", VERSION, BUILD_TIME)
    fmt.Println(versionStr)
}

func press(prompt string) {
    if prompt != "" {
        fmt.Print(prompt)
    }
    reader := bufio.NewReader(os.Stdin)
    fmt.Print("Press <enter> to continue ...")
    _, _ = reader.ReadString('\n')
}

func wait_on_regtest_ok_label_fn() {
    op      := ""

    silent_exec( fmt.Sprintf("kubectl label --overwrite namespace %s loop=%d", *namespace, 1000) )
    for true {
        op, _ = silent_exec( fmt.Sprintf("kubectl get namespace %s --no-headers --show-labels", *namespace) )
        fmt.Print("wait_on_regtest_ok_label_fn: " + op)
        if strings.Contains(op, "regtest=ok") {
            return;
        }
        sleep(CHECK_FIXED_SLEEP_SECS, "wait_on_regtest_ok_label_fn")
    }
}

func main() {
    fmt.Println("Using default URL: " + DEFAULT_URL)

    bool_args := map[string]*bool{
        "help": &help,
        "verbose": &verbose,
        //"debug": &dbg,
        "menu": &menu,
        "wait_ok": &wait_on_regtest_ok_label,
        "version": &version,
        "localServer": &localServer,
    }
    str_args := map[string]*string{
        "namespace": namespace,
        "serverUrl": serverUrl,
        "dir":       localDir,
        "zip":       zipFile,
    }
    num :=0 // just place-holder
    int_args := map[string]*int{
        "num": &num,
        "scenario": scenario,
    }

    // accept boolean args in form:
    //   option
    //   -option
    //   --option
    // accept string/integer args in form:
    //   option value
    //   -option value
    //   --option value

    for a := 1; a < len(os.Args); a++ {
        // fmt.Printf("---- Treating arg '%s'\n", os.Args[a]);
        option := os.Args[a]
        if verbose  { fmt.Printf("Treating arg '%s'\n", option); }

        if option[0] == '-' { option=option[1:]; }
        if option[0] == '-' { option=option[1:]; }

        ok := false
        _, ok = bool_args[option];
        if ok {
            if verbose { fmt.Printf("Found boolean option %s\n", option); }
            *bool_args[option]=true
            //fmt.Printf("bool %s=%t\n", option, *bool_args[option])
            //ok = true;
            continue;
        } else {
            if verbose { fmt.Printf("No such boolean option %s\n", option); }
	}

	_, ok = str_args[option];
        if ok {
            a=a+1
            if verbose { fmt.Printf("Found string option %s\n", option); }
            *str_args[option]=os.Args[a]
            //fmt.Printf("string %s='%s'\n", option, *str_args[option])
            //ok = true;
            continue;
        } else {
            if verbose { fmt.Printf("No such string option %s\n", option); }
        }

	_, ok = int_args[option];
	if ok {
	    a=a+1
	    if verbose { fmt.Printf("Found integer option %s\n", option); }
	    *int_args[option],_ = strconv.Atoi(os.Args[a])
	    ////fmt.Printf("integer %s=%d\n", option, *int_args[option])
	    //ok = true;
	    continue;
	} else {
	    if verbose { fmt.Printf("No such integer option %s\n", option); }
	}

	fmt.Printf("No such option %s\n", option)
        os.Exit(1)
    }

    if *scenario == -1 {
         menu=true
    }

    if help {
        fmt.Println( "Usage: " +  os.Args[0] )
        for item := range bool_args {
            // HIDE debug option!!:
            if item == "debug" { continue; }
            fmt.Printf("-%s [true|false] value: %t\n", item, *bool_args[item])
        }
        for item := range str_args  { fmt.Printf("-%s [<string>]   value: %s\n", item, *str_args[item]) }
        for item := range int_args  { fmt.Printf("-%s [<integer>]  value: %d\n", item, *int_args[item]) }
        os.Exit(0)
    }

    if verbose {
        for item := range bool_args {
            // HIDE debug option!!:
            if item == "debug" { continue; }
            fmt.Printf("-%s [true|false] value: %t\n", item, *bool_args[item])
        }
        for item := range str_args  { fmt.Printf("str  %s => %s\n", item, *str_args[item]) }
        for item := range int_args  { fmt.Printf("int  %s => %d\n", item, *int_args[item]) }
    }

    if version {
        showVersion()
        os.Exit(0)
    }

    if len(os.Args) == 1 {
        menu=true
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

        check_script, instructions, challenge_type, tags := install_scenario_zip(*zipFile, *scenario)

	abandon := loop_check(check_script, instructions, challenge_type, *scenario, tags)
	if abandon {
            fmt.Println("Abandoning failed test ...")
            os.Exit(0)
        }
        if wait_on_regtest_ok_label {
            wait_on_regtest_ok_label_fn()
            //press("")
        }
        os.Exit(0)
    }

    if *localDir != "" {
        *serverUrl = "file:///" + *localDir
    }

    if localServer {
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

    _, _, _, _ = install_scenario(*scenario)

    if incluster {
        sleep(INCLUSTER_SLEEP_SECS, "in cluster")
    }

    os.Exit(0)
}

