Unicode true
ManifestDPIAware true
SetCompressor /SOLID lzma

!include "MUI2.nsh"
!define UNINSTALL_KEY "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APP_ID}"
!define OLD_SHORTCUT "FuckEnglish Preview.lnk"

Name "FuckEnglish"
OutFile "${APP_STAGE}\FuckEnglish Setup ${APP_VERSION} amd64.exe"
InstallDir "$LOCALAPPDATA\Programs\FuckEnglish"
InstallDirRegKey HKCU "${UNINSTALL_KEY}" "InstallLocation"
RequestExecutionLevel user
BrandingText " "
!define MUI_FINISHPAGE_RUN "$INSTDIR\FuckEnglish.exe"
!define MUI_FINISHPAGE_SHOWREADME
!define MUI_FINISHPAGE_SHOWREADME_TEXT "Create a desktop shortcut"
!define MUI_FINISHPAGE_SHOWREADME_FUNCTION CreateDesktopShortcut
!insertmacro MUI_PAGE_DIRECTORY
!insertmacro MUI_PAGE_INSTFILES
!insertmacro MUI_PAGE_FINISH
!insertmacro MUI_UNPAGE_CONFIRM
!insertmacro MUI_UNPAGE_INSTFILES
!insertmacro MUI_LANGUAGE "English"

Section
  SetOutPath "$INSTDIR"
  File "${APP_STAGE}\FuckEnglish.exe"
  File /r "${APP_STAGE}\BrimDeck-themes"
  File /r "${APP_STAGE}\licenses"
  File "${APP_STAGE}\icon.png"
  File "${APP_STAGE}\NOTICE-BrimDeck.txt"
  Delete "$INSTDIR\FuckEnglish Preview.exe"
  Delete "$SMPROGRAMS\${OLD_SHORTCUT}"
  Delete "$DESKTOP\${OLD_SHORTCUT}"
  WriteUninstaller "$INSTDIR\Uninstall.exe"
  CreateShortCut "$SMPROGRAMS\FuckEnglish.lnk" "$INSTDIR\FuckEnglish.exe"
  WriteRegStr HKCU "${UNINSTALL_KEY}" "DisplayName" "FuckEnglish"
  WriteRegStr HKCU "${UNINSTALL_KEY}" "DisplayVersion" "${APP_VERSION}"
  WriteRegStr HKCU "${UNINSTALL_KEY}" "DisplayIcon" "$INSTDIR\FuckEnglish.exe"
  WriteRegStr HKCU "${UNINSTALL_KEY}" "InstallLocation" "$INSTDIR"
  WriteRegStr HKCU "${UNINSTALL_KEY}" "UninstallString" '"$INSTDIR\Uninstall.exe"'
  WriteRegStr HKCU "${UNINSTALL_KEY}" "QuietUninstallString" '"$INSTDIR\Uninstall.exe" /S'
  WriteRegDWORD HKCU "${UNINSTALL_KEY}" "NoModify" 1
  WriteRegDWORD HKCU "${UNINSTALL_KEY}" "NoRepair" 1
SectionEnd

Function CreateDesktopShortcut
  SetOutPath "$INSTDIR"
  CreateShortCut "$DESKTOP\FuckEnglish.lnk" "$INSTDIR\FuckEnglish.exe"
FunctionEnd

Section "Uninstall"
  Delete "$SMPROGRAMS\FuckEnglish.lnk"
  Delete "$DESKTOP\FuckEnglish.lnk"
  DeleteRegKey HKCU "${UNINSTALL_KEY}"
  RMDir /r "$INSTDIR"
SectionEnd
