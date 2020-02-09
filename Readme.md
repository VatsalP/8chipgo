# 8Chipgo

Chip8  emulator made in Go using [pixel](https://github.com/faiface/pixel) library

![zero.gif](zero.gif)

```
git clone https://github.com/VatsalP/8chipgo
cd 8chipgo
go get ./..
go build 8chip.go vm.go
```

You can try some roms from [here](https://github.com/dmatlack/chip8/tree/master/roms)

```
./8chip -rom romfile
```

There are additional requirements for building pixel (see more in pixel [repo](https://github.com/faiface/pixel)):

- On macOS, you need Xcode or Command Line Tools for Xcode (xcode-select --install) for required headers and libraries.
- On Ubuntu/Debian-like Linux distributions, you need libgl1-mesa-dev and xorg-dev packages.
- On CentOS/Fedora-like Linux distributions, you need libX11-devel libXcursor-devel libXrandr-devel libXinerama-devel mesa-libGL-devel libXi-devel packages.