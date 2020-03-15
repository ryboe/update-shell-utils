# update-shell-utils

[![CircleCI](https://circleci.com/gh/y0ssar1an/update-shell-utils/tree/master.svg?style=svg)](https://circleci.com/gh/y0ssar1an/update-shell-utils/tree/master)

This was made for my Mac, but you're welcome to fork it and modify it for your
system.

Run these common upgrade commands in parallel:

```sh
- brew update
  brew upgrade
- pip install --upgrade pip setuptools wheel
- rustup update
  cargo install <pkgs>
- softwareupdate -ia
```
