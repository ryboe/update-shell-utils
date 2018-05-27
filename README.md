# update-shell-utils

Run these common upgrade commands concurrently:

```sh
- brew update
  brew upgrade
  brew cleanup -s
  brew prune
- pip3 install --upgrade
  pip2 install --upgrade
- go get -u <bunch-of-go-binaries>
- rustup update
- softwareupdate -ia
```

Also updates Sublime Text, all Sublime Text packages, and all neovim packages.
