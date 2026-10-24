# Build Directory

The build directory is used to house all the build files and assets for your application. 

The structure is:

* bin - Output directory
* darwin - macOS specific files
* windows - Windows specific files

## Mac

The `darwin` directory holds files specific to Mac builds.
These may be customised and used as part of the build. To return these files to the default state, simply delete them
and
build with `wails build`.

The directory contains the following files:

- `Info.plist` - the main plist file used for Mac builds. It is used when building using `wails build`.
- `Info.dev.plist` - same as the main plist file but used when building using `wails dev`.

## Windows

The `windows` directory contains the manifest and rc files used when building with `wails build`.
These may be customised for your application. To return these files to the default state, simply delete them and
build with `wails build`.

- `icon.ico` - The icon used for the application. This is used when building using `wails build`. If you wish to
  use a different icon, simply replace this file with your own. If it is missing, a new `icon.ico` file
  will be created using the `appicon.png` file in the build directory.
- `installer/*` - The files used to create the Windows installer. These are used when building using `wails build`.
- `info.json` - Application details used for Windows builds. The data here will be used by the Windows installer,
  as well as the application itself (right click the exe -> properties -> details)
- `wails.exe.manifest` - The main application manifest file.

## Windows installer

NEST uses its own NSIS script instead of `wails build -nsis`, so the branding,
per-user install and uninstall behaviour stay under our control.

The installer is a **per-user** install (`%LOCALAPPDATA%\Programs\NEST`), so it never
prompts for administrator rights. It adds a Start Menu shortcut, an optional desktop
shortcut, an Add/Remove Programs entry, and a clean uninstaller that can keep the
user's settings.

Requirements: [NSIS 3](https://nsis.sourceforge.io/) and a built `bin/nest.exe`.

```bash
wails build
cd build/windows/installer
./build.sh            # or: makensis installer.nsi
```

Output: `build/bin/NEST-<version>-win64-installer.exe`.

The script accepts an optional version override:

```bash
./build.sh 1.2.1
```

Branded Modern UI bitmaps live in `windows/installer/resources/` (`welcome.bmp` is
164x314, `header.bmp` is 150x57). The installer intentionally has no version baked
into the artwork, so the bitmaps only need regenerating when the brand changes.