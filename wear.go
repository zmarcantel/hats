package main

import (
  "os"
  "io"
  "fmt"
  "flag"
  "errors"
  "strings"
  "os/exec"
  "os/user"
  "io/ioutil"

  "launchpad.net/goyaml"
)

//=======================================
// Type Definitions
//=======================================

type Head struct {
  Base struct {
    Name            string

    Image           string
    Update          bool
    Install         []string
    PackageManager  string `yaml:"package_manager"`

    Startup         []string
    Expose          []string
  }
}


//=======================================
// Program Entry
//=======================================

var dot_dir string

func main() {
  var dockerfile string     // final script to be used

  // specify command line flags
  var is_verbose = flag.Bool("v", false, "Display the Dockerfile before imaging")
  var is_dry_run = flag.Bool("d", false, "Dry run -- Show steps with no action")
  var head_filename = flag.String("H", "", "Make HEAD using given config file")
  flag.Parse()

  usr, err := user.Current()
  if err != nil { handleError(err) }
  dot_dir = usr.HomeDir + "/.hats"

  if *head_filename != "" {
    // load the HEAD file
    var head_file = readHeadFile(*head_filename)

    // build up the Dockerfil
    dockerfile += doBaseImage(head_file)
    dockerfile += doBaseInstall(head_file)
    dockerfile += doBaseStartup(head_file)
    dockerfile += doBasePortExpose(head_file)

    if *is_verbose {
      fmt.Println(dockerfile)
    }

    writeDockerfile(dockerfile, head_file.Base.Name)
    if !(*is_dry_run) {
      doDockerBuild(head_file.Base.Name)
    }

    err := ioutil.WriteFile(dot_dir + "/last.used", []byte(head_file.Base.Name), 0744)
    if err != nil { handleError(err) }
  }
}

//
// Generic Error Handler
//
func handleError(err error) {
  fmt.Printf("\n\nError:\n%s\n\n", err.Error())
  os.Exit(1);
}


//
// Parse the HEAD file
//
func readHeadFile(filename string) Head {
  head_bytes, err := ioutil.ReadFile(filename)
  if err != nil { handleError(err) }

  var head = Head{}
  err = goyaml.Unmarshal(head_bytes, &head)
  if err != nil { handleError(err) }

  return head
}


//
// Extract the base image info
//
func doBaseImage(head_file Head) string {
  if head_file.Base.Image == "" { handleError(errors.New("No base image specified")) }

  // check if base Ubuntu: if not, we have some workarounds
  if head_file.Base.Image != "ubuntu" {
    if head_file.Base.PackageManager == "" {
      handleError(errors.New("Base field [package_manager] required with non-'ubuntu' base iamges."))
    }
  } else {
    // Ubuntu Defaults
    if head_file.Base.PackageManager == "" { head_file.Base.PackageManager = "apt-get" }
  }

  return "FROM " + head_file.Base.Image + "\n\n\n"
}


//
// Loop through Base.Install list and set install commands
//
func doBaseInstall(head_file Head) string {
  var result string
  var no_confirm string
  var update_cache string
  var install_command string

  // get package manager specific variables set
  // TODO: expand this section
  if head_file.Base.PackageManager == "" {
    head_file.Base.PackageManager = "apt-get"
  }
  switch(head_file.Base.PackageManager) {
    case "apt-get":
      no_confirm = "-y"
      update_cache = "update"
      install_command = "install"
      break
    case "pacman":
      no_confirm = "--noconfirm"
      update_cache = "-Sy"
      install_command = "-S"
      break
  }

  var package_manager = head_file.Base.PackageManager
  var package_list = strings.Join(head_file.Base.Install, " ")

  if head_file.Base.Update {
    result += fmt.Sprintf("RUN %s %s %s\n", package_manager, update_cache, no_confirm)
  }
  result += fmt.Sprintf("RUN %s %s %s %s\n", package_manager, install_command, no_confirm, package_list)

  return result + "\n\n"
}


//
// Loop through the startup items
//
func doBaseStartup(head_file Head) string {
  var result string
  for i := range head_file.Base.Startup {
    result += fmt.Sprintf("ENTRYPOINT %s\n", head_file.Base.Startup[i])
  }

  return result + "\n\n"
}


//
// Loop through the ports that should be exposed
//
func doBasePortExpose(head_file Head) string {
  var result string
  for i := range head_file.Base.Expose {
    result += "EXPOSE " + head_file.Base.Expose[i] + "\n"
  }

  return result + "\n\n"
}


//
// Write the Dockerfile to disk
//
func writeDockerfile(dockerfile, name string) {
  var image_dot_dir = dot_dir + "/" + name

  err := os.MkdirAll(image_dot_dir, 0744)
  if err != nil { handleError(err) }

  err = ioutil.WriteFile(image_dot_dir + "/Dockerfile", []byte(dockerfile), 0744)
  if err != nil { handleError(err) }
}


//
// Build the docker image
//
func doDockerBuild(name string) {
  var dockerfile_dir = dot_dir + "/" + name

  var cmd *exec.Cmd
  if name != "" {
    cmd = exec.Command("docker", "build", "-t", name, dockerfile_dir)
  } else {
    cmd = exec.Command("docker", "build", dockerfile_dir)
  }

  stdout, err := cmd.StdoutPipe()
  if err != nil { handleError(err) }

  stderr, err := cmd.StderrPipe()
  if err != nil { handleError(err) }

  err = cmd.Start()
  if err != nil { handleError(err) }

  go io.Copy(os.Stdout, stdout)
  go io.Copy(os.Stderr, stderr)
  cmd.Wait()
}
