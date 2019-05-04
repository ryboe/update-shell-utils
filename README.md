# update-shell-utils
[![CircleCI](https://circleci.com/gh/y0ssar1an/update-shell-utils/tree/master.svg?style=svg)](https://circleci.com/gh/y0ssar1an/update-shell-utils/tree/master)

This was made for my Mac, but you're welcome to fork it and modify it for your
system.

Run these common upgrade commands in parallel:

```sh
- brew update
  brew upgrade
- pip3 install --upgrade pip setuptools wheel
  pip3 install --upgrade --user poetry
- rustup update
- softwareupdate -ia
```
