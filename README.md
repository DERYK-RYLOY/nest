<p align="center">
  <img src="docs/assets/logo.png" width="180" alt="NEST logo — the orange pouch that holds everything" />
</p>

<h1 align="center">NEST — File Organizer</h1>

<p align="center"><strong>The file organizer that cleans up, then disappears.</strong></p>

<p align="center">
  <img src="https://img.shields.io/badge/version-1.2.0-ff8a3d" alt="version" />
  <img src="https://img.shields.io/badge/platform-Windows%2010%20%7C%2011-0078D4" alt="platform" />
  <img src="https://img.shields.io/badge/Go-1.25-00ADD8?logo=go&logoColor=white" alt="Go" />
  <img src="https://img.shields.io/badge/Wails-v2-2DD4BF" alt="Wails" />
  <img src="https://img.shields.io/badge/React%20%2B%20TypeScript-19-61DAFB" alt="React" />
  <img src="https://img.shields.io/github/downloads/DERYK-RYLOY/nest/total?color=34d399" alt="downloads" />
  <img src="https://img.shields.io/badge/license-Proprietary-lightgrey" alt="license" />
</p>

<p align="center">
  <a href="#-download"><strong>⬇ Download the ready-to-run EXE</strong></a> — no installer, no setup, no background service.
</p>

---

NEST watches your messy folders — Downloads by default — and puts every file where it
belongs: pictures to **Pictures**, videos to **Videos**, music to **Music**, documents to
**Documents**, and everything left over into freshly created family folders
(`_Archives`, `_Code`, `_Misc`, …). Then it quits. No tray icon. No background process.
One toast, and silence.

## ✨ Features

- **Real Windows folders** — resolves Pictures/Videos/Music/Documents/Downloads through
  the Windows shell API. No hardcoded paths, works on any PC and any username.
- **Leftover clustering** — unknown file types are grouped; 3+ of a kind get their own
  folder, stragglers land in `_Misc`. Nothing stays loose.
- **Your rules** — point any extension at any Windows folder or custom cluster.
  Rules always beat the built-ins.
- **Daily schedule, 1×–N× a day** — one Task Scheduler entry per time you pick
  (`NEST Auto 0800`, …). NEST wakes, sorts, toasts, exits.
- **Instant mode** — a logon task waits with **zero CPU** until a download lands,
  organizes it after a settle delay, and goes quiet again.
- **Preview before anything moves**, full **history with one-click undo** and
  **Clear history** that resets every stat.
- **Reveal in Explorer**, per-category stats, dark + light glass themes.
- **Single ~13 MB portable EXE.** No installer. No runtime to install (Windows 10/11
  already ship WebView2).

## 🆕 What's new in 1.2.0

A hardening release — same tiny EXE, fewer ways to lose a file or a schedule.

- **Undo covers the whole run.** One pass over several folders is now one undoable
  action, and undo reports files it could not restore instead of claiming success.
- **Nothing is ever overwritten, even on a race.** Moves claim the destination
  exclusively (including cross-volume moves); a late file at that name is never replaced.
- **Crash-safe settings and history.** `config.json` and `history.json` are written to a
  temp file and renamed into place, so a power cut cannot leave half a JSON file.
- **One Save for automation.** Times, instant mode, delay, folders and smart rules are
  stored and registered with Task Scheduler together, and rolled back if Windows refuses.
  Instant mode without admin now explains exactly what to do.
- **Preview tells the truth.** Collisions show the final `file (1).zip` name, and the
  reveal button points at the original file instead of a destination that does not exist yet.
- **Empty folder list stays empty**, duplicate folders are ignored, rule targets are
  validated, and folder arguments are checked before Explorer is opened.
- **`nest --version` and `nest --help`** for headless diagnostics.
- **Non-destructive test suite.** Task Scheduler and toast tests are opt-in with
  `NEST_LIVE_TESTS=1`; the default run touches only temp folders.

## 🖥 How it looks

Dark and light aurora-glass themes, one orange pouch, zero clutter:

| Dashboard | Automation | History |
|---|---|---|
| Live stats, where-files-go map, top categories | Daily times, instant mode, watched folders, admin status | Every move logged, undo, reveal, clear |

> Screenshots land here on first launch week — star ⭐ the repo and check back.

## 🚀 Download

Get the portable build from [**Releases**](https://github.com/DERYK-RYLOY/nest/releases):
**`NEST-v1.2.0-win64.zip`** → unzip → double-click `nest.exe`. That's the whole install.

> Enabling the daily schedule works as a standard user. **Instant mode needs
> administrator rights once** — the app offers a one-click *Restart as administrator*.

## 📖 Usage

1. **Preview** (top bar) — see exactly what would move. Nothing changes.
2. **Organize Now** — one pass over every watched folder.
3. **Automation page** — pick daily times (presets: 1× midday, 2×, 3×), toggle
   **instant mode**, add watched folders, tune the smart rules.
4. **Categories page** — browse the extension families and add **your rules**
   (e.g. `.blend` → `_Blender`, `.psd` → `pictures`).
5. **History page** — undo the last run, reveal any file, or clear the log.

### Safety rules (always on)

- Never touches subfolders, hidden/system files, or `desktop.ini`
- Skips in-progress downloads (`.crdownload`, `.part`, …)
- Skips files modified in the last 60 s (still being written) and retries them
- Collisions get numbered names (`file (1).zip`) — **nothing is ever overwritten**
- Undo restores without overwriting anything new

## 🛠 Built with — tools & techniques

Like a real shipping desktop product, NEST is a thin, dependency-light stack:

| Layer | Tool | Used for |
|---|---|---|
| Core | **Go 1.25** (stdlib + `x/sys/windows`) | everything below |
| Shell integration | `SHGetKnownFolderPath` | real Pictures/Videos/Music/Documents paths |
| Filesystem watch | `ReadDirectoryChangesW` | zero-CPU instant mode |
| Automation | Windows Task Scheduler (`schtasks`) | daily + logon tasks, no service |
| Notifications | WinRT toasts (`go-toast`) | the single "organized" note |
| Elevation | process token inspection + `runas` relaunch | admin only where Windows demands it |
| UI shell | **Wails v2** (WebView2, single static EXE) | Go ↔ React bridge, app packaging |
| UI | **React 19 + TypeScript + Vite** | dashboard, code-split views |
| Design | hand-built CSS glass system | complementary teal/orange scale, dark + light |
| Tests | `go test` (temp-dir only) | organizer, watcher, scheduler, toasts |
| Shipping | `wails build` + GitHub Releases | portable ZIP, no installer |

Project layout:

```text
nest/
├── main.go / app.go          # entrypoint: UI, --auto, --watch
├── internal/
│   ├── classifier/           # extension → destination rules
│   ├── organizer/            # scan, atomic move, undo log
│   ├── engine/               # shared passes (UI + auto + watch)
│   ├── knownfolders/         # Windows shell folders
│   ├── scheduler/            # task CRUD, elevation, admin relaunch
│   ├── watch/                # zero-CPU directory watching
│   ├── notify/               # toast notifications
│   └── config/               # %APPDATA%\Nest\config.json
└── frontend/src/             # React: Dashboard, Categories, Automation, History
```

## 🔧 Build from source

Requirements: Go 1.25+, Node 18+, Wails CLI v2.

```powershell
go install github.com/wailsapp/wails/v2/cmd/wails@latest
wails build          # → build/bin/nest.exe
go test ./internal/...   # safe: temp folders only, never your files
```

Want the integration tests too? `NEST_LIVE_TESTS=1 go test ./internal/...` registers a
real `NEST Auto 23:59` task and pushes a real toast, then restores the schedule and
watcher state it found. The default run never touches Windows automation.

## ❓ FAQ

**Does NEST run in the background?**
No. Scheduled runs and instant mode are Windows tasks that start `nest.exe`, which
exits when done. The watcher blocks on a kernel call with 0% CPU while idle.

**Can it delete or overwrite my files?**
No code path deletes user files or overwrites. Failed moves keep the original in place.

**Why does Instant mode ask for admin?**
Windows requires elevation to register logon-triggered tasks. Daily schedules do not.

**Which Windows versions?**
Windows 10 and 11, 64-bit (WebView2 ships with both).

## 👤 Credits

Concept, design system, asset direction and engineering — **DERYK RYLOY**.

## 📄 License

Proprietary — Copyright © 2026 DERYK RYLOY. All rights reserved.
See [LICENSE](LICENSE). Binaries may only be distributed via official NEST releases.
