; NEST File Organizer - Windows installer
; Build with: makensis installer.nsi   (from this folder)
; Requires build\bin\nest.exe from `wails build`.

Unicode true
SetCompressor /SOLID lzma

!include "MUI2.nsh"
!include "LogicLib.nsh"
!include "x64.nsh"
!include "WinVer.nsh"
!include "FileFunc.nsh"

!define APP_NAME       "NEST File Organizer"
!define APP_SHORT      "NEST"
; Override on the command line with /DAPP_VERSION=x.y.z when releasing.
!ifndef APP_VERSION
  !define APP_VERSION "1.2.0"
!endif
!define APP_PUBLISHER  "DERYK RYLOY"
!define APP_URL        "https://github.com/DERYK-RYLOY/nest"
!define APP_EXE        "nest.exe"
!define APP_REG_KEY    "Software\DERYK RYLOY\NEST"
!define UNINST_KEY     "Software\Microsoft\Windows\CurrentVersion\Uninstall\NEST File Organizer"

!define SOURCE_EXE     "${__FILEDIR__}\..\..\bin\nest.exe"
!define SOURCE_README  "${__FILEDIR__}\..\..\..\README.md"
!define SOURCE_LICENSE "${__FILEDIR__}\..\..\..\LICENSE"
!define BRAND_DIR      "${__FILEDIR__}\resources"

Name "${APP_NAME}"
Caption "${APP_NAME} Setup"
OutFile "${__FILEDIR__}\..\..\bin\NEST-v${APP_VERSION}-win64-installer.exe"
InstallDir "$LOCALAPPDATA\Programs\${APP_SHORT}"
InstallDirRegKey HKCU "${APP_REG_KEY}" "InstallDir"
RequestExecutionLevel user
BrandingText "${APP_SHORT} ${APP_VERSION}  |  ${APP_PUBLISHER}"
Icon "${__FILEDIR__}\..\icon.ico"
UninstallIcon "${__FILEDIR__}\..\icon.ico"
WindowIcon on
ShowInstDetails nevershow
ShowUninstDetails nevershow
AutoCloseWindow false
CRCCheck on
XPStyle on
ManifestDPIAware true

VIProductVersion "${APP_VERSION}.0"
VIFileVersion    "${APP_VERSION}.0"
VIAddVersionKey "CompanyName"     "${APP_PUBLISHER}"
VIAddVersionKey "FileDescription" "${APP_NAME} Setup"
VIAddVersionKey "ProductVersion"  "${APP_VERSION}"
VIAddVersionKey "FileVersion"     "${APP_VERSION}"
VIAddVersionKey "LegalCopyright"  "Copyright (c) 2026 ${APP_PUBLISHER}"
VIAddVersionKey "ProductName"     "${APP_NAME}"
VIAddVersionKey "Comments"        "File organizer for Windows 10 and 11. Installs for the current user."

; ---- Modern UI styling -----------------------------------------------------
!define MUI_ICON "${__FILEDIR__}\..\icon.ico"
!define MUI_UNICON "${__FILEDIR__}\..\icon.ico"
!define MUI_ABORTWARNING
!define MUI_BGCOLOR "05090F"
!define MUI_TEXTCOLOR "EAF6F3"
!define MUI_HEADERIMAGE
!define MUI_HEADERIMAGE_RIGHT
!define MUI_HEADERIMAGE_BITMAP "${BRAND_DIR}\header.bmp"
!define MUI_HEADERIMAGE_UNBITMAP "${BRAND_DIR}\header.bmp"
!define MUI_WELCOMEFINISHPAGE_BITMAP "${BRAND_DIR}\welcome.bmp"
!define MUI_FINISHPAGE_LINK_COLOR "FF8A3D"

!define MUI_WELCOMEPAGE_TITLE "Welcome to NEST"
!define MUI_WELCOMEPAGE_TEXT "NEST puts every download where it belongs - then disappears.$\r$\n$\r$\nThis installs NEST ${APP_VERSION} for your Windows account. No administrator prompt, no background service, and your existing settings are kept."
!define MUI_FINISHPAGE_TITLE "NEST is ready."
!define MUI_FINISHPAGE_TEXT "Drop files into Downloads and NEST takes it from there.$\r$\n$\r$\nSettings and history live in %APPDATA%\Nest."
!define MUI_FINISHPAGE_RUN "$INSTDIR\${APP_EXE}"
!define MUI_FINISHPAGE_RUN_TEXT "Launch NEST"
!define MUI_FINISHPAGE_LINK "NEST on GitHub"
!define MUI_FINISHPAGE_LINK_LOCATION "${APP_URL}"
!define MUI_FINISHPAGE_NOREBOOTSUPPORT
!define MUI_DIRECTORYPAGE_TEXT_TOP "NEST installs into your user folder, so Windows will not ask for administrator rights."
!define MUI_UNCONFIRMPAGE_TEXT_TOP "This removes NEST from your account. Files NEST already organized are never touched."

!insertmacro MUI_PAGE_WELCOME
!insertmacro MUI_PAGE_COMPONENTS
!insertmacro MUI_PAGE_DIRECTORY
!insertmacro MUI_PAGE_INSTFILES
!insertmacro MUI_PAGE_FINISH

!insertmacro MUI_UNPAGE_CONFIRM
!insertmacro MUI_UNPAGE_INSTFILES

!insertmacro MUI_LANGUAGE "English"

Var DeleteUserData

; ---- Pre-flight ------------------------------------------------------------
Function .onInit
  SetShellVarContext current

  ${IfNot} ${AtLeastWin10}
    IfSilent winverSilent
    MessageBox MB_ICONSTOP "NEST requires Windows 10 or Windows 11."
    Quit
    winverSilent:
    SetErrorLevel 64
    Abort
  ${EndIf}

  ${If} ${IsNativeAMD64}
  ${OrIf} ${IsNativeARM64}
  ${Else}
    IfSilent archSilent
    MessageBox MB_ICONSTOP "NEST requires 64-bit Windows 10 or Windows 11."
    Quit
    archSilent:
    SetErrorLevel 65
    Abort
  ${EndIf}

  ; WebView2 ships with Windows 10/11, but verify before copying 12 MB.
  SetRegView 64
  ReadRegStr $0 HKLM "SOFTWARE\WOW6432Node\Microsoft\EdgeUpdate\Clients\{F3017226-FE2A-4295-8BDF-00C3A9A7E4C5}" "pv"
  ${If} $0 == ""
    ReadRegStr $0 HKCU "Software\Microsoft\EdgeUpdate\Clients\{F3017226-FE2A-4295-8BDF-00C3A9A7E4C5}" "pv"
  ${EndIf}
  ${If} $0 == ""
    IfSilent webviewSilent
    MessageBox MB_YESNO|MB_ICONEXCLAMATION|MB_DEFBUTTON1 "NEST needs the Microsoft Edge WebView2 Runtime.$\r$\n$\r$\nIt is included with Windows 10 and 11 but appears to be missing here. Open the official download page now?" IDNO webviewDone
    ExecShell "open" "https://go.microsoft.com/fwlink/p/?LinkId=2124703"
    webviewDone:
    Abort
    webviewSilent:
    SetErrorLevel 70
    Abort
  ${EndIf}
FunctionEnd

; ---- Install ---------------------------------------------------------------
Section "!NEST File Organizer" SecCore
  SectionIn RO
  SetShellVarContext current

  ; Close a running copy so the EXE can be replaced on upgrade.
  nsExec::ExecToLog '"$SYSDIR\taskkill.exe" /IM ${APP_EXE} /F'
  Pop $0
  Sleep 300

  SetOutPath "$INSTDIR"
  File "/oname=${APP_EXE}" "${SOURCE_EXE}"
  File "/oname=README.md" "${SOURCE_README}"
  File "/oname=LICENSE.txt" "${SOURCE_LICENSE}"
  WriteUninstaller "$INSTDIR\uninstall.exe"

  ; Start Menu
  CreateDirectory "$SMPROGRAMS\NEST"
  CreateShortcut "$SMPROGRAMS\NEST\NEST.lnk" "$INSTDIR\${APP_EXE}" "" "$INSTDIR\${APP_EXE}" 0 SW_SHOWNORMAL "" "The file organizer that cleans up, then disappears."
  CreateShortcut "$SMPROGRAMS\NEST\Uninstall NEST.lnk" "$INSTDIR\uninstall.exe"

  ; The optional section recreates this if the user kept it selected.
  Delete "$DESKTOP\NEST.lnk"

  ; Upgrade information + Add/Remove Programs entry.
  WriteRegStr HKCU "${APP_REG_KEY}" "InstallDir" "$INSTDIR"
  WriteRegStr HKCU "${APP_REG_KEY}" "Version" "${APP_VERSION}"

  WriteRegStr HKCU "${UNINST_KEY}" "DisplayName" "${APP_NAME}"
  WriteRegStr HKCU "${UNINST_KEY}" "DisplayVersion" "${APP_VERSION}"
  WriteRegStr HKCU "${UNINST_KEY}" "Publisher" "${APP_PUBLISHER}"
  WriteRegStr HKCU "${UNINST_KEY}" "DisplayIcon" "$INSTDIR\${APP_EXE},0"
  WriteRegStr HKCU "${UNINST_KEY}" "InstallLocation" "$INSTDIR"
  WriteRegStr HKCU "${UNINST_KEY}" "UninstallString" '"$INSTDIR\uninstall.exe"'
  WriteRegStr HKCU "${UNINST_KEY}" "QuietUninstallString" '"$INSTDIR\uninstall.exe" /S'
  WriteRegStr HKCU "${UNINST_KEY}" "URLInfoAbout" "${APP_URL}"
  WriteRegStr HKCU "${UNINST_KEY}" "HelpLink" "${APP_URL}"
  WriteRegDWORD HKCU "${UNINST_KEY}" "NoModify" 1
  WriteRegDWORD HKCU "${UNINST_KEY}" "NoRepair" 1
  WriteRegDWORD HKCU "${UNINST_KEY}" "EstimatedSize" 12100

  System::Call 'shell32::SHChangeNotify(i 0x08000000, i 0, i 0, i 0)'
SectionEnd

Section "Desktop shortcut" SecDesktop
  SetShellVarContext current
  CreateShortcut "$DESKTOP\NEST.lnk" "$INSTDIR\${APP_EXE}" "" "$INSTDIR\${APP_EXE}" 0 SW_SHOWNORMAL "" "The file organizer that cleans up, then disappears."
SectionEnd

!insertmacro MUI_FUNCTION_DESCRIPTION_BEGIN
  !insertmacro MUI_DESCRIPTION_TEXT ${SecCore} "The NEST app, Start Menu shortcut and uninstaller."
  !insertmacro MUI_DESCRIPTION_TEXT ${SecDesktop} "Add a NEST shortcut to your desktop."
!insertmacro MUI_FUNCTION_DESCRIPTION_END

; ---- Uninstall -------------------------------------------------------------
Function un.onInit
  SetShellVarContext current
  StrCpy $DeleteUserData "0"
  IfSilent unInitDone
  MessageBox MB_YESNO|MB_ICONQUESTION|MB_DEFBUTTON2 "Remove your NEST settings and history as well?$\r$\n$\r$\nChoose No to keep them for a future reinstall." IDNO unInitDone
  StrCpy $DeleteUserData "1"
  unInitDone:
FunctionEnd

Section "Uninstall"
  SetShellVarContext current

  nsExec::ExecToLog '"$SYSDIR\taskkill.exe" /IM ${APP_EXE} /F'
  Pop $0
  Sleep 300

  Delete "$INSTDIR\${APP_EXE}"
  Delete "$INSTDIR\README.md"
  Delete "$INSTDIR\LICENSE.txt"
  Delete "$INSTDIR\uninstall.exe"
  RMDir "$INSTDIR"

  Delete "$SMPROGRAMS\NEST\NEST.lnk"
  Delete "$SMPROGRAMS\NEST\Uninstall NEST.lnk"
  RMDir "$SMPROGRAMS\NEST"
  Delete "$DESKTOP\NEST.lnk"

  DeleteRegKey HKCU "${UNINST_KEY}"
  DeleteRegKey HKCU "${APP_REG_KEY}"

  ${If} $DeleteUserData == "1"
    RMDir /r "$APPDATA\Nest"
    RMDir /r "$LOCALAPPDATA\Nest"
  ${EndIf}

  System::Call 'shell32::SHChangeNotify(i 0x08000000, i 0, i 0, i 0)'
SectionEnd
