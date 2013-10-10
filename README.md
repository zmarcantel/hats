hats
====

Personal and portable work environment manager for whatever hat you're wearing

___Notice___: This software is still in heavy devlopment.

# Usage

## HEAD Files

Every hat needs a head, so HEAD files describe the base on which all your hats sit.

For instance, I'm keen on having:
* Pacman as my package manager
* Frequent software repository updates
* Vim as a text editor
* Basic development tools (make, gcc, autoconf)

on __EVERY__ machine I work on/with. It's entirely personal choice, and that's the beauty of `hats`.

Example:

```yaml
base:

  # What should Docker tag this image as?
  name: mydev/base

  # I like to start with @rahulg base image.
  image: rahulg/arch

  # If using any base other than the Docker-provided 'ubuntu'
  # you must specify the package manager command
  # Removing this need is on the todo list
  package_manager: pacman

  # Include these packages no matter what hat I'm wearing
  install:
    - python
    - python-pip
    - gcc
    - make
    - vim
    - zsh

  # These commands will happen when the environment starts
  startup:
    - echo "Welcome $USER."

  # These ports will open to the host machine
  # For use with personal APIs or anything you can imagine
  # EXAMPLES BELOW IN README
  expose:
    - 9001
```

HEAD files, as you may notice, are simply YAML. Simple enough right?

To create the base `Docker` image:

    wear -H my.head

This tells `hats` to build a head image (`-H`) using the HEAD descriptor in the file `my.head`.

If you supply the `-v` flag, the Dockerfile will be printed out as well.
