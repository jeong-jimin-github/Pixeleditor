# Pixel Editor Pro

Tkinter + Pillow 픽셀 에디터를 **Go + Fyne**으로 재작성한 데스크톱 앱입니다.

원본 Python 구현은 [`legacy/app.py`](legacy/app.py)에 있습니다.

## Features

- Pen / Eraser / Fill / Pipette (`P` `E` `F` `I`)
- X-axis symmetry
- Color picker, image palette, lock palette
- Open PNG/JPEG/BMP, Save / Save As PNG (`Ctrl+S`)
- Undo / Redo (`Ctrl+Z` / `Ctrl+Y`)
- Grid, grid gap, zoom (`+` `-`, mouse wheel)
- Space or middle-mouse drag to pan
- Checkerboard for transparency

## Run

```bash
go run .
```

## Build

```bash
# Windows GUI (no console)
go build -ldflags "-s -w -H windowsgui" -o pixeleditor.exe .

# Linux / macOS
go build -ldflags "-s -w" -o pixeleditor .
```

Needs CGO (a C compiler). On Windows use MinGW; on Linux install `gcc`, `libgl1-mesa-dev`, `xorg-dev`, `libxkbcommon-dev`.

## Release

Every push to `main` runs [`.github/workflows/release.yml`](.github/workflows/release.yml):

1. `go test ./internal/doc`
2. Build Windows x64, Linux x64, and macOS Apple Silicon binaries
3. Publish a GitHub Release tagged `build-<run number>`
